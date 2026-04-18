# Tested:
# 1. edit page loads: clicking a row edit link shows h1 "Modifier des données"
# 2. edit page requires auth (direct URL /record/edit/1)
# 3. edit success: update designation and submit → OK in #htmxInfo
# 4. form pre-populated: editRow JS fills account, category summary, designation from stored record
# 5. unknown ID: /record/edit/0 → "Identifiant inconnu." (no panic)
# 6. edit designation persists: change is visible in the alter/edit table after submit
# 7. edit amount persists: updated amount visible in the alter/edit table row after submit
# 8. edit category: open dropdown, pick different category, summary updates, submit → OK

from conftest import insert_record, open_advanced_mode_and_reload


def _open_edit_form(page, base_url, account, designation):
    """Navigate to the edit form for the row matching designation in the alter/edit table."""
    page.goto(f"{base_url}/record/alter/edit")
    # limit=500: the table defaults to 8 rows; in the full suite many records exist
    # and the target row may not be in the first 8 without a higher limit
    open_advanced_mode_and_reload(page, account, checked="0", limit=500)
    page.locator("tr", has_text=designation).first.locator("a[href*='/record/edit/']").click()
    page.wait_for_selector("h1:has-text('Modifier des données')")


# 1.
def test_record_edit_page_loads(logged_in_page, base_url, created_account):
    insert_record(logged_in_page, base_url, created_account, designation="test edit page loads")
    _open_edit_form(logged_in_page, base_url, created_account, "test edit page loads")
    assert logged_in_page.locator("h1", has_text="Modifier des données").is_visible()


# 2.
def test_record_edit_requires_auth(page, base_url):
    page.goto(f"{base_url}/record/edit/1")
    assert page.locator("text=Déconnecté").is_visible()


# 3.
def test_record_edit_success(logged_in_page, base_url, created_account):
    insert_record(logged_in_page, base_url, created_account, designation="test edit success")
    _open_edit_form(logged_in_page, base_url, created_account, "test edit success")
    logged_in_page.locator("input[name='FT.designation']").fill("test edit success edited")
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_selector("#htmxInfo:has-text('OK')")


# 4.
def test_record_edit_form_prepopulated(logged_in_page, base_url, created_account):
    # editRow JS runs on page load and pre-fills the form from the stored record values
    logged_in_page.goto(f"{base_url}/record/insert/")
    cat_name = logged_in_page.locator("input[type='radio'][name='categorie']").first.get_attribute("value")
    insert_record(logged_in_page, base_url, created_account,
                  category=cat_name, amount="77.77", designation="prefill-edit-test")
    _open_edit_form(logged_in_page, base_url, created_account, "prefill-edit-test")
    # account select pre-filled
    assert logged_in_page.locator("select[name='FT.compte']").input_value() == created_account
    # category summary set by editRow JS
    assert logged_in_page.locator("#summaryCategory").inner_text() == cat_name
    # designation pre-filled
    assert logged_in_page.locator("input[name='FT.designation']").input_value() == "prefill-edit-test"
    # amount contains the inserted value (sign may differ for expense — check substring)
    assert "77.77" in logged_in_page.locator("input[name='FT.prix']").input_value()


# 5.
def test_record_edit_unknown_id(logged_in_page, base_url):
    # /record/edit/0 returns an empty list → handler renders "Identifiant inconnu." without panicking
    logged_in_page.goto(f"{base_url}/record/edit/0")
    assert logged_in_page.locator("p", has_text="Identifiant inconnu.").is_visible()


# 6.
def test_record_edit_designation_persists(logged_in_page, base_url, created_account):
    # change the designation and verify the update is reflected in the alter/edit table
    insert_record(logged_in_page, base_url, created_account, designation="edit-desig-before")
    _open_edit_form(logged_in_page, base_url, created_account, "edit-desig-before")
    logged_in_page.locator("input[name='FT.designation']").fill("edit-desig-after")
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_selector("#htmxInfo:has-text('OK')")
    # reload the alter table and confirm the new designation appears; old one is gone
    logged_in_page.goto(f"{base_url}/record/alter/edit")
    open_advanced_mode_and_reload(logged_in_page, created_account, checked="0", limit=500)
    assert logged_in_page.locator("text=edit-desig-after").first.is_visible()
    assert logged_in_page.locator("text=edit-desig-before").count() == 0


# 7.
def test_record_edit_amount_persists(logged_in_page, base_url, created_account):
    # change the amount and verify the updated value appears in the alter/edit table row
    insert_record(logged_in_page, base_url, created_account,
                  amount="11.11", designation="edit-amount-test")
    _open_edit_form(logged_in_page, base_url, created_account, "edit-amount-test")
    logged_in_page.locator("input[name='FT.prix']").fill("22.22")
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_selector("#htmxInfo:has-text('OK')")
    logged_in_page.goto(f"{base_url}/record/alter/edit")
    open_advanced_mode_and_reload(logged_in_page, created_account, checked="0", limit=500)
    row = logged_in_page.locator("tr", has_text="edit-amount-test").first
    assert "22.22" in row.inner_text()


# 8.
def test_record_edit_category_change(logged_in_page, base_url, created_account):
    # open the category dropdown, pick a different category, verify summary updates, submit → OK
    logged_in_page.goto(f"{base_url}/record/insert/")
    radios = logged_in_page.locator("input[type='radio'][name='categorie']").all()
    if len(radios) < 2:
        return  # only one category — nothing to switch to
    first_cat = radios[0].get_attribute("value")
    second_cat = radios[1].get_attribute("value")

    insert_record(logged_in_page, base_url, created_account,
                  category=first_cat, designation="edit-cat-test")
    _open_edit_form(logged_in_page, base_url, created_account, "edit-cat-test")

    # open the dropdown and click the second category (radio name is FT.categorie on the edit page)
    logged_in_page.locator("#categoryDropdown summary").click()
    logged_in_page.locator(f"input[type='radio'][name='FT.categorie'][value='{second_cat}']").click()
    logged_in_page.wait_for_selector("#categoryDropdown:not([open])")
    assert logged_in_page.locator("#summaryCategory").inner_text() == second_cat

    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_selector("#htmxInfo:has-text('OK')")
