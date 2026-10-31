// Keeps a share page up to date without reloading it. Over a websocket (/share/.../live) the server says
// when what the page shows changed; the page then fetches itself and swaps in the parts marked
// data-live whose content differs. Species found since are lit up for a moment. A lost connection is
// made again, and the page looks again every ten minutes in any case.
(function () {
	if (!document.querySelector("[data-live]") || !window.fetch) {
		return;
	}
	var timer = 0;
	var delay = 1000;

	function found() {
		var ids = {};
		document.querySelectorAll("[data-live] .seen[data-species], [data-live] .recent[data-species]").forEach(function (el) {
			ids[el.dataset.species] = true;
		});
		return ids;
	}

	function refresh() {
		fetch(location.href, { cache: "no-store" }).then(function (r) {
			// No longer shared: show what the server says instead.
			if (r.status === 404) {
				location.reload();
				return null;
			}
			return r.ok ? r.text() : null;
		}).then(function (html) {
			if (!html) {
				return;
			}
			var next = new DOMParser().parseFromString(html, "text/html");
			var before = found();
			// What changes appears at once; the entrance animations were for the first look.
			document.documentElement.classList.add("live-updated");
			document.querySelectorAll("[data-live]").forEach(function (part) {
				var fresh = next.querySelector('[data-live="' + part.dataset.live + '"]');
				if (fresh && fresh.innerHTML !== part.innerHTML) {
					part.innerHTML = fresh.innerHTML;
				}
			});
			document.querySelectorAll("[data-live] .seen[data-species], [data-live] .recent[data-species]").forEach(function (el) {
				if (!before[el.dataset.species]) {
					el.classList.add("just");
				}
			});
			document.dispatchEvent(new Event("share:updated"));
		}).catch(function () { });
	}

	// A burst of changes (a list of sightings saved at once) is fetched once.
	function soon() {
		clearTimeout(timer);
		timer = setTimeout(refresh, 1500);
	}

	function connect() {
		if (!window.WebSocket) {
			return;
		}
		var socket = new WebSocket((location.protocol === "https:" ? "wss://" : "ws://") + location.host + location.pathname.replace(/\/$/, "") + "/live");
		socket.onopen = function () {
			delay = 1000;
		};
		socket.onmessage = soon;
		socket.onclose = function () {
			setTimeout(function () {
				connect();
				// Whatever changed while away.
				soon();
			}, delay);
			delay = Math.min(delay * 2, 60000);
		};
	}

	connect();
	setInterval(refresh, 10 * 60 * 1000);
})();
