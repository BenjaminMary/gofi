import datetime
from conftest import insert_record

# Tested:
# 1.  page loads: h1 "Insérer des données", form section, submit button visible
# 2.  page requires auth
# 3.  created account appears in the compte select options
# 4.  insert expense: row appears in HTMX recap after submit
# 5.  insert gain: row appears with correct amount in recap
# 6.  missing amount: HTML5 required blocks form — recap row count unchanged
# 7.  fields cleared after submit: prix and designation are empty after a successful insert
# 8.  category summary updates: clicking a different radio updates #summaryCategory text
# 9.  URL pre-fill: navigating with params pre-populates account, category, designation, amount, direction
# 10. prerecordlink button: clicking it navigates to a URL encoding the current form state
# 11. sequential inserts: second insert is prepended before the first (hx-swap="afterbegin")
# 12. minimum amount (0.01): accepted and record appears in recap
# 13. empty designation: optional field — submission succeeds without it
# 14. expense direction pre-selected: expense radio is checked on fresh page load
# 15. multiple inserts


# 1.
def test_record_insert_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/record/insert/")
    assert logged_in_page.locator("h1", has_text="Insérer des données").is_visible()
    assert logged_in_page.locator("section#form").is_visible()
    assert logged_in_page.locator("button#idSubmit1").is_visible()


# 2.
def test_record_insert_requires_auth(page, base_url):
    page.goto(f"{base_url}/record/insert/")
    assert page.locator("text=Déconnecté").is_visible()


# 3.
def test_record_insert_account_in_select(logged_in_page, base_url, created_account):
    # account created via fixture should appear in the compte select
    # option elements are never "visible" in Playwright — use count() instead
    logged_in_page.goto(f"{base_url}/record/insert/")
    assert logged_in_page.locator(f"select[name='compte'] option[value='{created_account}']").count() >= 1


# 4.
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


# 5.
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


# 6.
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


# 7.
def test_record_insert_fields_cleared_after_submit(logged_in_page, base_url, created_account):
    # hx-on::after-request clears prix and designation on a successful insert
    logged_in_page.goto(f"{base_url}/record/insert/")
    logged_in_page.locator("select[name='compte']").select_option(created_account)
    logged_in_page.locator("input[type='radio'][name='categorie']").first.check()
    logged_in_page.locator("input[name='prix']").fill("15.00")
    logged_in_page.locator("input[value='expense']").check()
    logged_in_page.locator("input[name='designation']").fill("fields clear test")
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_selector("text=fields clear test")
    # both fields must be empty after the HTMX handler runs
    assert logged_in_page.locator("input[name='prix']").input_value() == ""
    assert logged_in_page.locator("input[name='designation']").input_value() == ""


# 8.
def test_record_insert_category_summary_updates(logged_in_page, base_url):
    # clicking a radio inside #categoryDropdown updates #summaryCategory text via JS
    logged_in_page.goto(f"{base_url}/record/insert/")
    radios = logged_in_page.locator("input[type='radio'][name='categorie']").all()
    if len(radios) < 2:
        return  # only one category available — nothing to assert
    second_cat_name = radios[1].get_attribute("value")
    # open the dropdown so the radio is interactable, then click it
    logged_in_page.locator("#categoryDropdown summary").click()
    radios[1].click()
    assert logged_in_page.locator("#summaryCategory").inner_text() == second_cat_name


# 9.
def test_record_insert_url_prefill(logged_in_page, base_url, created_account):
    # navigating to /record/insert/{account}/{category}/{product}/{direction}/{price}
    # pre-populates all form fields — direction "+" maps to "gain"
    logged_in_page.goto(f"{base_url}/record/insert/")
    first_cat = logged_in_page.locator("input[type='radio'][name='categorie']").first.get_attribute("value")
    logged_in_page.goto(f"{base_url}/record/insert/{created_account}/{first_cat}/pre-fill-test/+/99.99")
    # account select pre-selected
    assert logged_in_page.locator("select[name='compte']").input_value() == created_account
    # category summary shows the category name
    assert logged_in_page.locator("#summaryCategory").inner_text() == first_cat
    # amount field pre-filled
    assert logged_in_page.locator("input[name='prix']").input_value() == "99.99"
    # designation pre-filled
    assert logged_in_page.locator("input[name='designation']").input_value() == "pre-fill-test"
    # direction "+" → gain radio checked
    assert logged_in_page.locator("input[value='gain']").is_checked()


# 10.
def test_record_insert_prerecordlink_navigates(logged_in_page, base_url, created_account):
    # clicking "Raccourci avec pré-saisie" calls window.location.href with encoded form state
    logged_in_page.goto(f"{base_url}/record/insert/")
    logged_in_page.locator("select[name='compte']").select_option(created_account)
    first_cat = logged_in_page.locator("input[type='radio'][name='categorie']").first.get_attribute("value")
    logged_in_page.locator("input[name='prix']").fill("7.77")
    logged_in_page.locator("input[value='expense']").check()
    logged_in_page.locator("input[name='designation']").fill("shortcut-test")
    with logged_in_page.expect_navigation():
        logged_in_page.locator("button#prerecordlink").click()
    url = logged_in_page.url
    assert created_account in url
    assert first_cat in url
    assert "shortcut-test" in url
    assert "7.77" in url


# 11.
def test_record_insert_sequential_inserts_prepended(logged_in_page, base_url, created_account):
    # hx-swap="afterbegin" means each new record is inserted at the top of #lastInsert
    # the most recently inserted record must appear before the previous one
    logged_in_page.goto(f"{base_url}/record/insert/")
    logged_in_page.locator("select[name='compte']").select_option(created_account)
    logged_in_page.locator("input[type='radio'][name='categorie']").first.check()
    logged_in_page.locator("input[value='expense']").check()

    # first insert
    logged_in_page.locator("input[name='prix']").fill("1.00")
    logged_in_page.locator("input[name='designation']").fill("seq-insert-first")
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_selector("text=seq-insert-first")

    # second insert — account/category/direction are preserved; only prix/designation were cleared
    logged_in_page.locator("input[name='prix']").fill("2.00")
    logged_in_page.locator("input[name='designation']").fill("seq-insert-second")
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_selector("text=seq-insert-second")

    # verify second insert row appears before the first in the DOM
    rows = logged_in_page.locator("#lastInsert tr").all_inner_texts()
    idx_second = next(i for i, t in enumerate(rows) if "seq-insert-second" in t)
    idx_first = next(i for i, t in enumerate(rows) if "seq-insert-first" in t)
    assert idx_second < idx_first, (
        f"Expected 'seq-insert-second' (row {idx_second}) before 'seq-insert-first' (row {idx_first})"
    )


# 12.
def test_record_insert_minimum_amount(logged_in_page, base_url, created_account):
    # the amount field allows a minimum of 0.01 (step="0.01" min="0.00")
    logged_in_page.goto(f"{base_url}/record/insert/")
    logged_in_page.locator("select[name='compte']").select_option(created_account)
    logged_in_page.locator("input[type='radio'][name='categorie']").first.check()
    logged_in_page.locator("input[name='prix']").fill("0.01")
    logged_in_page.locator("input[value='expense']").check()
    logged_in_page.locator("input[name='designation']").fill("min amount test")
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_selector("text=min amount test")


# 13.
def test_record_insert_empty_designation_allowed(logged_in_page, base_url, created_account):
    # designation has no 'required' attribute — omitting it must not block the submit
    # success is signalled by the hx-on::after-request handler clearing the prix field
    logged_in_page.goto(f"{base_url}/record/insert/")
    logged_in_page.locator("select[name='compte']").select_option(created_account)
    logged_in_page.locator("input[type='radio'][name='categorie']").first.check()
    logged_in_page.locator("input[name='prix']").fill("3.33")
    logged_in_page.locator("input[value='expense']").check()
    # leave designation empty — do not fill it
    logged_in_page.locator("button#idSubmit1").click()
    # wait for the JS handler to clear prix (fires only on a successful HTMX response)
    logged_in_page.wait_for_function("document.getElementById('prix').value === ''")


# 14.
def test_record_insert_expense_default_checked(logged_in_page, base_url):
    # expense is the default direction — it must be pre-selected on a fresh page load
    logged_in_page.goto(f"{base_url}/record/insert/")
    assert logged_in_page.locator("input[value='expense']").is_checked()
    assert not logged_in_page.locator("input[value='gain']").is_checked()


# 15.
def test_record_insert_row_fields(logged_in_page, base_url):
    # insert a record then verify each column of the rendered row individually
    # row structure (simple mode, no editMode/checkboxMode):
    #   td[0] date | td[1] account | td[2] category icon | td[3] price | td[4] designation
    insert_record(logged_in_page, base_url, 
        "CB", category="Envie", amount="55.55", 
        designation="CB Envie -55.55")

    logged_in_page.goto(f"{base_url}/record/insert/")
    row = logged_in_page.locator("#lastInsert tr", has_text="CB Envie -55.55").first
    cells = row.locator("td")

    # td[0]: date — verify today's day+year number appears in the cell
    today = datetime.date.today()
    assert str(today.day) in cells.nth(0).inner_text()
    assert str(today.year) in cells.nth(0).inner_text()

    # td[1]: account
    assert cells.nth(1).inner_text() == "CB"

    # td[2]: category icon — no visible text, only an icomoon span; verify the class is present
    assert "icomoon" in cells.nth(2).locator("span").get_attribute("class")
    assert "wantko-wine" in cells.nth(2).locator("span").get_attribute("class")

    # td[3]: price
    assert cells.nth(3).inner_text() == "-55.55"

    # td[4]: designation
    assert cells.nth(4).inner_text() == "CB Envie -55.55"

