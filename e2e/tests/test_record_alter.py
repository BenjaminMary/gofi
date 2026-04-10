def test_record_alter_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/record/alter/edit")
    assert logged_in_page.locator("h1", has_text="Editer des gains ou dépenses").is_visible()


def test_record_alter_requires_auth(page, base_url):
    page.goto(f"{base_url}/record/alter/edit")
    assert page.locator("text=Déconnecté").is_visible()
