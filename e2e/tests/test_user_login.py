# Tested:
# 1. page loads: email/password fields, submit button, and navigation links visible
# 2. bad request: empty fields submitted bypassing HTML5 required → ERREUR1
# 3. wrong password: valid email but wrong password → error message
# 4. empty fields: HTML5 required blocks submission, form stays on page
# 5. already logged in: warning banner with current user email shown


# 1.
def test_user_login_page_loads(page, base_url):
    page.goto(f"{base_url}/user/login")
    assert page.locator("input[name='email']").is_visible()
    assert page.locator("input[name='password']").is_visible()
    assert page.locator("button[id='idSubmit']").is_visible()
    assert page.locator("a[href='/']").is_visible()
    assert page.locator("a[href='/user/create']").count() >= 2
    assert page.locator("text=Nouveau").is_visible()


# 2.
def test_user_login_erreur1_bad_request(page, base_url):
    # bypass HTML5 'required' validation to send empty fields → 400 Bad Request
    page.goto(f"{base_url}/user/login")
    page.evaluate("document.querySelector('input[name=\"email\"]').removeAttribute('required')")
    page.evaluate("document.querySelector('input[name=\"password\"]').removeAttribute('required')")
    page.locator("button[type='submit']").click()
    page.wait_for_timeout(500)
    assert page.locator("text=ERREUR1: Impossible de se connecter").is_visible()


# 3.
def test_user_login_wrong_password(page, base_url, created_user):
    # valid email but wrong password → error response
    page.goto(f"{base_url}/user/login")
    page.locator("input[name='email']").fill(created_user)
    page.locator("input[name='password']").fill("wrongpassword")
    page.locator("button[type='submit']").click()
    page.wait_for_timeout(500)
    assert page.locator("text=Impossible de se connecter").is_visible()


# 4.
def test_user_login_empty_fields_blocked(page, base_url):
    # HTML5 required validation prevents submission — form stays on page
    page.goto(f"{base_url}/user/login")
    page.locator("button[type='submit']").click()
    assert page.locator("input[name='email']").is_visible()


# 5.
def test_user_login_already_logged_in_warning(logged_in_page, base_url, created_user):
    logged_in_page.goto(f"{base_url}/user/login")
    assert logged_in_page.locator("article.alert-warning").is_visible()
    assert logged_in_page.locator("article.alert-warning code").inner_text() == created_user
    assert logged_in_page.locator("input[name='email']").is_visible()
