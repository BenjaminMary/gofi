import pytest
from conftest import create_account, insert_record


@pytest.fixture(scope="module")
def deactivation_account(browser, base_url, auth_state):
    # created once for this module — PDC is only needed for deactivate/reactivate tests
    # keeps created_account (PCB) always active so record tests are unaffected
    return create_account(browser, base_url, auth_state, "PDC")

# Tested:
# 1.  /param page loads: h1 "Gérer les paramètres"
# 2.  /param requires auth
# 3.  /param/account page loads: h1 "Gérer les comptes"
# 4.  /param/account requires auth
# 5.  /param/category page loads: h1 "Gérer les catégories"
# 6.  /param/category requires auth
# 7.  account creation: created account appears in the list
# 8.  name too short (< 2 chars): JS error in #infoText
# 9.  name too long (> 5 chars): JS error in #infoText
# 10. name with dash: JS error in #infoText
# 11. name with space: JS error in #infoText
# 12. duplicate name: JS error in #infoText
# 13. category rendering set to "names": button#idSubmit3 disappears on success
# 14. category rendering set to "icons": button#idSubmit3 disappears on success
# 15. category edit: opens form, submit redirects back to category page
# 16. account deactivate: toggle switch removes PDC from active list; PDC absent from insert select (no records)
# 17. account reactivate: recreating PDC restores it (d < r alphabetically)
# 18. account reorder: moving second account up changes order in record insert select; order restored after
# 19. account deactivate with records: deactivation succeeds, PDC in unhandled section, absent from insert select (z runs after r)


# 1.
def test_param_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/param")
    assert logged_in_page.locator("h1", has_text="Gérer les paramètres").is_visible()


# 2.
def test_param_requires_auth(page, base_url):
    page.goto(f"{base_url}/param")
    assert page.locator("text=Déconnecté").is_visible()


# 3.
def test_param_account_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/param/account")
    assert logged_in_page.locator("h1", has_text="Gérer les comptes").is_visible()


# 4.
def test_param_account_requires_auth(page, base_url):
    page.goto(f"{base_url}/param/account")
    assert page.locator("text=Déconnecté").is_visible()


# 5.
def test_param_category_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/param/category")
    assert logged_in_page.locator("h1", has_text="Gérer les catégories").is_visible()


# 6.
def test_param_category_requires_auth(page, base_url):
    page.goto(f"{base_url}/param/category")
    assert page.locator("text=Déconnecté").is_visible()


# 7.
def test_param_account_create(logged_in_page, base_url, created_account):
    # created_account fixture handles creation — verify account appears in list
    logged_in_page.goto(f"{base_url}/param/account")
    assert logged_in_page.locator(f"text={created_account}").count() >= 1


# 8.
def test_param_account_create_too_short(logged_in_page, base_url):
    # JS blocks creation — error text appears in #infoText
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.locator("section#createAccSection summary").click()
    logged_in_page.locator("input#accountToCreate").fill("X")
    logged_in_page.locator("button#createAccount").click()
    logged_in_page.wait_for_selector("text=moins de 2 caractères")


# 9.
def test_param_account_create_too_long(logged_in_page, base_url):
    # JS blocks creation when name exceeds 5 chars — error text appears in #infoText
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.locator("section#createAccSection summary").click()
    logged_in_page.locator("input#accountToCreate").fill("TOOLONG")
    logged_in_page.locator("button#createAccount").click()
    logged_in_page.wait_for_selector("text=plus de 5 caractères")


# 10.
def test_param_account_create_forbidden_dash(logged_in_page, base_url):
    # JS blocks creation when name contains a dash — error text appears in #infoText
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.locator("section#createAccSection summary").click()
    logged_in_page.locator("input#accountToCreate").fill("A-B")
    logged_in_page.locator("button#createAccount").click()
    logged_in_page.wait_for_selector("text=caractères -")


# 11.
def test_param_account_create_forbidden_space(logged_in_page, base_url):
    # JS blocks creation when name contains a space — error text appears in #infoText
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.locator("section#createAccSection summary").click()
    logged_in_page.locator("input#accountToCreate").fill("A B")
    logged_in_page.locator("button#createAccount").click()
    logged_in_page.wait_for_selector("text=caractères espace")


# 12.
def test_param_account_create_duplicate(logged_in_page, base_url, created_account):
    # JS blocks creation when name already exists — error text appears in #infoText
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.locator("section#createAccSection summary").click()
    logged_in_page.locator("input#accountToCreate").fill(created_account)
    logged_in_page.locator("button#createAccount").click()
    logged_in_page.wait_for_selector("text=déjà existant")


# 13.
def test_param_category_rendering_names(logged_in_page, base_url):
    # switch category rendering to "names" and save — no error should appear
    logged_in_page.goto(f"{base_url}/param")
    logged_in_page.locator("input#names").check()
    logged_in_page.locator("button#idSubmit3").click()
    logged_in_page.wait_for_timeout(500)
    # idSubmit3 is removed on success (hx-on::after-request in template)
    assert logged_in_page.locator("button#idSubmit3").count() == 0


# 14.
def test_param_category_rendering_icons(logged_in_page, base_url):
    # switch category rendering back to "icons" and save
    logged_in_page.goto(f"{base_url}/param")
    logged_in_page.locator("input#icons").check()
    logged_in_page.locator("button#idSubmit3").click()
    logged_in_page.wait_for_timeout(500)
    assert logged_in_page.locator("button#idSubmit3").count() == 0


# 16.
def test_param_account_deactivate(logged_in_page, base_url, deactivation_account):
    # click the active toggle for PDC — JS posts the list without PDC then reloads the page
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.wait_for_selector(f"input#desactivate-{deactivation_account}")
    logged_in_page.locator(f"input#desactivate-{deactivation_account}").click()
    logged_in_page.wait_for_load_state("networkidle")
    assert logged_in_page.locator(f"input#desactivate-{deactivation_account}").count() == 0
    # deactivated account (no records) must not appear in the record insert select
    logged_in_page.goto(f"{base_url}/record/insert/")
    insert_options = logged_in_page.locator("select[name='compte'] option").all_text_contents()
    assert deactivation_account not in insert_options


# 17.
def test_param_account_reactivate(logged_in_page, base_url, deactivation_account):
    # runs after deactivate (d < r) — PDC is inactive; recreating it adds it back to the active list
    # no duplicate JS error since PDC is not currently in accountArray (it was deactivated)
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.locator("section#createAccSection summary").click()
    logged_in_page.locator("input#accountToCreate").fill(deactivation_account)
    logged_in_page.locator("button#createAccount").click()
    logged_in_page.wait_for_load_state("networkidle")
    assert logged_in_page.locator(f"input#desactivate-{deactivation_account}").is_visible()


# 18.
def test_param_account_reorder(logged_in_page, base_url):
    # read current account order from the insert form's select (no blank first option — forceSelect=false)
    logged_in_page.goto(f"{base_url}/record/insert/")
    options = logged_in_page.locator("select[name='compte'] option").all_text_contents()
    first_account = options[0]
    second_account = options[1]

    # click the up-arrow button for the second account — id: u-{thisAccount}-{prevAccount}
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.locator(f"button[id='u-{second_account}-{first_account}']").click()
    logged_in_page.wait_for_load_state("networkidle")

    # verify the order is reflected in the insert form's select
    logged_in_page.goto(f"{base_url}/record/insert/")
    new_options = logged_in_page.locator("select[name='compte'] option").all_text_contents()
    assert new_options[0] == second_account
    assert new_options[1] == first_account

    # restore original order so subsequent tests are unaffected
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.locator(f"button[id='u-{first_account}-{second_account}']").click()
    logged_in_page.wait_for_load_state("networkidle")


# 19.
def test_param_account_z_deactivate_with_records(logged_in_page, base_url, deactivation_account):
    # deactivating an account that has associated records is allowed — no guard in the UI
    # account disappears from active list and appears in "Comptes utilisés mais désactivés"
    # reactivate at the end so PDC is clean for any future use
    insert_record(logged_in_page, base_url, deactivation_account)
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.wait_for_selector(f"input#desactivate-{deactivation_account}")
    logged_in_page.locator(f"input#desactivate-{deactivation_account}").click()
    logged_in_page.wait_for_load_state("networkidle")
    assert logged_in_page.locator(f"input#desactivate-{deactivation_account}").count() == 0
    assert logged_in_page.locator("h5", has_text="Comptes utilisés mais désactivés").is_visible()
    # deactivated account with records must also be absent from the insert select
    # (AccountListUnhandled is for display only — it never feeds the insert dropdown)
    logged_in_page.goto(f"{base_url}/record/insert/")
    insert_options = logged_in_page.locator("select[name='compte'] option").all_text_contents()
    assert deactivation_account not in insert_options
    logged_in_page.goto(f"{base_url}/param/account")
    # reactivate PDC
    logged_in_page.locator("section#createAccSection summary").click()
    logged_in_page.locator("input#accountToCreate").fill(deactivation_account)
    logged_in_page.locator("button#createAccount").click()
    logged_in_page.wait_for_load_state("networkidle")
    assert logged_in_page.locator(f"input#desactivate-{deactivation_account}").is_visible()


# 15.
def test_param_category_edit(logged_in_page, base_url):
    # click edit on first active category (button id^='e-') — JS opens section#openForm
    # and populates the form fields; submit without changes triggers location.reload()
    logged_in_page.goto(f"{base_url}/param/category")
    logged_in_page.locator("button[id^='e-']").first.click()  # [id^='e-'] = CSS "starts with": matches e-0-1, e-1-3, …
    # JS sets section#openForm.hidden = false — button#editRR becomes visible
    logged_in_page.wait_for_selector("button#editRR", state="visible")
    with logged_in_page.expect_navigation():
        logged_in_page.locator("button#editRR").click()
    assert logged_in_page.locator("h1", has_text="Gérer les catégories").is_visible()
