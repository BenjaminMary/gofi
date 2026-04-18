# Tested:
# 1. page loads: h1 "Enregistrements réguliers", section#form visible
# 2. page requires auth
# 3. create recurrent record: row appears in the list after submit
# 4. create form hidden by default inside <details> (open attribute absent)
# 5. save row: clicking save button inserts a new row in #lastInsert
# 6. edit recurrent row: change designation via edit form, verify update in list
# 7. delete recurrent row: response "OK" in #infoMainForm after deleteRR click

# Alphabetical ordering is load-bearing:
#   create_success (c) → save_row (s) — save_row relies on "test recurrent" existing
#   create_success (c) → delete (d)   — delete picks "test recurrent" row


# 1.
def test_record_recurrent_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/record/recurrent")
    assert logged_in_page.locator("h1", has_text="Enregistrements réguliers").is_visible()
    assert logged_in_page.locator("section#form").is_visible()


# 2.
def test_record_recurrent_requires_auth(page, base_url):
    page.goto(f"{base_url}/record/recurrent")
    assert page.locator("text=Déconnecté").is_visible()


# 3.
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


# 4.
def test_record_recurrent_form_hidden_by_default(logged_in_page, base_url):
    # the create form is collapsed inside a <details> — summary is visible but form inputs are not
    logged_in_page.goto(f"{base_url}/record/recurrent")
    assert logged_in_page.locator("details#openForm > summary").is_visible()
    # details is closed by default: the open attribute is absent (None)
    assert logged_in_page.locator("details#openForm").get_attribute("open") is None


# 5.
def test_record_recurrent_save_row(logged_in_page, base_url):
    # "test recurrent" persists in DB from test_record_recurrent_create_success
    # server renders existing rows with init=true → save button (id^='s') is enabled
    # clicking save triggers HTMX formSaveRR → row appears in #lastInsert
    logged_in_page.goto(f"{base_url}/record/recurrent")
    logged_in_page.wait_for_selector("text=test recurrent")
    logged_in_page.locator("tr", has_text="test recurrent").first.locator("button[id^='s']").click()  # [id^='s'] = CSS "starts with": matches s123, s456, …
    logged_in_page.wait_for_selector("#lastInsert tr")


# 6.
def test_record_recurrent_edit(logged_in_page, base_url, created_account):
    # create a fresh recurrent, reload to enable row buttons (HTMX-inserted rows have disabled buttons),
    # click the row edit button, update the designation, submit via button#editRR
    logged_in_page.goto(f"{base_url}/record/recurrent")
    logged_in_page.locator("details#openForm > summary").click()
    logged_in_page.locator("select[name='recurrence']").select_option("mensuelle")
    logged_in_page.locator("select[name='compte']").select_option(created_account)
    logged_in_page.locator("input[type='radio'][name='categorie']").first.check()
    logged_in_page.locator("input[name='prix']").fill("30.00")
    logged_in_page.locator("input[value='expense']").check()
    logged_in_page.locator("input[name='designation']").fill("test recurrent to edit")
    logged_in_page.locator("button#createRR").click()
    logged_in_page.wait_for_selector("text=test recurrent to edit")
    # reload to server-render the row — makes save/edit buttons active
    logged_in_page.reload()
    logged_in_page.wait_for_selector("text=test recurrent to edit")
    # click the row edit button — JS populates the create form and reveals button#editRR
    logged_in_page.locator("tr", has_text="test recurrent to edit").first.locator("button[id^='e']").click()  # [id^='e'] = CSS "starts with"
    logged_in_page.wait_for_selector("button#editRR", state="visible")
    # overwrite the designation and submit — editRR POSTs to /record/recurrent/update, target #newRR afterbegin
    logged_in_page.locator("input[name='designation']").fill("test recurrent edited")
    logged_in_page.locator("button#editRR").click()
    logged_in_page.wait_for_selector("text=test recurrent edited")


# 7.
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
