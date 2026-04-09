def test_home_offline_page_loads(page, base_url):
    page.goto(base_url)
    assert page.locator("img[alt='Gopher']").is_visible()
    assert page.locator("a[href='/user/create']").is_visible()
    assert page.locator("a[href='/user/login']").is_visible()
    assert page.locator("code").count() == 0


def test_home_online_simple_mode(logged_in_page, base_url, created_user):
    logged_in_page.goto(base_url)
    assert logged_in_page.locator("img[alt='Gopher']").is_visible()
    assert logged_in_page.locator("code").inner_text() == created_user
    # simple mode links
    assert logged_in_page.locator("a[href='/checklist']").is_visible()
    assert logged_in_page.locator("a[href='/record/insert/']").is_visible()
    assert logged_in_page.locator("a[href='/record/alter/edit']").is_visible()
    assert logged_in_page.locator("a[href='/stats/false-0-false-false']").is_visible()
    assert logged_in_page.locator("a[href='/budget']").is_visible()
    assert logged_in_page.locator("a[href='/param/account']").is_visible()
    assert logged_in_page.locator("a[href='/user/logout']").is_visible()
    # advanced mode links hidden by default
    assert not logged_in_page.locator("a[href='/record/transfer']").is_visible()
    assert not logged_in_page.locator("a[href='/csv/export']").is_visible()


def test_home_online_advanced_mode(logged_in_page, base_url):
    logged_in_page.goto(base_url)
    logged_in_page.locator("#switchMode").click()
    # advanced mode links now visible
    assert logged_in_page.locator("a[href='/record/transfer']").is_visible()
    assert logged_in_page.locator("a[href='/record/recurrent']").is_visible()
    assert logged_in_page.locator("a[href='/record/lend-or-borrow']").is_visible()
    assert logged_in_page.locator("a[href='/record/alter/validate']").is_visible()
    assert logged_in_page.locator("a[href='/record/alter/cancel']").is_visible()
    assert logged_in_page.locator("a[href='/csv/export']").is_visible()
    assert logged_in_page.locator("a[href='/csv/import']").is_visible()
