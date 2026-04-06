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


def test_user_create_erreur1_bad_request(page, base_url):
    # bypass HTML5 'required' validation to send empty fields → 400 Bad Request
    page.goto(f"{base_url}/user/create")
    page.evaluate("document.querySelector('input[name=\"email\"]').removeAttribute('required')")
    page.evaluate("document.querySelector('input[name=\"password\"]').removeAttribute('required')")
    page.locator("button[type='submit']").click()
    page.wait_for_timeout(500)
    assert page.locator("text=ERREUR1: Impossible de créer le compte").is_visible()


def test_user_create_erreur2_duplicate(page, base_url, test_email):
    # submit the same email twice → 500 Internal Server Error (DB unique constraint)
    page.goto(f"{base_url}/user/create")
    page.locator("input[name='email']").fill(test_email)
    page.locator("input[name='password']").fill("testpassword")
    page.locator("button[type='submit']").click()
    page.wait_for_timeout(500)
    assert page.locator("text=ERREUR2: Impossible de créer le compte").is_visible()


def test_user_create_empty_fields_blocked(page, base_url):
    # HTML5 required validation prevents submission — form stays on page
    page.goto(f"{base_url}/user/create")
    page.locator("button[type='submit']").click()
    assert page.locator("input[name='email']").is_visible()
