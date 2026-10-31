// The species card (templates/species-card.html), for the bird search on the landing page and the
// species on the share pages. Reporting a species keeps it and the time in the browser and goes on to
// an account; the app opens the sighting once signed in (PendingReport in the app).
window.SpeciesCard = (function () {
	var playIcon = "M8 5v14l11-7z";
	var pauseIcon = "M6 19h4V5H6v14Zm8-14v14h4V5h-4Z";

	function name(t, lang) {
		return (lang === "sv" ? t.sv : lang === "de" ? t.de : t.en) || t.en || t.sv || t.la || "";
	}

	return function (card, lang) {
		var photo = card.querySelector(".species-card-photo img");
		var failed = card.querySelector(".species-card-failed");
		var listen = card.querySelector("[data-listen]");
		var wiki = card.querySelector(".species-card-wiki");
		var chosen = null;
		var audio = null;

		function stop() {
			if (audio) {
				audio.pause();
			}
			listen.classList.remove("playing");
			listen.querySelector("path").setAttribute("d", playIcon);
			listen.querySelector("span").textContent = listen.dataset.play;
		}

		listen.addEventListener("click", function () {
			if (!audio) {
				return;
			}
			if (listen.classList.contains("playing")) {
				stop();
				return;
			}
			audio.play().then(function () {
				listen.classList.add("playing");
				listen.querySelector("path").setAttribute("d", pauseIcon);
				listen.querySelector("span").textContent = listen.dataset.pause;
			}).catch(stop);
		});

		// Kept for a day: the species and when it was seen, which is now.
		card.querySelector("[data-report]").addEventListener("click", function () {
			if (!chosen) {
				return;
			}
			try {
				localStorage.setItem("pending-report", JSON.stringify({ uuid: chosen.uuid, at: new Date().toISOString() }));
			} catch (e) {
			}
			location.href = /(?:^|;\s*)sf_session=/.test(document.cookie) ? "/home" : "/sign-up";
		});

		return {
			name: function (t) {
				return name(t, lang);
			},
			stop: stop,
			// species needs uuid; its name (or label) shows while the rest loads.
			show: function (species, label) {
				chosen = species;
				stop();
				audio = null;
				var n = species.name || {};
				card.querySelector(".species-card-name").textContent = name(n, lang) || label || "";
				card.querySelector(".species-card-latin").textContent = n.la || "";
				card.querySelector(".species-card-order").textContent = "";
				card.querySelector(".species-card-summary").textContent = "";
				failed.hidden = true;
				photo.hidden = true;
				listen.hidden = true;
				wiki.hidden = true;
				card.hidden = false;
				card.classList.add("loading");
				fetch("/api/v1/species/" + encodeURIComponent(species.uuid), { headers: { "X-App-Language": lang } }).then(function (r) {
					if (!r.ok) {
						throw new Error(String(r.status));
					}
					return r.json();
				}).then(function (d) {
					if (chosen !== species) {
						return;
					}
					card.classList.remove("loading");
					if (d.result) {
						card.querySelector(".species-card-name").textContent = name(d.result.name, lang);
						card.querySelector(".species-card-latin").textContent = d.result.name.la || "";
						if (d.result.order) {
							card.querySelector(".species-card-order").textContent = name(d.result.order.name, lang);
						}
					}
					var w = d.wikipedia || {};
					card.querySelector(".species-card-summary").textContent = w.summary || "";
					if (w.thumbnail) {
						photo.src = w.thumbnail;
						photo.hidden = false;
					}
					if (w.url) {
						wiki.href = w.url;
						wiki.hidden = false;
					}
					var song = (d.recordings || []).filter(function (r) { return r.url; })[0];
					if (song) {
						audio = new Audio(song.url);
						audio.addEventListener("ended", stop);
						listen.hidden = false;
					}
				}).catch(function () {
					if (chosen === species) {
						card.classList.remove("loading");
						failed.hidden = false;
					}
				});
			}
		};
	};
})();

// On a page with a species dialog (#species-dialog), anything with data-species="<uuid>" opens it.
(function () {
	var dialog = document.getElementById("species-dialog");
	if (!dialog || !dialog.showModal) {
		return;
	}
	var card = window.SpeciesCard(dialog.querySelector(".species-card"), document.documentElement.lang);
	document.addEventListener("click", function (e) {
		var el = e.target.closest ? e.target.closest("[data-species]") : null;
		if (!el) {
			return;
		}
		e.preventDefault();
		card.show({ uuid: el.dataset.species, name: {} }, el.dataset.name);
		dialog.showModal();
	});
	document.addEventListener("keydown", function (e) {
		if ((e.key === "Enter" || e.key === " ") && e.target.dataset && e.target.dataset.species) {
			e.preventDefault();
			e.target.click();
		}
	});
	dialog.addEventListener("click", function (e) {
		// A click on the backdrop, or the close button.
		if (e.target === dialog || (e.target.closest && e.target.closest("[data-close]"))) {
			dialog.close();
		}
	});
	dialog.addEventListener("close", card.stop);
})();
