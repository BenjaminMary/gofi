def test_record_lend_or_borrow_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/record/lend-or-borrow")
    assert logged_in_page.locator("h1", has_text="Prêt / Emprunt").is_visible()
    assert logged_in_page.locator("section#form").is_visible()
    assert logged_in_page.locator("button#idSubmit1").is_visible()


def test_record_lend_or_borrow_requires_auth(page, base_url):
    page.goto(f"{base_url}/record/lend-or-borrow")
    assert page.locator("text=Déconnecté").is_visible()


def test_record_lend_or_borrow_mode_select_visible(logged_in_page, base_url):
    # the mode select (modeStr) should be visible with lend/borrow options
    logged_in_page.goto(f"{base_url}/record/lend-or-borrow")
    assert logged_in_page.locator("select#modeStr").is_visible()
    assert logged_in_page.locator("select#modeStr option[value='1']").count() == 1
    assert logged_in_page.locator("select#modeStr option[value='2']").count() == 1


def test_record_lend_or_borrow_success(logged_in_page, base_url, created_account):
    # select mode 1 (J'emprunte), create a new tier "Tiers LB", fill the form and submit
    logged_in_page.goto(f"{base_url}/record/lend-or-borrow")
    logged_in_page.locator("select#modeStr").select_option("1")
    logged_in_page.wait_for_timeout(200)
    logged_in_page.locator("input#createLenderBorrowerName").fill("Tiers LB")
    logged_in_page.locator("select[name='FT.compte']").select_option(created_account)
    # category radio is pre-checked (index 0) by the template — no need to interact with it
    logged_in_page.locator("input[name='FT.prix']").fill("100.00")
    logged_in_page.locator("input[name='FT.designation']").fill("test emprunt")
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_timeout(500)
    assert logged_in_page.locator("text=test emprunt").is_visible()
