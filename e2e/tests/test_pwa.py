import json

# Tested (Phase 1 — installable + static caching):
# 1. /manifest.json is served with the right content-type and required fields
# 2. /sw.js is served as JavaScript from the root (needed for full-site scope)
# 3. /offline.html is served and contains the expected text
# 4. every page includes <link rel="manifest"> and <meta name="theme-color">
# 5. service worker registers successfully on a real page
# 6. after SW activation, a reload makes navigator.serviceWorker.controller non-null
# 7. offline navigation falls back to /offline.html when the network is down


# 1.
def test_pwa_manifest_json(logged_in_page, base_url):
    # manifest must be served at /manifest.json (conventional root location)
    # and include the fields required for the browser to treat the app as installable
    response = logged_in_page.request.get(f"{base_url}/manifest.json")
    assert response.status == 200
    # content-type must be a JSON-ish manifest type (browsers accept both)
    ctype = response.headers.get("content-type", "")
    assert "json" in ctype, f"unexpected content-type: {ctype}"
    data = json.loads(response.text())
    # required fields for installability
    assert data.get("name") == "GOFI"
    assert data.get("start_url") == "/"
    assert data.get("display") == "standalone"
    assert data.get("scope") == "/"
    # must include at least a 192 and 512 icon
    sizes = {icon["sizes"] for icon in data.get("icons", [])}
    assert "192x192" in sizes
    assert "512x512" in sizes


# 2.
def test_pwa_service_worker_served_at_root(logged_in_page, base_url):
    # /sw.js must be served from root (not /js/sw.js) so the SW scope
    # covers the whole origin
    response = logged_in_page.request.get(f"{base_url}/sw.js")
    assert response.status == 200
    ctype = response.headers.get("content-type", "")
    assert "javascript" in ctype, f"unexpected content-type: {ctype}"
    body = response.text()
    # quick sanity check — the file must be the SW, not some other JS
    assert "serviceWorker" in body or "self.addEventListener" in body


# 3.
def test_pwa_offline_page_served(logged_in_page, base_url):
    # /offline.html is the static fallback the SW returns for failed navigations
    response = logged_in_page.request.get(f"{base_url}/offline.html")
    assert response.status == 200
    ctype = response.headers.get("content-type", "")
    assert "html" in ctype
    body = response.text()
    assert "Hors ligne" in body


# 4.
def test_pwa_manifest_and_theme_color_in_head(logged_in_page, base_url):
    # every page must declare its manifest and theme color for the browser
    # to recognize installability and paint the chrome correctly
    logged_in_page.goto(f"{base_url}/")
    assert logged_in_page.locator("link[rel='manifest'][href='/manifest.json']").count() >= 1
    assert logged_in_page.locator("meta[name='theme-color']").count() >= 1


# 5.
def test_pwa_service_worker_registers(logged_in_page, base_url):
    # on page load, the inline script in Header() must register /sw.js
    # and the registration must resolve (not throw)
    logged_in_page.goto(f"{base_url}/")
    # wait for registration and activation
    logged_in_page.evaluate("() => navigator.serviceWorker.ready")
    registrations = logged_in_page.evaluate(
        "async () => (await navigator.serviceWorker.getRegistrations()).map(r => r.scope)"
    )
    assert any(scope.rstrip("/") == base_url.rstrip("/") for scope in registrations), \
        f"expected a registration with scope {base_url}, got {registrations}"


# 6.
def test_pwa_service_worker_controls_page_after_reload(logged_in_page, base_url):
    # a page loaded BEFORE the SW activates is NOT controlled by it;
    # after a reload, navigator.serviceWorker.controller must be set
    logged_in_page.goto(f"{base_url}/")
    logged_in_page.evaluate("() => navigator.serviceWorker.ready")
    logged_in_page.reload()
    logged_in_page.evaluate("() => navigator.serviceWorker.ready")
    has_controller = logged_in_page.evaluate("() => navigator.serviceWorker.controller !== null")
    assert has_controller, "SW did not take control of the page after reload"


# 7.
def test_pwa_offline_fallback(logged_in_page, base_url):
    # simulate offline, navigate to any page, the SW must serve /offline.html
    logged_in_page.goto(f"{base_url}/")
    logged_in_page.evaluate("() => navigator.serviceWorker.ready")
    # reload so the SW controls the page
    logged_in_page.reload()
    logged_in_page.evaluate("() => navigator.serviceWorker.ready")
    # now go offline and try to navigate
    logged_in_page.context.set_offline(True)
    try:
        logged_in_page.goto(f"{base_url}/checklist")
        # the SW's catch handler returns /offline.html on navigation failure
        assert logged_in_page.locator("text=Hors ligne").is_visible()
    finally:
        logged_in_page.context.set_offline(False)
