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
# 12. step pages require auth: /checklist/1 redirects to login when unauthenticated
# 13. section content rendered: section#N is present on each step page
# 14. CTA links: each step's action button points to the correct feature page
# 15. next step navigation: step 1 has "Etape 2/8" link pointing to /checklist/2
# 16. prev/next navigation on middle step: step 4 has links to /checklist/3 and /checklist/5
# 17. visit tracking: visiting a step causes its icon to appear checked in the summary


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


# 12.
def test_checklist_step_requires_auth(page, base_url):
    page.goto(f"{base_url}/checklist/1")
    assert page.locator("text=Déconnecté").is_visible()


# 13.
def test_checklist_step_sections_rendered(logged_in_page, base_url):
    # each step page must render its content inside section#N
    for step in range(1, 9):
        logged_in_page.goto(f"{base_url}/checklist/{step}")
        assert logged_in_page.locator(f"section[id='{step}']").is_visible(), \
            f"section[id='{step}'] not visible on /checklist/{step}"


# 14.
def test_checklist_cta_links(logged_in_page, base_url):
    # each step's action button must point to the correct feature page
    cta_map = [
        (1, "/param/account"),
        (2, "/param/category"),
        (3, "/record/insert/"),
        (4, "/param/category"),
        (5, "/budget"),
        (6, "/stats/false-0-false-false"),
        (7, "/record/alter/edit"),
    ]
    for step, href in cta_map:
        logged_in_page.goto(f"{base_url}/checklist/{step}")
        assert logged_in_page.locator(f"section[id='{step}'] a[href='{href}']").count() >= 1, \
            f"step {step}: expected CTA link to '{href}'"


# 15.
def test_checklist_step1_next_nav(logged_in_page, base_url):
    # step 1 has no previous step — only a "Etape 2/8" forward link
    logged_in_page.goto(f"{base_url}/checklist/1")
    with logged_in_page.expect_navigation():
        logged_in_page.locator("a[href='/checklist/2']").click()
    logged_in_page.locator("h2", has_text="2. configuration des catégories").wait_for()


# 16.
def test_checklist_middle_step_prev_next_nav(logged_in_page, base_url):
    # step 4 must have both a prev link (step 3) and a next link (step 5)
    logged_in_page.goto(f"{base_url}/checklist/4")
    assert logged_in_page.locator("a[href='/checklist/3']").count() >= 1
    assert logged_in_page.locator("a[href='/checklist/5']").count() >= 1
    # clicking next navigates to step 5
    with logged_in_page.expect_navigation():
        logged_in_page.locator("a[href='/checklist/5']").click()
    logged_in_page.locator("h2", has_text="5. stats liées au budget").wait_for()


# 17.
def test_checklist_visit_tracking(logged_in_page, base_url):
    # visiting a step records it in OnboardingCheckList — the summary must show
    # the checked icon (class "contrast outline") for that step
    logged_in_page.goto(f"{base_url}/checklist/2")
    logged_in_page.locator("h2", has_text="2. configuration des catégories").wait_for()
    logged_in_page.goto(f"{base_url}/checklist")
    # the link for step 2 must have class "contrast outline" (checked state)
    step2_link = logged_in_page.locator("section#links a[href='/checklist/2']")
    assert "outline" in (step2_link.get_attribute("class") or ""), \
        "step 2 link should have class 'contrast outline' after visiting /checklist/2"
