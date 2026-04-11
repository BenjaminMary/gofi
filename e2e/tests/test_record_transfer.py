def test_record_transfer_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/record/transfer")
    assert logged_in_page.locator("h1", has_text="Transfert").is_visible()
    assert logged_in_page.locator("section#form").is_visible()
    assert logged_in_page.locator("button#idSubmit1").is_visible()


def test_record_transfer_requires_auth(page, base_url):
    page.goto(f"{base_url}/record/transfer")
    assert page.locator("text=Déconnecté").is_visible()


def test_record_transfer_both_accounts_in_selects(logged_in_page, base_url, created_account, la_account):
    # CB and LA accounts should both appear in the from/to selects
    # option elements are never "visible" in Playwright — use count() instead
    logged_in_page.goto(f"{base_url}/record/transfer")
    assert logged_in_page.locator(f"select[name='compteDepuis'] option[value='{created_account}']").count() >= 1
    assert logged_in_page.locator(f"select[name='compteDepuis'] option[value='{la_account}']").count() >= 1
    assert logged_in_page.locator(f"select[name='compteVers'] option[value='{created_account}']").count() >= 1
    assert logged_in_page.locator(f"select[name='compteVers'] option[value='{la_account}']").count() >= 1


def test_record_transfer_same_account_blocked(logged_in_page, base_url, created_account):
    # JS guard: htmx:confirm fires a browser alert and blocks the request when from == to
    # all required fields must be filled first — HTML5 validation blocks submit before htmx:confirm fires
    logged_in_page.goto(f"{base_url}/record/transfer")
    dialog_messages = []
    logged_in_page.on("dialog", lambda d: (dialog_messages.append(d.message), d.accept()))
    logged_in_page.locator("select[name='compteDepuis']").select_option(created_account)
    logged_in_page.locator("select[name='compteVers']").select_option(created_account)
    logged_in_page.locator("input[name='prix']").fill("10.00")
    rows_before = logged_in_page.locator("#lastInsert tr").count()  # server may pre-fill with existing records
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_timeout(300)  # let JS process the confirm event
    assert any("différents" in m for m in dialog_messages)
    # no new rows added — count must be unchanged
    assert logged_in_page.locator("#lastInsert tr").count() == rows_before


def test_record_transfer_success(logged_in_page, base_url, created_account, la_account):
    # transfer from LA to CB — backend creates two records: Transfert- and Transfert+
    logged_in_page.goto(f"{base_url}/record/transfer")
    logged_in_page.locator("select[name='compteDepuis']").select_option(la_account)
    logged_in_page.locator("select[name='compteVers']").select_option(created_account)
    logged_in_page.locator("input[name='prix']").fill("50.00")
    logged_in_page.locator("button#idSubmit1").click()
    logged_in_page.wait_for_selector("text=Transfert+")
    assert logged_in_page.locator("text=Transfert-").is_visible()
