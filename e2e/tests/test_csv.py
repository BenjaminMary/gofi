def test_csv_export_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/csv/export")
    assert logged_in_page.locator("h1", has_text="Export CSV").is_visible()


def test_csv_export_requires_auth(page, base_url):
    page.goto(f"{base_url}/csv/export")
    assert page.locator("text=Déconnecté").is_visible()


def test_csv_import_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/csv/import")
    assert logged_in_page.locator("h1", has_text="Import CSV").is_visible()


def test_csv_import_requires_auth(page, base_url):
    page.goto(f"{base_url}/csv/import")
    assert page.locator("text=Déconnecté").is_visible()
