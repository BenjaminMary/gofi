// Registers the service worker and requests persistent storage on page load.
// Loaded with <script defer> from Header() so it never blocks rendering.
if ('serviceWorker' in navigator) {
    window.addEventListener('load', function () {
        navigator.serviceWorker.register('/sw.js').catch(function () {});
        
        // Request persistent storage so the browser won't evict our caches
        // (offline.html, runtime-cached assets, future IndexedDB queue) under
        // quota pressure. Granted silently for installed PWAs on Chrome.
        if (navigator.storage && navigator.storage.persist) {
            navigator.storage.persist().catch(function () {});
        }
    });
}
