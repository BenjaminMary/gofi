def test_record_recurrent_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/record/recurrent")
    assert logged_in_page.locator("h1", has_text="Enregistrements réguliers").is_visible()
    assert logged_in_page.locator("section#form").is_visible()


def test_record_recurrent_requires_auth(page, base_url):
    page.goto(f"{base_url}/record/recurrent")
    assert page.locator("text=Déconnecté").is_visible()
