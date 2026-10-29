// Keeps the app working without a connection, so sightings can be reported in the field: the app
// itself, and the answers it needs to report, are kept from the last time they were fetched.
//
// Everything goes to the network first and the kept copy is only used when that fails or is too
// slow, so online the app is always current. Sightings made offline are kept by the app (Outbox.ts),
// not here.
const VERSION = "v2";
const APP_CACHE = "300sf-app-" + VERSION;
// The signed-in user's data; the app deletes it on sign-out (Session.forgetOfflineCopies).
const API_CACHE = "300sf-api";

// Without these the app can't start offline.
const APP_SHELL = [
	"/home",
	"/js/app.js",
	"/css/tokens.css",
	"/css/theme.css",
	"/css/ui.css",
	"/css/base.css",
	"/css/fonts.css",
	"/css/icons.css",
	"/fonts/inter-latin.woff2",
	"/fonts/inter-latin-ext.woff2",
	"/fonts/material-symbols-rounded.woff2",
	"/vendor/jquery/jquery-3.7.1.min.js",
	"/vendor/leaflet/leaflet.js",
	"/vendor/leaflet/leaflet.css",
	"/vendor/leaflet/images/marker-icon.png",
	"/vendor/leaflet/images/marker-icon-2x.png",
	"/vendor/leaflet/images/marker-shadow.png",
	"/img/messagebox-256.webp",
	"/favicon.png",
	"/manifest.json",
];

const STATIC_PREFIXES = ["/js/", "/css/", "/fonts/", "/vendor/", "/img/"];

// What reporting a sighting and the home page read. Other answers are never kept.
const API_KEPT = /^\/api\/v1\/(users\/whoami|users\/me|species\/list|list_templates\/list|lists\/list|lists\/species\/[^/]+|lists\/[^/]+\/reports|tags\/list|reports\/list|competitions\/list\/me|announcements\/list)$/;

// Past this, a kept copy is shown while the network keeps going to refresh it.
const PATIENCE_MS = 4000;

self.addEventListener("install", (event) => {
	event.waitUntil(caches.open(APP_CACHE).then((cache) =>
		// One missing file must not stop the rest from being kept.
		Promise.allSettled(APP_SHELL.map((url) => cache.add(new Request(url, { cache: "reload" }))))
	).then(() => self.skipWaiting()));
});

self.addEventListener("activate", (event) => {
	event.waitUntil(caches.keys().then((keys) =>
		Promise.all(keys.filter((k) => k.startsWith("300sf-app-") && k !== APP_CACHE).map((k) => caches.delete(k)))
	).then(() => self.clients.claim()));
});

self.addEventListener("fetch", (event) => {
	const request = event.request;
	const url = new URL(request.url);
	if (request.method !== "GET" || url.origin !== self.location.origin) {
		return;
	}

	if (request.mode === "navigate") {
		// Offline, a page never opened before starts as the home page.
		event.respondWith(networkFirst(event, APP_CACHE).then((r) => r || caches.match("/home", { cacheName: APP_CACHE })).then((r) => r || Response.error()));
	} else if (API_KEPT.test(url.pathname)) {
		event.respondWith(networkFirst(event, API_CACHE).then((r) => r || Response.error()));
	} else if (STATIC_PREFIXES.some((p) => url.pathname.startsWith(p)) || APP_SHELL.includes(url.pathname)) {
		event.respondWith(networkFirst(event, APP_CACHE).then((r) => r || Response.error()));
	}
});

// The network's answer, kept for later; the kept copy when the network fails or takes too long.
// Answers that aren't 200 (signed out, errors) are passed on but never kept.
function networkFirst(event, cacheName) {
	const request = event.request;
	const fromNetwork = fetch(request).then((response) => {
		if (response.status === 200) {
			const copy = response.clone();
			event.waitUntil(caches.open(cacheName).then((cache) => cache.put(request, copy)));
		}
		return response;
	});
	event.waitUntil(fromNetwork.catch(() => { }));

	return new Promise((resolve) => {
		let settled = false;
		const fallback = () => caches.match(request, { cacheName: cacheName });
		const timer = setTimeout(() => {
			fallback().then((kept) => {
				if (kept && !settled) {
					settled = true;
					resolve(kept);
				}
			});
		}, PATIENCE_MS);
		fromNetwork.then((response) => {
			clearTimeout(timer);
			if (!settled) {
				settled = true;
				resolve(response);
			}
		}, () => {
			clearTimeout(timer);
			fallback().then((kept) => {
				if (!settled) {
					settled = true;
					resolve(kept || null);
				}
			});
		});
	});
}
