# Tested:
# 1.  main checklist page loads: h1 "GOFI" and h2 "Sommaire" visible
# 2.  page requires auth
# 3.  step 1 loads: "1. configuration des comptes"
# 4.  step 2 loads: "2. configuration des catégories"
# 5.  step 3 loads: "3. saisie de données"
# 6.  step 4 loads: "4. configuration de budget"
# 7.  step 5 loads: "5. stats liées au budget"
# 8.  step 6 loads: "6. stats générales"
# 9.  step 7 loads: "7. éditer une saisie"
# 10. step 8 loads: "8. annuler une saisie"
# 11. summary links navigate: click each step link from /checklist, verify h2


# 1.
def test_checklist_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist")
    assert logged_in_page.locator("h1", has_text="GOFI").is_visible()
    assert logged_in_page.locator("h2", has_text="Sommaire").is_visible()


# 2.
def test_checklist_requires_auth(page, base_url):
    page.goto(f"{base_url}/checklist")
    assert page.locator("text=Déconnecté").is_visible()


# 3.
def test_checklist_step1_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist/1")
    assert logged_in_page.locator("h2", has_text="1. configuration des comptes").is_visible()


# 4.
def test_checklist_step2_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist/2")
    assert logged_in_page.locator("h2", has_text="2. configuration des catégories").is_visible()


# 5.
def test_checklist_step3_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist/3")
    assert logged_in_page.locator("h2", has_text="3. saisie de données").is_visible()


# 6.
def test_checklist_step4_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist/4")
    assert logged_in_page.locator("h2", has_text="4. configuration de budget").is_visible()


# 7.
def test_checklist_step5_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist/5")
    assert logged_in_page.locator("h2", has_text="5. stats liées au budget").is_visible()


# 8.
def test_checklist_step6_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist/6")
    assert logged_in_page.locator("h2", has_text="6. stats générales").is_visible()


# 9.
def test_checklist_step7_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist/7")
    assert logged_in_page.locator("h2", has_text="7. éditer une saisie").is_visible()


# 10.
def test_checklist_step8_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/checklist/8")
    assert logged_in_page.locator("h2", has_text="8. annuler une saisie").is_visible()


# 11.
def test_checklist_summary_links_navigate(logged_in_page, base_url):
    steps = [
        ("1", "1. configuration des comptes"),
        ("2", "2. configuration des catégories"),
        ("3", "3. saisie de données"),
        ("4", "4. configuration de budget"),
        ("5", "5. stats liées au budget"),
        ("6", "6. stats générales"),
        ("7", "7. éditer une saisie"),
        ("8", "8. annuler une saisie"),
    ]
    for step, heading in steps:
        logged_in_page.goto(f"{base_url}/checklist")
        logged_in_page.locator("section#links").locator(f"a[href='/checklist/{step}']").click()
        logged_in_page.locator("h2", has_text=heading).wait_for()
