def test_home_offline_page_loads(page, base_url):
    page.goto(base_url)
    assert page.locator("img[alt='Gopher']").is_visible()
    assert page.locator("a[href='/user/create']").is_visible()
    assert page.locator("a[href='/user/login']").is_visible()
    assert page.locator("code").count() == 0
