def test_stats_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/stats/false-0-false-false")
    assert logged_in_page.locator("h1", has_text="Statistiques").is_visible()


def test_stats_requires_auth(page, base_url):
    page.goto(f"{base_url}/stats/false-0-false-false")
    assert page.locator("text=Déconnecté").is_visible()


def test_stats_lender_borrower_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/stats/lender-borrower/0")
    assert logged_in_page.locator("h1", has_text="Stats Prêt / Emprunt").is_visible()


def test_stats_lender_borrower_requires_auth(page, base_url):
    page.goto(f"{base_url}/stats/lender-borrower/0")
    assert page.locator("text=Déconnecté").is_visible()
