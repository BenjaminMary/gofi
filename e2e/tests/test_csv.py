import pathlib

# Tested:
# 1. import page loads: h1 "Import CSV"
# 2. import page requires auth
# 3. import sample CSV file: upload gofi1-UTF8-LF.csv, verify HTMX response
# 4. export page loads: h1 "Export CSV"
# 5. export page requires auth
# 6. export reset: opening the reset section and confirming clears exported flag
# 7. export download: native POST returns a .csv file with correct headers
# 8. export download again: second download returns "Rien à télécharger" (all already exported)

# Ordering note: "a_import_*" sorts before "export_*" (a < e)
# so import tests always run before export tests.
# "export_y_download" sorts after "export_reset" (y > r)
# so reset runs before download, ensuring all records are marked unexported.

SAMPLE_CSV = pathlib.Path(__file__).parent / "fixtures" / "gofi1-UTF8-LF.csv"


# 1.
def test_csv_a_import_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/csv/import")
    assert logged_in_page.locator("h1", has_text="Import CSV").is_visible()


# 2.
def test_csv_a_import_requires_auth(page, base_url):
    page.goto(f"{base_url}/csv/import")
    assert page.locator("text=Déconnecté").is_visible()


# 3.
def test_csv_a_import_sample(logged_in_page, base_url):
    # upload gofi1-UTF8-LF.csv via the file input — HTMX POSTs to /csv/import
    # response goes into #textarea; form is removed after request
    logged_in_page.goto(f"{base_url}/csv/import")
    logged_in_page.locator("input#csvFile").set_input_files(str(SAMPLE_CSV))
    logged_in_page.locator("section#form button[type='submit']").click()
    # wait for HTMX to swap the response into #textarea and remove the section
    logged_in_page.wait_for_selector("textarea#textarea:not(:empty)")
    assert logged_in_page.locator("section#form").count() == 0  # section removed on success


# 4.
def test_csv_export_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/csv/export")
    assert logged_in_page.locator("h1", has_text="Export CSV").is_visible()


# 5.
def test_csv_export_requires_auth(page, base_url):
    page.goto(f"{base_url}/csv/export")
    assert page.locator("text=Déconnecté").is_visible()


# 6.
def test_csv_export_reset(logged_in_page, base_url):
    # the reset section is inside a <details> — open it first, then submit
    # marks all records as unexported so the download test gets all data
    logged_in_page.goto(f"{base_url}/csv/export")
    logged_in_page.locator("section#reset details summary").click()
    logged_in_page.locator("form#formReset button").click()
    logged_in_page.wait_for_selector("text=Reset effectué")


# 7.
def test_csv_export_y_download(logged_in_page, base_url):
    # the export form does a native POST (not HTMX) — Playwright intercepts it as a download
    # runs after export_reset (y > r) so all records are unexported and the file has data
    logged_in_page.goto(f"{base_url}/csv/export")
    with logged_in_page.expect_download() as download_info:
        logged_in_page.locator("form#formDL button").click()
    download = download_info.value
    assert download.suggested_filename.endswith(".csv")
    # read and verify the header line — file is UTF-8 with BOM
    content = pathlib.Path(download.path()).read_text(encoding="utf-8-sig")
    header = content.splitlines()[0]
    for expected_col in ["𫝀é ꮖꭰ", "Date", "Mode", "Account", "Product", "PriceStr", "Category", "Checked", "Exported"]:
        assert expected_col in header, f"Missing column in CSV header: {expected_col}"


# 8.
def test_csv_export_z_download_empty(logged_in_page, base_url):
    # second download after the first — all records already exported, file should be header-only
    logged_in_page.goto(f"{base_url}/csv/export")
    with logged_in_page.expect_download() as download_info:
        logged_in_page.locator("form#formDL button").click()
    download = download_info.value
    assert download.suggested_filename.endswith(".csv")
    content = pathlib.Path(download.path()).read_text(encoding="utf-8-sig")
    assert "Rien à télécharger" in content
