# Tested:
# 1. stats page loads: h1 "Statistiques"
# 2. stats page requires auth
# 3. lender-borrower stats page loads: h1 "Stats Prêt / Emprunt"
# 4. lender-borrower stats requires auth
# 5. switch mode (validated only): clicking toggle navigates to URL with "true"
# 6. year filter: changing the year input submits the form and keeps h1 visible

# Note: deactivate/reactivate tier and tier details tests are in test_record_lend_or_borrow.py
# because they depend on tiers created there (Tiers LB, Tiers Prêt)


# 1.
def test_stats_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/stats/false-0-false-false")
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()


# 2.
def test_stats_requires_auth(page, base_url):
    page.goto(f"{base_url}/stats/false-0-false-false")
    assert page.locator("text=Déconnecté").is_visible()


# 3.
def test_stats_lender_borrower_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/stats/lender-borrower/0")
    assert logged_in_page.locator("h1", has_text="Stats Prêt / Emprunt").is_visible()


# 4.
def test_stats_lender_borrower_requires_auth(page, base_url):
    page.goto(f"{base_url}/stats/lender-borrower/0")
    assert page.locator("text=Déconnecté").is_visible()


# 5.
def test_stats_switch_mode_validated_only(logged_in_page, base_url):
    # clicking switchMode checkbox submits the form and navigates to validated-only stats URL
    logged_in_page.goto(f"{base_url}/stats/false-0-false-false")
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    with logged_in_page.expect_navigation():
        logged_in_page.locator("#switchMode").click()
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    # URL changes to reflect the new mode (true = validated only)
    assert "true" in logged_in_page.url


# 6.
def test_stats_year_filter(logged_in_page, base_url):
    # changing the year input fires a change event that submits the form
    # page reloads with the new year in the URL
    logged_in_page.goto(f"{base_url}/stats/false-0-false-false")
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    with logged_in_page.expect_navigation():
        logged_in_page.locator("input#annee").fill("2020")
        logged_in_page.locator("input#annee").dispatch_event("change")
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    assert "2020" in logged_in_page.url
