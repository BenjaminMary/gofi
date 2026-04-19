// sw.js — GOFI service worker, Phase 1: installable + static caching only
// Served at /sw.js (root) so its scope covers the whole site.
//
// What this does:
//   - Precaches /offline.html and the two app icons on install
//   - Runtime cache-first for /css/, /js/, /img/, /fonts/ assets
//   - Network-only for HTML pages (user-specific, server-rendered)
//   - Falls back to /offline.html on navigation requests when the network is down
//
// What this does NOT do (Phase 2):
//   - No IndexedDB queue for offline POSTs
// (Phase 3):
//   - No HTMX fragment caching
//   - No push notifications

// VERSION is only a cache-name suffix used to bust stale entries. Bump it when
// PRECACHE_URLS changes or a cached asset's content changed at the same URL —
// NOT just for logic changes (any byte change already triggers a reinstall).
// Users don't reinstall anything; the new SW takes over seamlessly on reload.
const VERSION = 'v1';
const STATIC_CACHE = `static-${VERSION}`;

const PRECACHE_URLS = [
    '/offline.html',
    '/img/android-chrome-192x192.png',
    '/img/favicon-32x32.png',
];

// install: fires once per newly-installed SW instance (any byte change to this
// file triggers a new install). Precache the offline page + app icons so the
// shell is available even on the very first offline visit.
self.addEventListener('install', (event) => {
    event.waitUntil(
        caches.open(STATIC_CACHE).then((cache) => cache.addAll(PRECACHE_URLS))
    );
    self.skipWaiting();
});

// activate: fires when the new SW takes over. Purge any cache from a previous
// VERSION and claim existing clients so they use the new SW immediately.
self.addEventListener('activate', (event) => {
    event.waitUntil(
        caches.keys().then((keys) =>
            Promise.all(
                keys.filter((k) => !k.endsWith(VERSION)).map((k) => caches.delete(k))
            )
        ).then(() => self.clients.claim())
    );
});

// fetch: intercepts every request the page makes. Routes static assets
// through the cache, navigations to the network (with offline fallback),
// and lets everything else fall through to the browser default.
self.addEventListener('fetch', (event) => {
    const { request } = event;
    const url = new URL(request.url);

    // Only handle same-origin GET requests
    if (request.method !== 'GET' || url.origin !== self.location.origin) return;

    // Static assets: cache-first with runtime population
    if (/^\/(css|js|img|fonts)\//.test(url.pathname)) {
        event.respondWith(
            caches.match(request).then((cached) => {
                if (cached) return cached;
                return fetch(request).then((response) => {
                    if (response.ok) {
                        const copy = response.clone();
                        caches.open(STATIC_CACHE).then((c) => c.put(request, copy));
                    }
                    return response;
                });
            })
        );
        return;
    }

    // Navigation requests: network-only, fall back to /offline.html if offline
    // (we do NOT cache HTML pages — they contain user-specific server-rendered data)
    if (request.mode === 'navigate') {
        event.respondWith(
            fetch(request).catch(() => caches.match('/offline.html'))
        );
        return;
    }
});

// message: lets the page tell the SW to activate a pending update immediately
// (page posts 'skip-waiting' after the user confirms a reload).
self.addEventListener('message', (event) => {
    if (event.data === 'skip-waiting') self.skipWaiting();
});
