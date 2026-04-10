def test_record_insert_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/record/insert/")
    assert logged_in_page.locator("h1", has_text="Insérer des données").is_visible()
    assert logged_in_page.locator("section#form").is_visible()
    assert logged_in_page.locator("button#idSubmit1").is_visible()


def test_record_insert_requires_auth(page, base_url):
    page.goto(f"{base_url}/record/insert/")
    assert page.locator("text=Déconnecté").is_visible()


def test_record_insert_account_in_select(logged_in_page, base_url, created_account):
    # account created via fixture should appear in the compte select
    # option elements are never "visible" in Playwright — use count() instead
    logged_in_page.goto(f"{base_url}/record/insert/")
    assert logged_in_page.locator(f"select[name='compte'] option[value='{created_account}']").count() >= 1


def test_record_insert_success(logged_in_page, base_url, created_record):
    # created_record fixture inserts a record — verify it appears in the recap table
    # use .first because "test playwright cancel" also contains the substring
    logged_in_page.goto(f"{base_url}/record/insert/")
    assert logged_in_page.locator("text=test playwright").first.is_visible()
