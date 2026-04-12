# Tested:
# 1. offline home page: gopher image, login and create links visible
# 2. online simple mode: user email shown, simple-mode links visible, advanced links hidden
# 3. online advanced mode: switching mode reveals transfer, recurrent, lend/borrow, validate, cancel, csv links
# 4. simple mode links navigate: click each simple-mode link, verify destination h1/h2 (logout excluded — covered by test_user_logout.py)
# 5. advanced mode links navigate: switch to advanced mode, click each link, verify destination h1

# home_offline < home_online_advanced < home_online_simple (alphabetical order)

# 1.
def test_home_offline_page_loads(page, base_url):
    page.goto(base_url)
    assert page.locator("img[alt='Gopher']").is_visible()
    assert page.locator("a[href='/user/create']").is_visible()
    assert page.locator("a[href='/user/login']").is_visible()
    assert page.locator("code").count() == 0


# 2.
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


# 3.
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


# 4.
def test_home_simple_links_navigate(logged_in_page, base_url):
    simple_links = [
        ("/checklist", "h2", "Sommaire"),
        ("/record/insert/", "h1", "Insérer des données"),
        ("/record/alter/edit", "h1", "Editer des gains ou dépenses"),
        ("/stats/false-0-false-false", "h1", "Statistiques"),
        ("/budget", "h1", "Budgets"),
        ("/param/account", "h1", "Gérer les comptes"),
    ]
    for href, tag, text in simple_links:
        logged_in_page.goto(base_url)
        logged_in_page.locator(f"a[href='{href}']").click()
        logged_in_page.locator(tag, has_text=text).wait_for()


# 5.
def test_home_advanced_links_navigate(logged_in_page, base_url):
    advanced_links = [
        ("/record/transfer", "h1", "Transfert"),
        ("/record/recurrent", "h1", "Enregistrements réguliers"),
        ("/record/lend-or-borrow", "h1", "Prêt / Emprunt"),
        ("/record/alter/validate", "h1", "Valider des gains ou dépenses"),
        ("/record/alter/cancel", "h1", "Annuler des gains ou dépenses"),
        ("/csv/export", "h1", "Export CSV"),
        ("/csv/import", "h1", "Import CSV"),
    ]
    for href, tag, text in advanced_links:
        logged_in_page.goto(base_url)
        logged_in_page.locator("#switchMode").click()
        logged_in_page.locator(f"a[href='{href}']").click()
        logged_in_page.locator(tag, has_text=text).wait_for()
