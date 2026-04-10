def test_record_lend_or_borrow_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/record/lend-or-borrow")
    assert logged_in_page.locator("h1", has_text="Prêt / Emprunt").is_visible()
    assert logged_in_page.locator("section#form").is_visible()
    assert logged_in_page.locator("button#idSubmit1").is_visible()


def test_record_lend_or_borrow_requires_auth(page, base_url):
    page.goto(f"{base_url}/record/lend-or-borrow")
    assert page.locator("text=Déconnecté").is_visible()
