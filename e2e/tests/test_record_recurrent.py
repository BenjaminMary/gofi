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
