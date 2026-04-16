import datetime
from conftest import insert_record

# Tested:
# 1.  stats page loads: h1 "Statistiques"
# 2.  stats page requires auth
# 3.  lender-borrower stats page loads: h1 "Stats Prêt / Emprunt"
# 4.  lender-borrower stats requires auth
# 5.  switch mode (validated only): clicking toggle navigates to URL with "true"
# 6.  year filter: changing the year input submits the form and keeps h1 visible
# 7.  all main sections rendered: account-stats, category-stats, graph-accounts, graph-expenses
# 8.  mode text reflects state: "Toutes les données" for false, "Données validées" for true
# 9.  account appears in account-stats table after an insert
# 10. category appears in category-stats table after an insert
# 11. TOTAUX row present in both account and category tables
# 12. switchStatsYearMonth toggle: navigates to URL with updated year/month flag
# 13. switchStatsGainExpense toggle: navigates to URL with updated gain/expense flag
# 14. hide/show category buttons toggle each other's disabled state

# Note: deactivate/reactivate tier and tier details tests are in test_record_lend_or_borrow.py
# because they depend on tiers created there (Tiers LB, Tiers Prêt)

STATS_DEFAULT = "false-0-false-false"


# 1.
def test_stats_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/stats/{STATS_DEFAULT}")
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()


# 2.
def test_stats_requires_auth(page, base_url):
    page.goto(f"{base_url}/stats/{STATS_DEFAULT}")
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
    logged_in_page.goto(f"{base_url}/stats/{STATS_DEFAULT}")
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
    logged_in_page.goto(f"{base_url}/stats/{STATS_DEFAULT}")
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    with logged_in_page.expect_navigation():
        logged_in_page.locator("input#annee").fill("2020")
        logged_in_page.locator("input#annee").dispatch_event("change")
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    assert "2020" in logged_in_page.url


# 7.
def test_stats_main_sections_rendered(logged_in_page, base_url):
    # all four data sections must be present regardless of whether any records exist
    logged_in_page.goto(f"{base_url}/stats/{STATS_DEFAULT}")
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    assert logged_in_page.locator("section#account-stats").is_visible()
    assert logged_in_page.locator("section#category-stats").is_visible()
    assert logged_in_page.locator("section#graph-accounts").is_visible()
    assert logged_in_page.locator("section#graph-expenses").is_visible()


# 8.
def test_stats_mode_text_reflects_state(logged_in_page, base_url):
    # #mode span shows which data set is active based on the URL flag
    logged_in_page.goto(f"{base_url}/stats/false-0-false-false")
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    assert "Toutes les données" in logged_in_page.locator("#mode").inner_text()

    logged_in_page.goto(f"{base_url}/stats/true-0-false-false")
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    assert "Données validées" in logged_in_page.locator("#mode").inner_text()


# 9.
def test_stats_account_appears_after_insert(logged_in_page, base_url, created_account):
    # insert a record for this session's account then verify it appears in the account table
    insert_record(logged_in_page, base_url, created_account,
                  amount="100.00", designation="stats-account-test")
    year = datetime.date.today().year
    logged_in_page.goto(f"{base_url}/stats/false-{year}-false-false")
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    assert logged_in_page.locator("section#account-stats td", has_text=created_account).is_visible()


# 10.
def test_stats_category_appears_after_insert(logged_in_page, base_url, created_account):
    # insert a record with the first available category and verify it shows in category-stats
    logged_in_page.goto(f"{base_url}/record/insert/")
    cat_name = logged_in_page.locator("input[type='radio'][name='categorie']").first.get_attribute("value")
    insert_record(logged_in_page, base_url, created_account,
                  category=cat_name, amount="50.00", designation="stats-category-test")
    year = datetime.date.today().year
    logged_in_page.goto(f"{base_url}/stats/false-{year}-false-false")
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    assert logged_in_page.locator("section#category-stats td small", has_text=cat_name).is_visible()


# 11.
def test_stats_totaux_row_in_tables(logged_in_page, base_url):
    # both account and category tables must have a TOTAUX row in their tfoot
    logged_in_page.goto(f"{base_url}/stats/{STATS_DEFAULT}")
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    assert logged_in_page.locator("section#account-stats tfoot td", has_text="TOTAUX").is_visible()
    assert logged_in_page.locator("section#category-stats tfoot td", has_text="TOTAUX").is_visible()


# 12.
def test_stats_switch_year_month_toggle(logged_in_page, base_url):
    # clicking switchStatsYearMonth submits the form — the year/month flag changes in the URL
    logged_in_page.goto(f"{base_url}/stats/false-0-false-false")
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    with logged_in_page.expect_navigation():
        logged_in_page.locator("#switchStatsYearMonth").click()
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    # third segment of the slug switches from "false" to "true"; strip query string before checking
    assert logged_in_page.url.split("?")[0].endswith("-true-false")


# 13.
def test_stats_switch_gain_expense_toggle(logged_in_page, base_url):
    # clicking switchStatsGainExpense submits the form — the gain/expense flag changes in the URL
    logged_in_page.goto(f"{base_url}/stats/false-0-false-false")
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    with logged_in_page.expect_navigation():
        logged_in_page.locator("#switchStatsGainExpense").click()
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    # fourth segment of the slug switches from "false" to "true"; strip query string before checking
    assert logged_in_page.url.split("?")[0].endswith("-true")


# 14.
def test_stats_hide_show_categories_buttons(logged_in_page, base_url):
    # "Tout masquer" and "Tout afficher" buttons toggle each other's disabled state
    logged_in_page.goto(f"{base_url}/stats/{STATS_DEFAULT}")
    logged_in_page.locator("h1", has_text="Statistiques").wait_for()
    # initial state: hide enabled, show disabled
    assert not logged_in_page.locator("button#btn-hide-categories").is_disabled()
    assert logged_in_page.locator("button#btn-show-categories").is_disabled()
    # click hide → hide becomes disabled, show becomes enabled
    logged_in_page.locator("button#btn-hide-categories").click()
    assert logged_in_page.locator("button#btn-hide-categories").is_disabled()
    assert not logged_in_page.locator("button#btn-show-categories").is_disabled()
    # click show → back to initial state
    logged_in_page.locator("button#btn-show-categories").click()
    assert not logged_in_page.locator("button#btn-hide-categories").is_disabled()
    assert logged_in_page.locator("button#btn-show-categories").is_disabled()
