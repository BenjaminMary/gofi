# Tested:
# 1. /param page loads: h1 "Gérer les paramètres"
# 2. /param requires auth
# 3. category rendering set to "names": button#idSubmit3 disappears on success
# 4. category rendering set to "icons": button#idSubmit3 disappears on success


# 1.
def test_param_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/param")
    assert logged_in_page.locator("h1", has_text="Gérer les paramètres").is_visible()


# 2.
def test_param_requires_auth(page, base_url):
    page.goto(f"{base_url}/param")
    assert page.locator("text=Déconnecté").is_visible()


# 3.
def test_param_category_rendering_names(logged_in_page, base_url):
    # switch category rendering to "names" and save — no error should appear
    logged_in_page.goto(f"{base_url}/param")
    logged_in_page.locator("input#names").check()
    logged_in_page.locator("button#idSubmit3").click()
    logged_in_page.wait_for_timeout(500)
    # idSubmit3 is removed on success (hx-on::after-request in template)
    assert logged_in_page.locator("button#idSubmit3").count() == 0


# 4.
def test_param_category_rendering_icons(logged_in_page, base_url):
    # switch category rendering back to "icons" and save
    logged_in_page.goto(f"{base_url}/param")
    logged_in_page.locator("input#icons").check()
    logged_in_page.locator("button#idSubmit3").click()
    logged_in_page.wait_for_timeout(500)
    assert logged_in_page.locator("button#idSubmit3").count() == 0
