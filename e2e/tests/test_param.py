def test_param_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/param")
    assert logged_in_page.locator("h1", has_text="Gérer les paramètres").is_visible()


def test_param_requires_auth(page, base_url):
    page.goto(f"{base_url}/param")
    assert page.locator("text=Déconnecté").is_visible()


def test_param_account_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/param/account")
    assert logged_in_page.locator("h1", has_text="Gérer les comptes").is_visible()


def test_param_account_requires_auth(page, base_url):
    page.goto(f"{base_url}/param/account")
    assert page.locator("text=Déconnecté").is_visible()


def test_param_category_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/param/category")
    assert logged_in_page.locator("h1", has_text="Gérer les catégories").is_visible()


def test_param_category_requires_auth(page, base_url):
    page.goto(f"{base_url}/param/category")
    assert page.locator("text=Déconnecté").is_visible()


def test_param_account_create(logged_in_page, base_url, created_account):
    # created_account fixture handles creation — verify account appears in list
    logged_in_page.goto(f"{base_url}/param/account")
    assert logged_in_page.locator(f"text={created_account}").count() >= 1


def test_param_account_create_too_short(logged_in_page, base_url):
    # JS blocks creation — error text appears in #infoText
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.locator("section#createAccSection summary").click()
    logged_in_page.locator("input#accountToCreate").fill("X")
    logged_in_page.locator("button#createAccount").click()
    logged_in_page.wait_for_selector("text=moins de 2 caractères")


def test_param_account_create_too_long(logged_in_page, base_url):
    # JS blocks creation when name exceeds 5 chars — error text appears in #infoText
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.locator("section#createAccSection summary").click()
    logged_in_page.locator("input#accountToCreate").fill("TOOLONG")
    logged_in_page.locator("button#createAccount").click()
    logged_in_page.wait_for_selector("text=plus de 5 caractères")


def test_param_account_create_forbidden_dash(logged_in_page, base_url):
    # JS blocks creation when name contains a dash — error text appears in #infoText
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.locator("section#createAccSection summary").click()
    logged_in_page.locator("input#accountToCreate").fill("A-B")
    logged_in_page.locator("button#createAccount").click()
    logged_in_page.wait_for_selector("text=caractères -")


def test_param_account_create_forbidden_space(logged_in_page, base_url):
    # JS blocks creation when name contains a space — error text appears in #infoText
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.locator("section#createAccSection summary").click()
    logged_in_page.locator("input#accountToCreate").fill("A B")
    logged_in_page.locator("button#createAccount").click()
    logged_in_page.wait_for_selector("text=caractères espace")


def test_param_account_create_duplicate(logged_in_page, base_url, created_account):
    # JS blocks creation when name already exists — error text appears in #infoText
    logged_in_page.goto(f"{base_url}/param/account")
    logged_in_page.locator("section#createAccSection summary").click()
    logged_in_page.locator("input#accountToCreate").fill(created_account)
    logged_in_page.locator("button#createAccount").click()
    logged_in_page.wait_for_selector("text=déjà existant")


def test_param_category_rendering_names(logged_in_page, base_url):
    # switch category rendering to "names" and save — no error should appear
    logged_in_page.goto(f"{base_url}/param")
    logged_in_page.locator("input#names").check()
    logged_in_page.locator("button#idSubmit3").click()
    logged_in_page.wait_for_timeout(500)
    # idSubmit3 is removed on success (hx-on::after-request in template)
    assert logged_in_page.locator("button#idSubmit3").count() == 0


def test_param_category_rendering_icons(logged_in_page, base_url):
    # switch category rendering back to "icons" and save
    logged_in_page.goto(f"{base_url}/param")
    logged_in_page.locator("input#icons").check()
    logged_in_page.locator("button#idSubmit3").click()
    logged_in_page.wait_for_timeout(500)
    assert logged_in_page.locator("button#idSubmit3").count() == 0


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
