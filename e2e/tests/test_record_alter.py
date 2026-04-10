def test_record_alter_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/record/alter/edit")
    assert logged_in_page.locator("h1", has_text="Editer des gains ou dépenses").is_visible()


def test_record_alter_requires_auth(page, base_url):
    page.goto(f"{base_url}/record/alter/edit")
    assert page.locator("text=Déconnecté").is_visible()


def test_record_alter_shows_inserted_record(logged_in_page, base_url, created_record):
    # the record created by fixture should appear in the alter/edit list
    logged_in_page.goto(f"{base_url}/record/alter/edit")
    assert logged_in_page.locator("text=test playwright").first.is_visible()


def test_record_validate_success(logged_in_page, base_url, created_record):
    logged_in_page.goto(f"{base_url}/record/alter/validate")
    assert logged_in_page.locator("h1", has_text="Valider des gains ou dépenses").is_visible()
    assert logged_in_page.locator("text=test playwright").first.is_visible()
    # must check a checkbox before submitting — HTMX confirm JS collects :checked inputs
    logged_in_page.locator("input[type='checkbox'][name='idCheckbox']").first.check()
    logged_in_page.locator("button#submitValid").first.click()
    logged_in_page.wait_for_selector("#htmxInfo:has-text('OK')")


def test_record_cancel_success_on_previously_validated_row(logged_in_page, base_url, created_record):
    # test_record_validate_success (same file, runs before this) validated "test playwright"
    # the alter page uses WhereCheckedStr="2" by default which needs to be switched to shows all records
    logged_in_page.goto(f"{base_url}/record/alter/cancel")
    assert logged_in_page.locator("h1", has_text="Annuler des gains ou dépenses").is_visible()
    # change mode to see previously validated rows
    logged_in_page.locator("#advancedMode").click()
    logged_in_page.locator("#checked").select_option(value="1")
    # select the account to refresh rows
    logged_in_page.locator("#compte").select_option(value="CB")
    logged_in_page.wait_for_timeout(500)
    assert logged_in_page.locator("text=test playwright").first.is_visible()
    logged_in_page.locator("tr", has_text="test playwright").first.locator("input[type='checkbox']").check()
    logged_in_page.locator("button#submitCancel").click()
    logged_in_page.wait_for_selector("#htmxInfo:has-text('OK')")
