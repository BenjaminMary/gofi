import pytest

# Tested:
# 1. admin can hit the maintenance toggle endpoint and get 200
# 2. banner appears via body::before content when read-only flag is on
# 3. /js/maintenance-readonly.js script tag is present when flag is on
# 4. submit buttons get the disabled attribute when flag is on (Pico styles them)
# 5. attempting to submit a record while in maintenance does not create a row
# 6. when flag is off, banner is gone and submit buttons are enabled (regression)
#
# Requires admin access — see conftest's admin_session_id fixture. Tests FAIL
# (not skip) if GOFI_E2E_ADMIN_EMAIL is not set, so a misconfigured runner is
# noisy rather than silently passing.


@pytest.fixture
def maintenance_on(page, admin_session_id, base_url):
    """Enable read-only mode for the test, guarantee disable on teardown so a
    failure mid-test doesn't leave the app stuck in maintenance."""
    page.request.get(f"{base_url}/api/maintenance/readonly/on",
                     headers={"sessionID": admin_session_id})
    yield
    page.request.get(f"{base_url}/api/maintenance/readonly/off",
                     headers={"sessionID": admin_session_id})


# 1.
def test_maintenance_toggle_endpoint(page, base_url, admin_session_id):
    # admin can hit the GET toggle endpoint and get 200 for both on and off
    response = page.request.get(f"{base_url}/api/maintenance/readonly/on",
                                headers={"sessionID": admin_session_id})
    try:
        assert response.status == 200, f"toggle on returned {response.status}"
    finally:
        off = page.request.get(f"{base_url}/api/maintenance/readonly/off",
                               headers={"sessionID": admin_session_id})
        assert off.status == 200, f"toggle off returned {off.status}"


# 2.
def test_maintenance_banner_visible(logged_in_page, base_url, maintenance_on):
    # body::before content is set via CSS only when ReadOnlyFlag is on; read it
    # via getComputedStyle so we verify the rule actually applied
    logged_in_page.goto(f"{base_url}/")
    content = logged_in_page.evaluate(
        "() => getComputedStyle(document.body, '::before').content"
    )
    assert "Maintenance" in content, f"banner content not found, got: {content!r}"


# 3.
def test_maintenance_script_loaded(logged_in_page, base_url, maintenance_on):
    # the disable script must be on every page when flag is on (it adds the
    # disabled attribute to submit buttons and blocks HTMX mutating verbs)
    logged_in_page.goto(f"{base_url}/")
    assert logged_in_page.locator("script[src='/js/maintenance-readonly.js']").count() >= 1


# 4.
def test_maintenance_submit_disabled(logged_in_page, base_url, created_account, maintenance_on):
    # the JS sets disabled on every button[type="submit"] on DOMContentLoaded
    logged_in_page.goto(f"{base_url}/record/insert/")
    submit = logged_in_page.locator("button#idSubmit1")
    submit.wait_for()
    assert submit.is_disabled(), "submit button should be disabled in maintenance mode"


# 5.
def test_maintenance_form_submit_blocked(logged_in_page, base_url, created_account, maintenance_on):
    # filling the form and trying to submit should be a no-op — the disabled
    # button can't be clicked and no record reaches the recap table
    logged_in_page.goto(f"{base_url}/record/insert/")
    logged_in_page.locator("select[name='compte']").select_option(created_account)
    logged_in_page.locator("#categoryDropdown summary").click()
    logged_in_page.locator("input[type='radio'][name='categorie']").first.click()
    logged_in_page.wait_for_selector("#categoryDropdown:not([open])")
    logged_in_page.locator("input[name='prix']").fill("99.99")
    logged_in_page.locator("input[name='designation']").fill("blocked-by-maintenance")
    # button is disabled — confirm we can't click it
    button = logged_in_page.locator("button#idSubmit1")
    assert button.is_disabled()
    # give any in-flight handler a chance, then verify nothing was inserted
    logged_in_page.wait_for_timeout(300)
    assert logged_in_page.locator("text=blocked-by-maintenance").count() == 0, \
        "record should not appear in recap when maintenance is on"


# 6.
def test_maintenance_off_restores_normal(logged_in_page, base_url, admin_session_id):
    # explicitly turn flag off (idempotent), then verify the page is back to normal
    logged_in_page.request.get(f"{base_url}/api/maintenance/readonly/off",
                               headers={"sessionID": admin_session_id})
    logged_in_page.goto(f"{base_url}/record/insert/")
    submit = logged_in_page.locator("button#idSubmit1")
    submit.wait_for()
    assert not submit.is_disabled(), "submit button should be enabled when flag is off"
    # banner pseudo-element shouldn't be rendering "Maintenance" text
    content = logged_in_page.evaluate(
        "() => getComputedStyle(document.body, '::before').content"
    )
    assert "Maintenance" not in content, f"banner should be gone, got: {content!r}"
