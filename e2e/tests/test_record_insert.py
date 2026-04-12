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


def test_record_insert_success(logged_in_page, base_url, created_account):
    # fill and submit the form — verify the record appears in the HTMX recap response
    # (navigating fresh to the page shows an empty recap table, so we insert here directly)
    logged_in_page.goto(f"{base_url}/record/insert/")
    logged_in_page.locator("select[name='compte']").select_option(created_account)
    logged_in_page.locator("input[type='radio'][name='categorie']").first.check()
    logged_in_page.locator("input[name='prix']").fill("5.00")
    logged_in_page.locator("input[value='expense']").check()
    logged_in_page.locator("input[name='designation']").fill("test insert success")
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_selector("text=test insert success")


def test_record_insert_gain_direction(logged_in_page, base_url, created_account):
    # insert a gain — the HTMX recap should show the row with a positive amount
    logged_in_page.goto(f"{base_url}/record/insert/")
    logged_in_page.locator("select[name='compte']").select_option(created_account)
    logged_in_page.locator("input[type='radio'][name='categorie']").first.check()
    logged_in_page.locator("input[name='prix']").fill("42.00")
    logged_in_page.locator("input[value='gain']").check()
    logged_in_page.locator("input[name='designation']").fill("test insert gain")
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_selector("text=test insert gain")
    # gain amounts are stored positive — the recap row should show +42.00
    assert logged_in_page.locator("text=42.00").first.is_visible()


def test_record_insert_missing_amount_blocked(logged_in_page, base_url, created_account):
    # prix field is required (HTML5) — submitting without it blocks the request client-side
    # no HTMX request fires, so #lastInsert is unchanged
    logged_in_page.goto(f"{base_url}/record/insert/")
    logged_in_page.locator("select[name='compte']").select_option(created_account)
    logged_in_page.locator("input[type='radio'][name='categorie']").first.check()
    logged_in_page.locator("input[value='expense']").check()
    # snapshot existing rows from DB before clicking — server pre-renders them on page load
    rows_before = logged_in_page.locator("#lastInsert tr").count()
    first_row_text = logged_in_page.locator("#lastInsert tr").first.inner_text() if rows_before > 0 else None
    # do NOT fill prix — leave it empty
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_timeout(300)  # wait to confirm no HTMX response fires
    assert logged_in_page.locator("#lastInsert tr").count() == rows_before
    if first_row_text is not None:
        assert logged_in_page.locator("#lastInsert tr").first.inner_text() == first_row_text
