def test_budget_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/budget")
    assert logged_in_page.locator("h1", has_text="Budgets").is_visible()


def test_budget_requires_auth(page, base_url):
    page.goto(f"{base_url}/budget")
    assert page.locator("text=Déconnecté").is_visible()
