import datetime
from conftest import edit_category, insert_record

# Tested:
# 1. page loads: h1 "Budgets" visible
# 2. page requires auth
# 3. section#budgets and h3 "Catégories" always rendered
# 4. color code <details> starts closed and opens on summary click
# 5. budget set (reset/mensuelle): category appears with meter; inserting a record updates spent amount; cleared after
# 6. budget with spending: same as 5 but focuses on the meter value after a second record


# 1.
def test_budget_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/budget")
    assert logged_in_page.locator("h1", has_text="Budgets").is_visible()


# 2.
def test_budget_requires_auth(page, base_url):
    page.goto(f"{base_url}/budget")
    assert page.locator("text=Déconnecté").is_visible()


# 3.
def test_budget_categories_section_visible(logged_in_page, base_url):
    # section#budgets and its h3 heading are always rendered, regardless of whether budgets are set
    logged_in_page.goto(f"{base_url}/budget")
    assert logged_in_page.locator("section#budgets").is_visible()
    assert logged_in_page.locator("section#budgets h3", has_text="Catégories").is_visible()


# 4.
def test_budget_color_code_toggle(logged_in_page, base_url):
    # the color code example section is inside a <details> — closed by default
    logged_in_page.goto(f"{base_url}/budget")
    assert logged_in_page.locator("section#fonctionnement details").get_attribute("open") is None
    logged_in_page.locator("section#fonctionnement details summary").click()
    assert logged_in_page.locator("section#fonctionnement details").get_attribute("open") is not None


# 5.
def test_budget_set_and_clear(logged_in_page, base_url, created_account):
    # set a monthly reset budget of 100 on the first active category
    # /budget uses filter "budget" (budgetPrice <> 0) — category must appear only when price > 0
    logged_in_page.goto(f"{base_url}/param/category")
    cat_name = logged_in_page.locator("#tableActiveCat td small").first.text_content().strip()
    budget_price = 100
    period_start = datetime.date.today().replace(day=1).isoformat()

    edit_category(logged_in_page, base_url, cat_name,
                  budget_type="reset", budget_period="mensuelle", budget_price=budget_price,
                  budget_start_date=period_start)

    # category must appear in /budget with the correct period/type label and a meter
    logged_in_page.goto(f"{base_url}/budget")
    assert logged_in_page.locator("section#budgets p", has_text=cat_name).is_visible()
    assert logged_in_page.locator("section#budgets p", has_text="mensuelle-reset").is_visible()
    # type="reset" renders 2 meters (previous + current period) — use count to avoid strict mode
    assert logged_in_page.locator(f"section#budgets meter[max='{budget_price}']").count() >= 1

    # insert a record with this category — current-period meter (last) must show > 0
    insert_record(logged_in_page, base_url, created_account,
                  category=cat_name, amount="10.00", designation="budget set test record")
    logged_in_page.goto(f"{base_url}/budget")
    meter = logged_in_page.locator(f"section#budgets meter[max='{budget_price}']").last
    assert float(meter.get_attribute("value")) > 0

    # clear the budget — category must disappear from /budget
    edit_category(logged_in_page, base_url, cat_name,
                  budget_type="-", budget_period="-", budget_price=0)
    logged_in_page.goto(f"{base_url}/budget")
    assert logged_in_page.locator("section#budgets p", has_text=cat_name).count() == 0


# 6.
def test_budget_spending_updates_amount(logged_in_page, base_url, created_account):
    # same setup, focuses on a larger spend and verifies the exact meter movement
    logged_in_page.goto(f"{base_url}/param/category")
    cat_name = logged_in_page.locator("#tableActiveCat td small").first.text_content().strip()
    budget_price = 500
    period_start = datetime.date.today().replace(day=1).isoformat()

    edit_category(logged_in_page, base_url, cat_name,
                  budget_type="reset", budget_period="mensuelle", budget_price=budget_price,
                  budget_start_date=period_start)

    # read current-period meter (last — type="reset" also renders a previous-period meter first)
    logged_in_page.goto(f"{base_url}/budget")
    meter_loc = logged_in_page.locator(f"section#budgets meter[max='{budget_price}']")
    spent_before = float(meter_loc.last.get_attribute("value"))

    insert_record(logged_in_page, base_url, created_account,
                  category=cat_name, amount="42.00", designation="budget spending test record")

    # current-period meter value must have increased by exactly 42
    logged_in_page.goto(f"{base_url}/budget")
    meter_loc = logged_in_page.locator(f"section#budgets meter[max='{budget_price}']")
    spent_after = float(meter_loc.last.get_attribute("value"))
    assert spent_after == spent_before + 42.0, f"Expected {spent_before + 42.0}, got {spent_after}"

    # restore
    edit_category(logged_in_page, base_url, cat_name,
                  budget_type="-", budget_period="-", budget_price=0)
