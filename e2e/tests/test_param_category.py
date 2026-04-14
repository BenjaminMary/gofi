from conftest import edit_category

# Tested:
# 1. /param/category page loads: h1 "Gérer les catégories"
# 2. /param/category requires auth
# 3. category edit: opens form, submit redirects back to category page
# 4. category deactivate: toggle removes category from active table; appears in inactive table on reload
# 5. category reactivate: toggle removes category from inactive table; appears in active table on reload (r > d)
# 6. category reorder: moving second active category up changes order in table and record insert select
# 7. catWhereToUse="basic": category appears in insert form, absent from recurrent form
# 8. catWhereToUse="periodic": category absent from insert form, appears in recurrent form
# 9. catWhereToUse="all": category appears in both insert and recurrent forms


def _radio_names(page, route, name="categorie"):
    """Return all radio[name={name}] values on {route}."""
    page.goto(route)
    return [r.get_attribute("value") for r in page.locator(f"input[type='radio'][name='{name}']").all()]


# 1.
def test_param_category_page_loads(logged_in_page, base_url):
    logged_in_page.goto(f"{base_url}/param/category")
    assert logged_in_page.locator("h1", has_text="Gérer les catégories").is_visible()


# 2.
def test_param_category_requires_auth(page, base_url):
    page.goto(f"{base_url}/param/category")
    assert page.locator("text=Déconnecté").is_visible()


# 3.
def test_param_category_edit(logged_in_page, base_url):
    # click edit on first active category (button id^='e-') — JS opens section#openForm
    # and populates the form fields; submit without changes triggers location.reload()
    logged_in_page.goto(f"{base_url}/param/category")
    logged_in_page.locator("button[id^='e-']").first.click()  # [id^='e-'] = CSS "starts with": matches e-0-1, e-1-3, …
    # JS sets section#openForm.hidden = false — button#editRR becomes visible
    logged_in_page.wait_for_selector("button#editRR", state="visible")
    with logged_in_page.expect_navigation():
        logged_in_page.locator("button#editRR").click()
    assert logged_in_page.locator("h1", has_text="Gérer les catégories").is_visible()


# 4.
def test_param_category_deactivate(logged_in_page, base_url):
    # JS removes the row from the active table immediately on click (optimistic),
    # then fires HTMX PATCH /param/category/in-use (hx-swap="none") — reload to confirm
    logged_in_page.goto(f"{base_url}/param/category")
    toggle = logged_in_page.locator("#tableActiveCat input[id^='desactivate-']").first
    cat_id = toggle.get_attribute("id").split("-")[1]  # "desactivate-42" → "42"
    toggle.click()
    logged_in_page.wait_for_load_state("networkidle")  # PATCH completes (hx-swap="none")
    assert logged_in_page.locator(f"#tableActiveCat tr#tr-{cat_id}").count() == 0
    # reload to confirm server persisted the change — category must appear in inactive table
    logged_in_page.reload()
    logged_in_page.wait_for_load_state("networkidle")
    assert logged_in_page.locator(f"#tableInactiveCat input#activate-{cat_id}").is_visible()
    # leave it inactive so test_param_category_reactivate (r > d) has a category to work with


# 5.
def test_param_category_reactivate(logged_in_page, base_url):
    # runs after deactivate (d < r) — at least one inactive category exists
    # JS removes the row from the inactive table immediately on click, then HTMX PATCH fires
    logged_in_page.goto(f"{base_url}/param/category")
    toggle = logged_in_page.locator("#tableInactiveCat input[id^='activate-']").first
    cat_id = toggle.get_attribute("id").split("-")[1]  # "activate-42" → "42"
    toggle.click()
    logged_in_page.wait_for_load_state("networkidle")
    assert logged_in_page.locator(f"#tableInactiveCat tr#tr-{cat_id}").count() == 0
    # reload to confirm — category must appear back in active table
    logged_in_page.reload()
    logged_in_page.wait_for_load_state("networkidle")
    assert logged_in_page.locator(f"#tableActiveCat input#desactivate-{cat_id}").is_visible()


# 6.
def test_param_category_reorder(logged_in_page, base_url):
    # read the first two rows of the active category table to get their IDs and names
    logged_in_page.goto(f"{base_url}/param/category")
    rows = logged_in_page.locator("#tableActiveCat tr[id]")
    first_row_id = rows.nth(0).get_attribute("id").split("-")[1]   # "tr-3" → "3"
    second_row_id = rows.nth(1).get_attribute("id").split("-")[1]  # "tr-7" → "7"
    first_name = rows.nth(0).locator("small").text_content().strip()
    second_name = rows.nth(1).locator("small").text_content().strip()

    # click the up-arrow button for the second category — id: u-{thisID}-{prevID}
    logged_in_page.locator(f"button#u-{second_row_id}-{first_row_id}").click()
    logged_in_page.wait_for_load_state("networkidle")  # HTMX updates #activeCategoryTable innerHTML

    # verify the table order swapped
    new_rows = logged_in_page.locator("#tableActiveCat tr[id]")
    assert new_rows.nth(0).get_attribute("id") == f"tr-{second_row_id}"
    assert new_rows.nth(1).get_attribute("id") == f"tr-{first_row_id}"

    # verify the order is also reflected in the record insert select (if both categories appear there)
    # radio value = category.Name; insert form filters type="basic"/"all" — periodic categories are excluded
    logged_in_page.goto(f"{base_url}/record/insert/")
    radio_names = [r.get_attribute("value") for r in logged_in_page.locator("input[type='radio'][name='categorie']").all()]
    if first_name in radio_names and second_name in radio_names:
        assert radio_names.index(second_name) < radio_names.index(first_name)

    # restore original order
    logged_in_page.goto(f"{base_url}/param/category")
    logged_in_page.locator(f"button#u-{first_row_id}-{second_row_id}").click()
    logged_in_page.wait_for_load_state("networkidle")


# 7.
def test_param_category_type_basic(logged_in_page, base_url):
    # set first active category to type="basic" — must appear in insert, absent from recurrent
    logged_in_page.goto(f"{base_url}/param/category")
    cat_name = logged_in_page.locator("#tableActiveCat td small").first.text_content().strip()

    edit_category(logged_in_page, base_url, cat_name, where_to_use="basic")

    assert cat_name in _radio_names(logged_in_page, f"{base_url}/record/insert/")
    assert cat_name not in _radio_names(logged_in_page, f"{base_url}/record/recurrent")

    edit_category(logged_in_page, base_url, cat_name, where_to_use="all")  # restore


# 8.
def test_param_category_type_periodic(logged_in_page, base_url):
    # set first active category to type="periodic" — absent from insert, must appear in recurrent
    logged_in_page.goto(f"{base_url}/param/category")
    cat_name = logged_in_page.locator("#tableActiveCat td small").first.text_content().strip()

    edit_category(logged_in_page, base_url, cat_name, where_to_use="periodic")

    assert cat_name not in _radio_names(logged_in_page, f"{base_url}/record/insert/")
    assert cat_name in _radio_names(logged_in_page, f"{base_url}/record/recurrent")

    edit_category(logged_in_page, base_url, cat_name, where_to_use="all")  # restore


# 9.
def test_param_category_type_all(logged_in_page, base_url):
    # set first active category to type="all" — must appear in both insert and recurrent
    logged_in_page.goto(f"{base_url}/param/category")
    cat_name = logged_in_page.locator("#tableActiveCat td small").first.text_content().strip()

    edit_category(logged_in_page, base_url, cat_name, where_to_use="all")

    assert cat_name in _radio_names(logged_in_page, f"{base_url}/record/insert/")
    assert cat_name in _radio_names(logged_in_page, f"{base_url}/record/recurrent")
