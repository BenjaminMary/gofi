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


def test_record_lend_or_borrow_reimburse_borrow(logged_in_page, base_url, created_account):
    # mode 3 = Rembourse emprunt: uses an existing tier from select#who (option value = tier name)
    # "Tiers LB" was created in test_record_lend_or_borrow_borrow_success which runs first (b < r)
    # in mode 3, createLenderBorrowerName is disabled — must pick from select#who
    logged_in_page.goto(f"{base_url}/record/lend-or-borrow")
    logged_in_page.locator("select#modeStr").select_option("3")
    # JS shows #whoDiv when mode is 3 — wait for select#who to become visible
    logged_in_page.wait_for_selector("select#who", state="visible")
    logged_in_page.locator("select#who").select_option("Tiers LB")
    logged_in_page.locator("select[name='FT.compte']").select_option(created_account)
    # category radio is pre-checked by the JS mode handler — do not call .check()
    logged_in_page.locator("input[name='FT.prix']").fill("25.00")
    logged_in_page.locator("input[name='FT.designation']").fill("test remboursement emprunt")
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_selector("text=test remboursement emprunt")


def test_record_lend_or_borrow_reimburse_lend(logged_in_page, base_url, created_account):
    # mode 4 = On me rembourse un prêt: pick the first available tier from select#who
    # mode 4 is the default selected value — change event won't fire, but whoDiv is already visible
    # "lend_success" < "reimburse_lend" alphabetically — at least "Tiers LB" exists by then
    logged_in_page.goto(f"{base_url}/record/lend-or-borrow")
    # mode 4 is already selected by default (first option) — whoDiv and select#who are already visible
    logged_in_page.wait_for_selector("select#who", state="visible")
    # select the first available tier — avoids encoding issues with special characters
    logged_in_page.locator("select#who").select_option(index=0)
    logged_in_page.locator("select[name='FT.compte']").select_option(created_account)
    logged_in_page.locator("input[name='FT.prix']").fill("20.00")
    logged_in_page.locator("input[name='FT.designation']").fill("test remboursement pret")
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_selector("text=test remboursement pret")


# STATS PART

def test_stats_lender_borrower_state_change(logged_in_page, base_url):
    # "Tiers LB" was created in test_record_lend_or_borrow.py (runs before this file alphabetically)
    # clicking input[id^='active-'] triggers JS: removes element + submits form2 (full page POST)
    # after reload, Tiers LB should appear in #lbTableRowsInactive
    logged_in_page.goto(f"{base_url}/stats/lender-borrower/0")
    logged_in_page.wait_for_selector("#lbTableRows")
    assert logged_in_page.locator("#lbTableRows tr", has_text="Tiers LB").count() >= 1
    with logged_in_page.expect_navigation():
        logged_in_page.locator("#lbTableRows tr", has_text="Tiers LB").first.locator("input[id^='active-']").click()  # [id^='active-'] = CSS "starts with": matches active-1, active-2, …
    assert logged_in_page.locator("#lbTableRowsInactive tr", has_text="Tiers LB").count() >= 1


def test_stats_lender_borrower_state_reactivate(logged_in_page, base_url):
    # runs after test_stats_lender_borrower_state_change ("state_change" < "state_reactivate" alphabetically)
    # "Tiers LB" was deactivated in state_change — click input[id^='inactive-'] to reactivate it
    logged_in_page.goto(f"{base_url}/stats/lender-borrower/0")
    logged_in_page.wait_for_selector("#lbTableRowsInactive")
    assert logged_in_page.locator("#lbTableRowsInactive tr", has_text="Tiers LB").count() >= 1
    with logged_in_page.expect_navigation():
        logged_in_page.locator("#lbTableRowsInactive tr", has_text="Tiers LB").first.locator("input[id^='inactive-']").click()  # [id^='inactive-'] = CSS "starts with": matches inactive-1, …
    assert logged_in_page.locator("#lbTableRows tr", has_text="Tiers LB").count() >= 1


def test_stats_lender_borrower_tier_details(logged_in_page, base_url):
    # /stats/lender-borrower/0 auto-selects lbListActive[0] and renders its details immediately
    # the h2 in section#lendBorrowDetails shows "Détails de: <tier name>"
    # no click needed — server-side rendering already populates the details section
    logged_in_page.goto(f"{base_url}/stats/lender-borrower/0")
    logged_in_page.wait_for_selector("#lbTableRows tr")
    assert logged_in_page.locator("section#lendBorrowDetails h2", has_text="Détails de").is_visible()
