# Tested:
# 1. page loads: email/password fields and submit button visible
# 2. bad request: empty fields submitted bypassing HTML5 required → ERREUR1
# 3. duplicate email: same account submitted twice → ERREUR2
# 4. empty fields: HTML5 required blocks submission, form stays on page
# 5. already logged in: warning banner with current user email shown


# 1.
def test_user_create_page_loads(page, base_url):
    page.goto(f"{base_url}/user/create")
    assert page.locator("input[name='email']").is_visible()
    assert page.locator("input[name='password']").is_visible()
    assert page.locator("button[id='idSubmit']").is_visible()
    assert page.locator("a[href='/']").is_visible()
    assert page.locator("a[href='/user/login']").is_visible()


# 2.
def test_user_create_erreur1_bad_request(page, base_url):
    # bypass HTML5 'required' validation to send empty fields → 400 Bad Request
    page.goto(f"{base_url}/user/create")
    page.evaluate("document.querySelector('input[name=\"email\"]').removeAttribute('required')")
    page.evaluate("document.querySelector('input[name=\"password\"]').removeAttribute('required')")
    page.locator("button[type='submit']").click()
    page.wait_for_timeout(500)
    assert page.locator("text=ERREUR1: Impossible de créer le compte").is_visible()


# 3.
def test_user_create_erreur2_duplicate(page, base_url, created_user, test_password):
    # submit the same email twice → 500 Internal Server Error (DB unique constraint)
    page.goto(f"{base_url}/user/create")
    page.locator("input[name='email']").fill(created_user)
    page.locator("input[name='password']").fill(test_password)
    page.locator("button[type='submit']").click()
    page.wait_for_timeout(500)
    assert page.locator("text=ERREUR2: Impossible de créer le compte").is_visible()


# 4.
def test_user_create_empty_fields_blocked(page, base_url):
    # HTML5 required validation prevents submission — form stays on page
    page.goto(f"{base_url}/user/create")
    page.locator("button[type='submit']").click()
    assert page.locator("input[name='email']").is_visible()


# 5.
def test_user_create_already_logged_in_warning(logged_in_page, base_url, created_user):
    logged_in_page.goto(f"{base_url}/user/create")
    assert logged_in_page.locator("article.alert-warning").is_visible()
    assert logged_in_page.locator("article.alert-warning code").inner_text() == created_user
    assert logged_in_page.locator("input[name='email']").is_visible()
