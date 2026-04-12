# Tested:
# 1. edit page loads: clicking a row edit link shows h1 "Modifier des données"
# 2. edit page requires auth (direct URL /record/edit/1)
# 3. edit success: update designation and submit → OK in #htmxInfo

from conftest import insert_record, open_advanced_mode_and_reload


# 1.
def test_record_edit_page_loads(logged_in_page, base_url, created_account):
    insert_record(logged_in_page, base_url, created_account, designation="test edit page loads")
    logged_in_page.goto(f"{base_url}/record/alter/edit")
    open_advanced_mode_and_reload(logged_in_page, created_account, checked="0")
    logged_in_page.locator("a[href*='/record/edit/']").first.click()  # [href*=] = CSS "contains": matches /record/edit/1, /record/edit/42, …
    assert logged_in_page.locator("h1", has_text="Modifier des données").is_visible()


# 2.
def test_record_edit_requires_auth(page, base_url):
    page.goto(f"{base_url}/record/edit/1")
    assert page.locator("text=Déconnecté").is_visible()


# 3.
def test_record_edit_success(logged_in_page, base_url, created_account):
    insert_record(logged_in_page, base_url, created_account, designation="test edit success")
    logged_in_page.goto(f"{base_url}/record/alter/edit")
    open_advanced_mode_and_reload(logged_in_page, created_account, checked="0")
    logged_in_page.locator("a[href*='/record/edit/']").first.click()  # [href*=] = CSS "contains": matches /record/edit/1, /record/edit/42, …
    assert logged_in_page.locator("h1", has_text="Modifier des données").is_visible()
    logged_in_page.locator("input[name='FT.designation']").fill("test edit success edited")
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_selector("#htmxInfo:has-text('OK')")
