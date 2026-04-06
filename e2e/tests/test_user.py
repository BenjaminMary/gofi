def test_user_create_page_loads(page, base_url):
    page.goto(f"{base_url}/user/create")
    assert page.locator("input[name='email']").is_visible()
    assert page.locator("input[name='password']").is_visible()


def test_user_create_success(page, base_url, test_email):
    page.goto(f"{base_url}/user/create")
    page.locator("input[name='email']").fill(test_email)
    page.locator("input[name='password']").fill("testpassword")
    page.locator("button[type='submit']").click()
    page.wait_for_timeout(500)
    assert page.locator("text=Création du compte terminée").is_visible()


def test_user_create_duplicate_fails(page, base_url, test_email):
    page.goto(f"{base_url}/user/create")
    page.locator("input[name='email']").fill(test_email)
    page.locator("input[name='password']").fill("testpassword")
    page.locator("button[type='submit']").click()
    page.wait_for_timeout(500)
    assert page.locator("text=Impossible de créer le compte").is_visible()


def test_user_create_empty_fields_blocked(page, base_url):
    page.goto(f"{base_url}/user/create")
    page.locator("button[type='submit']").click()
    # HTML5 required validation prevents submission — form stays on page
    assert page.locator("input[name='email']").is_visible()
