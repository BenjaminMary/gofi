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


def test_stats_lender_borrower_state_change(logged_in_page, base_url):
    # "Tiers LB" was created in test_record_lend_or_borrow.py (runs before this file alphabetically)
    # clicking input[id^='active-'] triggers JS: removes element + submits form2 (full page POST)
    # after reload, Tiers LB should appear in #lbTableRowsInactive
    logged_in_page.goto(f"{base_url}/stats/lender-borrower/0")
    logged_in_page.wait_for_selector("#lbTableRows")
    assert logged_in_page.locator("#lbTableRows tr", has_text="Tiers LB").count() >= 1
    with logged_in_page.expect_navigation():
        logged_in_page.locator("#lbTableRows tr", has_text="Tiers LB").first.locator("input[id^='active-']").click()  # [id^='active-'] = CSS "starts with": matches active-1, active-2, …
    assert logged_in_page.locator("#lbTableRowsInactive tr", has_text="Tiers LB").count() >= 1


def test_stats_lender_borrower_state_reactivate(logged_in_page, base_url):
    # runs after test_stats_lender_borrower_state_change ("state_change" < "state_reactivate" alphabetically)
    # "Tiers LB" was deactivated in state_change — click input[id^='inactive-'] to reactivate it
    logged_in_page.goto(f"{base_url}/stats/lender-borrower/0")
    logged_in_page.wait_for_selector("#lbTableRowsInactive")
    assert logged_in_page.locator("#lbTableRowsInactive tr", has_text="Tiers LB").count() >= 1
    with logged_in_page.expect_navigation():
        logged_in_page.locator("#lbTableRowsInactive tr", has_text="Tiers LB").first.locator("input[id^='inactive-']").click()  # [id^='inactive-'] = CSS "starts with": matches inactive-1, …
    assert logged_in_page.locator("#lbTableRows tr", has_text="Tiers LB").count() >= 1
