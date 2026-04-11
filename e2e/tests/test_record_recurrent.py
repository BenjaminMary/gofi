def test_record_recurrent_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/record/recurrent")
    assert logged_in_page.locator("h1", has_text="Enregistrements réguliers").is_visible()
    assert logged_in_page.locator("section#form").is_visible()


def test_record_recurrent_requires_auth(page, base_url):
    page.goto(f"{base_url}/record/recurrent")
    assert page.locator("text=Déconnecté").is_visible()


def test_record_recurrent_create_success(logged_in_page, base_url, created_account):
    # open the create form, fill it, submit — designation should appear in the recurrent list
    logged_in_page.goto(f"{base_url}/record/recurrent")
    logged_in_page.locator("details#openForm > summary").click()
    logged_in_page.locator("select[name='recurrence']").select_option("mensuelle")
    logged_in_page.locator("select[name='compte']").select_option(created_account)
    logged_in_page.locator("input[type='radio'][name='categorie']").first.check()
    logged_in_page.locator("input[name='prix']").fill("75.00")
    logged_in_page.locator("input[value='expense']").check()
    logged_in_page.locator("input[name='designation']").fill("test recurrent")
    logged_in_page.locator("button#createRR").click()
    logged_in_page.wait_for_timeout(500)
    assert logged_in_page.locator("text=test recurrent").is_visible()


def test_record_recurrent_form_hidden_by_default(logged_in_page, base_url):
    # the create form is collapsed inside a <details> — summary is visible but form inputs are not
    logged_in_page.goto(f"{base_url}/record/recurrent")
    assert logged_in_page.locator("details#openForm > summary").is_visible()
    # details is closed by default: the open attribute is absent (None)
    assert logged_in_page.locator("details#openForm").get_attribute("open") is None


def test_record_recurrent_save_row(logged_in_page, base_url):
    # "test recurrent" persists in DB from test_record_recurrent_create_success
    # server renders existing rows with init=true → save button (id^='s') is enabled
    # clicking save triggers HTMX formSaveRR → row appears in #lastInsert
    logged_in_page.goto(f"{base_url}/record/recurrent")
    logged_in_page.wait_for_selector("text=test recurrent")
    logged_in_page.locator("tr", has_text="test recurrent").first.locator("button[id^='s']").click()  # [id^='s'] = CSS "starts with": matches s123, s456, …
    logged_in_page.wait_for_selector("#lastInsert tr")


def test_record_recurrent_delete(logged_in_page, base_url):
    # click the edit button (id^='e') to load row data into the create form
    # the JS opens details#openForm and reveals editRR/deleteRR buttons
    # then click deleteRR — response "OK, ligne supprimée." appears in #infoMainForm
    logged_in_page.goto(f"{base_url}/record/recurrent")
    logged_in_page.wait_for_selector("text=test recurrent")
    logged_in_page.locator("tr", has_text="test recurrent").first.locator("button[id^='e']").click()  # [id^='e'] = CSS "starts with": matches e123, e456, …
    logged_in_page.wait_for_selector("button#deleteRR", state="visible")
    logged_in_page.locator("button#deleteRR").click()
    logged_in_page.wait_for_selector("#infoMainForm:has-text('OK')")
