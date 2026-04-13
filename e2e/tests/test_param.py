from conftest import insert_record

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
# 16. account deactivate: toggle switch removes account from active list
# 17. account reactivate: recreating a deactivated account restores it (d < r alphabetically)
# 18. account deactivate with records: deactivation succeeds, account appears in unhandled section (z runs after r)


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
def test_param_account_deactivate(logged_in_page, base_url, created_account):
    # click the active toggle for CB — JS posts the list without CB then reloads the page
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.wait_for_selector(f"input#desactivate-{created_account}")
    logged_in_page.locator(f"input#desactivate-{created_account}").click()
    logged_in_page.wait_for_load_state("networkidle")
    assert logged_in_page.locator(f"input#desactivate-{created_account}").count() == 0


# 17.
def test_param_account_reactivate(logged_in_page, base_url, created_account):
    # runs after deactivate (d < r) — CB is inactive; recreating it adds it back to the active list
    # no duplicate JS error since CB is not currently in accountArray (it was deactivated)
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.locator("section#createAccSection summary").click()
    logged_in_page.locator("input#accountToCreate").fill(created_account)
    logged_in_page.locator("button#createAccount").click()
    logged_in_page.wait_for_load_state("networkidle")
    assert logged_in_page.locator(f"input#desactivate-{created_account}").is_visible()


# 18.
def test_param_account_z_deactivate_with_records(logged_in_page, base_url, created_account):
    # deactivating an account that has associated records is allowed — no guard in the UI
    # the account disappears from the active list and appears in "Comptes utilisés mais désactivés"
    # reactivate at the end so subsequent record tests (test_record_*.py) can still use CB
    insert_record(logged_in_page, base_url, created_account)
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.wait_for_selector(f"input#desactivate-{created_account}")
    logged_in_page.locator(f"input#desactivate-{created_account}").click()
    logged_in_page.wait_for_load_state("networkidle")
    assert logged_in_page.locator(f"input#desactivate-{created_account}").count() == 0
    assert logged_in_page.locator("h5", has_text="Comptes utilisés mais désactivés").is_visible()
    # reactivate CB
    logged_in_page.locator("section#createAccSection summary").click()
    logged_in_page.locator("input#accountToCreate").fill(created_account)
    logged_in_page.locator("button#createAccount").click()
    logged_in_page.wait_for_load_state("networkidle")
    assert logged_in_page.locator(f"input#desactivate-{created_account}").is_visible()


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
