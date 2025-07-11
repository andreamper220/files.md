const urlsToCache = [
    '/',
    '/favicon.ico',
    '/icon.png',
    '/icon_small.png',
    '/manifest.json',
    '/app.css',
    '/lib/normalize.css',
    '/lib/sidebar.css',
    '/lib/codemirror.css',
    '/lib/hypermd.css',
    '/lib/theme-light.css',
    '/lib/theme-dark.css',
    '/chat.css',
    '/lib/sidebar.js',
    '/lib/codemirror.js',
    '/lib/core.js',
    '/lib/markdown.js',
    '/lib/hypermd.js',
    '/lib/keymap.js',
    '/lib/click.js',
    '/lib/hide-token.js',
    '/lib/fold.js',
    '/lib/fold-image.js',
    '/lib/fold-link.js',
    '/lib/table-align.js',
    '/lib/autocomplete-link.js',
    '/lib/show-hint.js',
    '/lib/autoscroll.js',
    '/lib/codemirror-go.js',
    '/lib/codemirror-php.js',
    '/lib/codemirror-shell.js',
    '/lib/similarity.js',
    '/lib/emoji.js',
    '/welcome.js',
    '/files.js',
    '/wasm_exec.js',
    '/app.js',
    '/wasm.js',
    '/chat.js',
    '/modals.js',
];

const urlParams = new URLSearchParams(self.location.search);
const COMMIT_HASH = urlParams.get('v') ? `?v=${urlParams.get('v')}` : '';

const cacheName = `files-md-v${COMMIT_HASH}`;


self.addEventListener('install', event => {
    event.waitUntil(
        caches.open(cacheName)
            .then(cache => {
                const cachePromises = urlsToCache.map(url => {
                    if (url !== "/" && url !== 'favicon.ico' && url !== 'small_icon.png' && url !== 'icon.png') {
                        url += COMMIT_HASH;
                    }
                    return cache.add(url)
                        .catch(err => console.error('✗ Failed to cache:', url, err));
                });
                return Promise.allSettled(cachePromises); // Won't fail if one fails
            })
            .then(() => {
                return self.skipWaiting();
            })
    );
});

self.addEventListener("activate", (event) => {
    console.log("Service worker is activated");

    event.waitUntil(
        caches.keys().then((cacheNames) => {
            return cacheNames.map((cache) => {
                if (cache !== cacheName) {
                    caches.delete(cache);
                }
            });
        })
    );
});

self.addEventListener("fetch", (event) => {
    // Skip chrome-extension URLs and non-GET requests for caching
    if (event.request.method !== 'GET' || event.request.url.startsWith('chrome-extension:') ||
        event.request.url.startsWith('moz-extension:')) {
        return; // Let the browser handle it normally
    }

    event.respondWith(
        fetch(event.request)
            .then(response => {
                if (response.ok) {
                    const responseClone = response.clone();
                    caches.open(cacheName).then(cache => {
                        cache.put(event.request, responseClone);
                    });
                }
                return response;
            })
            .catch(() => {
                // Network failed, try cache (only works for GET anyway)
                return caches.match(event.request)
                    .then(cached => {
                        if (cached) return cached;
                        return new Response('Offline and not cached', {
                            status: 503,
                            statusText: 'Service Unavailable'
                        });
                    });
            })
    );
});