# Tested:
# 1.  edit page loads: h1 "Editer des gains ou dépenses"
# 2.  edit page requires auth
# 3.  inserted record appears in the edit table
# 4.  validate page loads: h1 "Valider des gains ou dépenses"
# 5.  validate page requires auth
# 6.  cancel page loads: h1 "Annuler des gains ou dépenses"
# 7.  cancel page requires auth
# 8.  validate success: check a row checkbox and submit → OK in #htmxInfo
# 9.  toggle-all: clicking thead checkbox checks all row checkboxes
# 10. cancel success: validate a row then cancel it via the cancel page

from conftest import insert_record, open_advanced_mode_and_reload


# 1.
def test_record_alter_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/record/alter/edit")
    assert logged_in_page.locator("h1", has_text="Editer des gains ou dépenses").is_visible()


# 2.
def test_record_alter_requires_auth(page, base_url):
    page.goto(f"{base_url}/record/alter/edit")
    assert page.locator("text=Déconnecté").is_visible()


# 4.
def test_record_alter_validate_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/record/alter/validate")
    assert logged_in_page.locator("h1", has_text="Valider des gains ou dépenses").is_visible()


# 5.
def test_record_alter_validate_requires_auth(page, base_url):
    page.goto(f"{base_url}/record/alter/validate")
    assert page.locator("text=Déconnecté").is_visible()


# 6.
def test_record_alter_cancel_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/record/alter/cancel")
    assert logged_in_page.locator("h1", has_text="Annuler des gains ou dépenses").is_visible()


# 7.
def test_record_alter_cancel_requires_auth(page, base_url):
    page.goto(f"{base_url}/record/alter/cancel")
    assert page.locator("text=Déconnecté").is_visible()


# 3.
def test_record_alter_shows_inserted_record(logged_in_page, base_url, created_account):
    insert_record(logged_in_page, base_url, created_account, designation="test alter show")
    logged_in_page.goto(f"{base_url}/record/alter/edit")
    assert logged_in_page.locator("text=test alter show").first.is_visible()


# 8.
def test_record_validate_success(logged_in_page, base_url, created_account):
    insert_record(logged_in_page, base_url, created_account, designation="test validate")
    logged_in_page.goto(f"{base_url}/record/alter/validate")
    assert logged_in_page.locator("h1", has_text="Valider des gains ou dépenses").is_visible()
    assert logged_in_page.locator("text=test validate").first.is_visible()
    # must check a checkbox before submitting — HTMX confirm JS collects :checked inputs
    logged_in_page.locator("tr", has_text="test validate").first.locator("input[type='checkbox'][name='idCheckbox']").check()
    logged_in_page.locator("button#submitValid").first.click()
    logged_in_page.wait_for_selector("#htmxInfo:has-text('OK')")


# 9.
def test_record_alter_toggle_all_checkboxes(logged_in_page, base_url, created_account):
    # insert a fresh unchecked record so the validate page has at least one row
    insert_record(logged_in_page, base_url, created_account, designation="test toggle all")
    logged_in_page.goto(f"{base_url}/record/alter/validate")
    logged_in_page.wait_for_selector("input[type='checkbox'][name='idCheckbox']")
    # click the thead toggle — all row checkboxes should become checked
    logged_in_page.locator("input[name='toggle']").check()
    checkboxes = logged_in_page.locator("input[type='checkbox'][name='idCheckbox']")
    for i in range(checkboxes.count()):
        assert checkboxes.nth(i).is_checked()


# 10.
def test_record_cancel_success_on_previously_validated_row(logged_in_page, base_url, created_account):
    insert_record(logged_in_page, base_url, created_account, designation="test cancel")
    # validate it so it can be cancelled
    logged_in_page.goto(f"{base_url}/record/alter/validate")
    logged_in_page.locator("tr", has_text="test cancel").first.locator("input[type='checkbox'][name='idCheckbox']").check()
    logged_in_page.locator("button#submitValid").click()
    logged_in_page.wait_for_selector("#htmxInfo:has-text('OK')")
    # cancel page defaults to non-validated rows — open advanced mode and switch to validated (1=Oui)
    logged_in_page.goto(f"{base_url}/record/alter/cancel")
    assert logged_in_page.locator("h1", has_text="Annuler des gains ou dépenses").is_visible()
    open_advanced_mode_and_reload(logged_in_page, created_account, checked="1")
    logged_in_page.locator("tr", has_text="test cancel").first.locator("input[type='checkbox'][name='idCheckbox']").check()
    logged_in_page.locator("button#submitCancel").click()
    logged_in_page.wait_for_selector("#htmxInfo:has-text('OK')")
