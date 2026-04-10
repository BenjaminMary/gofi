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


@pytest.fixture(scope="session")
def created_account(browser, base_url, auth_state):
    # scope="session": creates account "CB" once, required by record insert tests
    context = browser.new_context(storage_state=auth_state)
    page = context.new_page()
    page.goto(f"{base_url}/param/account")
    page.locator("section#createAccSection summary").click()
    page.locator("input#accountToCreate").fill("CB")
    page.locator("button#createAccount").click()
    page.wait_for_timeout(500)
    assert page.locator("text=CB").first.is_visible()
    context.close()
    return "CB"


@pytest.fixture(scope="session")
def la_account(browser, base_url, auth_state):
    # scope="session": creates account "LA" once, needed for transfer tests
    context = browser.new_context(storage_state=auth_state)
    page = context.new_page()
    page.goto(f"{base_url}/param/account")
    page.locator("section#createAccSection summary").click()
    page.locator("input#accountToCreate").fill("LA")
    page.locator("button#createAccount").click()
    page.wait_for_timeout(500)
    assert page.locator("text=LA").first.is_visible()
    context.close()
    return "LA"


@pytest.fixture(scope="session")
def created_record(browser, base_url, auth_state, created_account):
    # scope="session": inserts one record once, required by alter/validate tests
    context = browser.new_context(storage_state=auth_state)
    page = context.new_page()
    page.goto(f"{base_url}/record/insert/")
    page.locator("select[name='compte']").select_option(created_account)
    page.locator("input[type='radio'][name='categorie']").first.check()
    page.locator("input[name='prix']").fill("10.00")
    page.locator("input[value='expense']").check()
    page.locator("input[name='designation']").fill("test playwright")
    page.locator("button#idSubmit1").click()
    page.wait_for_timeout(500)
    context.close()

