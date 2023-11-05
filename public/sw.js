const VERSION = "v2";

const CACHE_NAME = `300sf-${VERSION}`;

const APP_STATIC_RESOURCES = [
	"/",
	"/js/app.js",
	"/css/base.css",
	"/css/dark.css",
	"/css/ui.css",
	"/img/3f638133-a433-439b-919c-fdef170b882e.png",
	"/img/300sf.svg",
	"/img/856ff37d-1d00-497d-b493-d45794dca191.png",
	"/img/99249498-452f-4b0b-a1f5-45902f7ffe48.png",
	"/img/error.png",
	"/img/logo-256x256.png"
];

self.addEventListener("install", (event) => {
	event.waitUntil(
		(async () => {
			const cache = await caches.open(CACHE_NAME);
			cache.addAll(APP_STATIC_RESOURCES);
		})(),
	);
});

self.addEventListener("activate", (event) => {
	event.waitUntil(
		(async () => {
			const names = await caches.keys();
			await Promise.all(
				names.map((name) => {
					if (name !== CACHE_NAME) {
						return caches.delete(name);
					}
				}),
			);
			await clients.claim();
		})(),
	);
});

self.addEventListener("fetch", (event) => {
	if (event.request.mode === "navigate") {
		event.respondWith(caches.match("/"));
		return;
	}

	event.respondWith(
		(async () => {
			const cache = await caches.open(CACHE_NAME);
			const cachedResponse = await cache.match(event.request.url);
			if (cachedResponse) {
				return cachedResponse;
			}
			return new Response(null, { status: 404 });
		})(),
	);
});