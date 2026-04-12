def test_budget_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/budget")
    assert logged_in_page.locator("h1", has_text="Budgets").is_visible()


def test_budget_requires_auth(page, base_url):
    page.goto(f"{base_url}/budget")
    assert page.locator("text=Déconnecté").is_visible()


def test_budget_categories_section_visible(logged_in_page, base_url):
    # section#budgets and its h3 heading are always rendered, regardless of whether budgets are set
    logged_in_page.goto(f"{base_url}/budget")
    assert logged_in_page.locator("section#budgets").is_visible()
    assert logged_in_page.locator("section#budgets h3", has_text="Catégories").is_visible()


def test_budget_color_code_toggle(logged_in_page, base_url):
    # section#fonctionnement contains a <details> with a color code example — closed by default
    logged_in_page.goto(f"{base_url}/budget")
    assert logged_in_page.locator("section#fonctionnement details").get_attribute("open") is None
    logged_in_page.locator("section#fonctionnement details summary").click()
    assert logged_in_page.locator("section#fonctionnement details").get_attribute("open") is not None
