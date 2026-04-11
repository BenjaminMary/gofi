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


def test_csv_export_download(logged_in_page, base_url):
    # the export form does a native POST (not HTMX) — Playwright intercepts it as a download
    logged_in_page.goto(f"{base_url}/csv/export")
    with logged_in_page.expect_download() as download_info:
        logged_in_page.locator("form#formDL button").click()
    download = download_info.value
    assert download.suggested_filename.endswith(".csv")


def test_csv_export_reset(logged_in_page, base_url):
    # the reset section is inside a <details> — open it first, then submit
    logged_in_page.goto(f"{base_url}/csv/export")
    logged_in_page.locator("section#reset details summary").click()
    logged_in_page.locator("form#formReset button").click()
    logged_in_page.wait_for_selector("text=Reset effectué")
