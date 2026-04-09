def test_user_login_page_loads(page, base_url):
    page.goto(f"{base_url}/user/login")
    assert page.locator("input[name='email']").is_visible()
    assert page.locator("input[name='password']").is_visible()
    assert page.locator("button[id='idSubmit']").is_visible()
    assert page.locator("a[href='/']").is_visible()
    assert page.locator("a[href='/user/create']").count() >= 2
    assert page.locator("text=Nouveau").is_visible()


def test_user_login_success(page, base_url, created_user):
    # created_user fixture ensures the user exists in DB before this test runs
    page.goto(f"{base_url}/user/login")
    page.locator("input[name='email']").fill(created_user)
    page.locator("input[name='password']").fill("testpassword")
    page.locator("button[type='submit']").click()
    page.wait_for_timeout(500)
    assert page.locator("text=Login réussi").is_visible()


def test_user_login_erreur1_bad_request(page, base_url):
    # bypass HTML5 'required' validation to send empty fields → 400 Bad Request
    page.goto(f"{base_url}/user/login")
    page.evaluate("document.querySelector('input[name=\"email\"]').removeAttribute('required')")
    page.evaluate("document.querySelector('input[name=\"password\"]').removeAttribute('required')")
    page.locator("button[type='submit']").click()
    page.wait_for_timeout(500)
    assert page.locator("text=ERREUR1: Impossible de se connecter").is_visible()


def test_user_login_wrong_password(page, base_url, created_user):
    # valid email but wrong password → error response
    page.goto(f"{base_url}/user/login")
    page.locator("input[name='email']").fill(created_user)
    page.locator("input[name='password']").fill("wrongpassword")
    page.locator("button[type='submit']").click()
    page.wait_for_timeout(500)
    assert page.locator("text=Impossible de se connecter").is_visible()


def test_user_login_empty_fields_blocked(page, base_url):
    # HTML5 required validation prevents submission — form stays on page
    page.goto(f"{base_url}/user/login")
    page.locator("button[type='submit']").click()
    assert page.locator("input[name='email']").is_visible()
