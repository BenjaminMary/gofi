import os
import uuid
import pytest
from playwright.sync_api import sync_playwright

# scope="session"  → runs once for the entire test session
# scope="function" → runs once per test function (default)


@pytest.fixture(scope="session")
def base_url():
    return os.environ.get("GOFI_BASE_URL", "http://localhost:8083")


@pytest.fixture(scope="session")
def test_email():
    # unique email per test session to avoid conflicts on replay
    return f"playwright-{uuid.uuid4().hex[:8]}@test.test"


@pytest.fixture(scope="session")
def test_password():
    # unique password per test session
    return uuid.uuid4().hex


@pytest.fixture(scope="session")
def playwright_instance():
    with sync_playwright() as p:
        yield p


@pytest.fixture(scope="session")
def browser(playwright_instance):
    # one browser instance shared across all tests
    # set HEADED=true to see the browser during test execution
    headless = os.environ.get("HEADED", "false").lower() != "true"
    browser = playwright_instance.chromium.launch(headless=headless, slow_mo=100)
    yield browser
    browser.close()


@pytest.fixture(scope="function")
def page(browser):
    # fresh page for every test, closed after each test
    context = browser.new_context()
    page = context.new_page()
    page.set_default_timeout(5000)
    yield page
    context.close()


@pytest.fixture(scope="session")
def created_user(browser, base_url, test_email, test_password):
    # scope="session": creates the user once in DB, required by login tests
    # test_email uses a fresh UUID each session so no conflict on reruns
    context = browser.new_context()
    page = context.new_page()
    page.goto(f"{base_url}/user/create")
    page.locator("input[name='email']").fill(test_email)
    page.locator("input[name='password']").fill(test_password)
    page.locator("button[type='submit']").click()
    page.wait_for_timeout(500)
    assert page.locator("text=Création du compte terminée").is_visible()
    context.close()
    return test_email


@pytest.fixture(scope="session")
def auth_state(browser, base_url, created_user, test_password):
    # log in once, save session cookies — reused by all logged-in tests
    context = browser.new_context()
    page = context.new_page()
    page.goto(f"{base_url}/user/login")
    page.locator("input[name='email']").fill(created_user)
    page.locator("input[name='password']").fill(test_password)
    page.locator("button[type='submit']").click()
    page.wait_for_timeout(500)
    assert page.locator("text=Login réussi").is_visible()
    state = context.storage_state()
    context.close()
    return state


@pytest.fixture(scope="function")
def logged_in_page(browser, auth_state):
    # fresh page per test, pre-authenticated via saved session cookies
    context = browser.new_context(storage_state=auth_state)
    page = context.new_page()
    page.set_default_timeout(5000)
    yield page
    context.close()


def create_account(browser, base_url, auth_state, account_name):
    """Create a param account via the /param/account UI and return its name.

    Call this from session-scoped fixtures (pass browser + auth_state) or
    directly from a test function when you need an ad-hoc account.
    """
    context = browser.new_context(storage_state=auth_state)
    page = context.new_page()
    page.goto(f"{base_url}/param/account")
    page.locator("section#createAccSection summary").click()
    page.locator("input#accountToCreate").fill(account_name)
    page.locator("button#createAccount").click()
    page.wait_for_timeout(500)
    assert page.locator(f"text={account_name}").first.is_visible()
    context.close()
    return account_name


@pytest.fixture(scope="session")
def created_account(browser, base_url, auth_state):
    # scope="session": creates account "PCB" once, required by record insert tests
    return create_account(browser, base_url, auth_state, "PCB")


def open_advanced_mode_and_reload(page, account, checked="0"):
    """Open the advanced mode filter panel and trigger the HTMX table reload.

    Mimics what a user does: click the advanced mode section to expand filters,
    set the validation filter, then switch account to fire the JS change event
    which POSTs to /record/getviapost and replaces #recap with fresh rows.

    checked: "0"=Toutes, "1"=Oui (validated), "2"=Non (default)
    """
    page.locator("#advancedMode").click()
    page.locator("#checked").select_option(checked)
    # account select starts on "-" (value="") — switching to any account fires the change event
    page.locator("#compte").select_option(account)
    page.wait_for_load_state("networkidle")


def edit_category(page, base_url, cat_name, where_to_use=None, budget_type=None,
                  budget_period=None, budget_price=None, budget_start_date=None):
    """Edit a category's settings via /param/category.

    Only the keyword args that are not None are changed; the rest keep
    whatever value the JS pre-filled from the stored JSON.

    where_to_use      : "all" | "basic" | "periodic"
    budget_type       : "-" | "reset" | "cumulative"
    budget_period     : "-" | "mensuelle" | "annuelle" | "hebdomadaire"
    budget_price      : int (pass 0 to clear the budget)
    budget_start_date : ISO date string "YYYY-MM-DD" (sets the period start)
    """
    page.goto(f"{base_url}/param/category")
    row = page.locator("#tableActiveCat tr").filter(
        has=page.locator("small", has_text=cat_name)
    ).first
    row.locator("button[id^='e-']").click()
    page.wait_for_selector("button#editRR", state="visible")
    if where_to_use is not None:
        page.locator("select#catWhereToUse").select_option(where_to_use)
    if budget_type is not None:
        page.locator("select#budgetType").select_option(budget_type)
    if budget_period is not None:
        page.locator("select#budgetPeriod").select_option(budget_period)
    if budget_price is not None:
        page.locator("input#budgetPrice").fill(str(budget_price))
    if budget_start_date is not None:
        page.locator("input#BudgetCurrentPeriodStartDate").fill(budget_start_date)
    with page.expect_navigation():
        page.locator("button#editRR").click()


def insert_record(page, base_url, account, designation="test playwright", amount="10.00",
                  direction="expense", category=None):
    """Insert one record via the /record/insert/ form.

    Use this helper inside a test when you need a fresh record at a specific
    point (e.g. before validating then cancelling). The page is left on
    /record/insert/ after the call.

    direction: "expense" (default) or "gain"
    category : category name to select (default: first radio in the list)
    """
    page.goto(f"{base_url}/record/insert/")
    page.locator("select[name='compte']").select_option(account)
    if category is not None:
        # radios are inside a <details> that starts closed — open it first so the target is interactable
        page.locator("#categoryDropdown summary").click()
        page.locator(f"input[type='radio'][name='categorie'][value='{category}']").check()
    else:
        # first radio is pre-checked in HTML (categoryNumber=0) — .check() is a no-op, no click needed
        page.locator("input[type='radio'][name='categorie']").first.check()
    page.locator("input[name='prix']").fill(amount)
    page.locator(f"input[value='{direction}']").check()
    page.locator("input[name='designation']").fill(designation)
    page.locator("button#idSubmit1").click()
    page.wait_for_selector(f"text={designation}")  # wait for HTMX response before returning


