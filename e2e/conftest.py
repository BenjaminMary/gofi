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
    yield page
    context.close()


@pytest.fixture(scope="session")
def created_user(browser, base_url, test_email):
    # scope="session": creates the user once in DB, required by login tests
    # test_email uses a fresh UUID each session so no conflict on reruns
    context = browser.new_context()
    page = context.new_page()
    page.goto(f"{base_url}/user/create")
    page.locator("input[name='email']").fill(test_email)
    page.locator("input[name='password']").fill("testpassword")
    page.locator("button[type='submit']").click()
    page.wait_for_timeout(500)
    assert page.locator("text=Création du compte terminée").is_visible()
    context.close()
    return test_email
