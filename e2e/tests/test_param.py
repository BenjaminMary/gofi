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
