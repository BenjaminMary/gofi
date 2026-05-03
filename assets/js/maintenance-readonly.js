// Loaded only when appdata.ReadOnlyFlag is true (conditional <script> in Header()).
// Blocks mutating requests at the client so the user doesn't even hit the server's
// 503. The server-side middleware remains as defense-in-depth for non-browser clients.

// Disable every submit button. Pico's :disabled CSS picks up the visual styling
// automatically. A disabled button blocks both click and Enter-key activation
// natively, and HTMX won't trigger off a disabled element either.
document.addEventListener('DOMContentLoaded', function () {
    document.querySelectorAll('button[type="submit"]').forEach(function (b) {
        b.disabled = true;
    });
});

// Catch HTMX requests that don't go through a submit button: forms triggered by
// custom events (e.g. param.templ's checkbox/reorder forms with hx-trigger=...).
document.addEventListener('htmx:beforeRequest', function (e) {
    var verb = (e.detail && e.detail.requestConfig && e.detail.requestConfig.verb || '').toUpperCase();
    if (verb && verb !== 'GET' && verb !== 'HEAD') {
        e.preventDefault();
    }
});
