def test_checklist_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist")
    assert logged_in_page.locator("h1", has_text="GOFI").is_visible()
    assert logged_in_page.locator("h2", has_text="Sommaire").is_visible()


def test_checklist_requires_auth(page, base_url):
    page.goto(f"{base_url}/checklist")
    assert page.locator("text=Déconnecté").is_visible()


def test_checklist_step1_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist/1")
    assert logged_in_page.locator("h2", has_text="1. configuration des comptes").is_visible()


def test_checklist_step2_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist/2")
    assert logged_in_page.locator("h2", has_text="2. configuration des catégories").is_visible()


def test_checklist_step3_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist/3")
    assert logged_in_page.locator("h2", has_text="3. saisie de données").is_visible()


def test_checklist_step4_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist/4")
    assert logged_in_page.locator("h2", has_text="4. configuration de budget").is_visible()


def test_checklist_step5_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist/5")
    assert logged_in_page.locator("h2", has_text="5. stats liées au budget").is_visible()


def test_checklist_step6_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist/6")
    assert logged_in_page.locator("h2", has_text="6. stats générales").is_visible()


def test_checklist_step7_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist/7")
    assert logged_in_page.locator("h2", has_text="7. éditer une saisie").is_visible()


def test_checklist_step8_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist/8")
    assert logged_in_page.locator("h2", has_text="8. annuler une saisie").is_visible()
