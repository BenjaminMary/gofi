# Tested:
# 1. page loads: "Déconnexion réussi" and navigation links visible
# 2. session cleared: visiting a protected page after logout shows "Déconnecté"


# 1.
def test_user_logout_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/user/logout")
    assert logged_in_page.locator("text=Déconnexion réussi").is_visible()
    assert logged_in_page.locator("a[href='/']").is_visible()
    assert logged_in_page.locator("a[href='/user/login']").is_visible()


# 2.
def test_user_logout_clears_session(logged_in_page, base_url):
    # after logout, visiting a protected page should redirect to lost/login
    logged_in_page.goto(f"{base_url}/user/logout")
    logged_in_page.goto(f"{base_url}/record/insert/")
    assert logged_in_page.locator("text=Déconnecté").is_visible()
