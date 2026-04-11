def test_record_lend_or_borrow_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/record/lend-or-borrow")
    assert logged_in_page.locator("h1", has_text="Prêt / Emprunt").is_visible()
    assert logged_in_page.locator("section#form").is_visible()
    assert logged_in_page.locator("button#idSubmit1").is_visible()


def test_record_lend_or_borrow_requires_auth(page, base_url):
    page.goto(f"{base_url}/record/lend-or-borrow")
    assert page.locator("text=Déconnecté").is_visible()


def test_record_lend_or_borrow_mode_select_visible(logged_in_page, base_url):
    # the mode select (modeStr) should be visible with all 4 lend/borrow options
    logged_in_page.goto(f"{base_url}/record/lend-or-borrow")
    assert logged_in_page.locator("select#modeStr").is_visible()
    assert logged_in_page.locator("select#modeStr option[value='1']").count() == 1  # J'emprunte
    assert logged_in_page.locator("select#modeStr option[value='2']").count() == 1  # Je prête
    assert logged_in_page.locator("select#modeStr option[value='3']").count() == 1  # Rembourse emprunt
    assert logged_in_page.locator("select#modeStr option[value='4']").count() == 1  # Rembourse prêt


def test_record_lend_or_borrow_borrow_success(logged_in_page, base_url, created_account):
    # mode 1 = J'emprunte: create new tier "Tiers LB", fill form, verify row appears in recap
    logged_in_page.goto(f"{base_url}/record/lend-or-borrow")
    logged_in_page.locator("select#modeStr").select_option("1")
    logged_in_page.wait_for_selector("input#createLenderBorrowerName")
    logged_in_page.locator("input#createLenderBorrowerName").fill("Tiers LB")
    logged_in_page.locator("select[name='FT.compte']").select_option(created_account)
    # category radio is pre-checked (index 0) by the JS mode handler — do not call .check()
    logged_in_page.locator("input[name='FT.prix']").fill("100.00")
    logged_in_page.locator("input[name='FT.designation']").fill("test emprunt")
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_selector("text=test emprunt")


def test_record_lend_or_borrow_lend_success(logged_in_page, base_url, created_account):
    # mode 2 = Je prête: create new tier "Tiers Prêt", fill form, verify row appears in recap
    logged_in_page.goto(f"{base_url}/record/lend-or-borrow")
    logged_in_page.locator("select#modeStr").select_option("2")
    logged_in_page.wait_for_selector("input#createLenderBorrowerName")
    logged_in_page.locator("input#createLenderBorrowerName").fill("Tiers Prêt")
    logged_in_page.locator("select[name='FT.compte']").select_option(created_account)
    # category radio is pre-checked (index 0) by the JS mode handler — do not call .check()
    logged_in_page.locator("input[name='FT.prix']").fill("50.00")
    logged_in_page.locator("input[name='FT.designation']").fill("test pret")
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_selector("text=test pret")
