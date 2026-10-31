/// <reference path="ui/Div.ts"/>
/// <reference path="ui/PopUp.ts"/>
/// <reference path="ui/Sheet.ts"/>
/// <reference path="App.ts"/>
/// <reference path="ImageViewer.ts"/>
/// <reference path="models/Recording.ts"/>
/// <reference path="RecordingPopUp.ts"/>
/// <reference path="OrderPopUp.ts"/>

// A species at a glance, opened from checklists, sightings and search. It is a sheet over the page so
// the list behind keeps its place. The sheet has a fixed height and every section reserves its space
// before the Wikipedia and xeno-canto data arrives, so nothing moves while it loads; the report button
// sits in the footer.
function openSpeciesPopUp(species: Models.Species, list: Models.List | null, closeCallback: (speciesAdded: boolean) => void) {
	let speciesAdded = false;

	const sheet = new Ui.Sheet({ title: App.getTranslation(species.name), subtitle: species.name.la, latin: true, hero: true });

	if (species.order) {
		const order = document.createElement("button");
		order.className = "sheet-chip";
		order.textContent = App.getTranslation(species.order.name);
		order.addEventListener("click", function () {
			openOrderPopUp(species.order);
		});
		sheet.names.insertBefore(order, sheet.title);
	}

	// About: placeholder lines until the summary arrives.
	const about = sheet.section(App.translations.AboutSpecies);
	const summary = document.createElement("p");
	summary.className = "species-summary";
	summary.innerHTML = "<span class='skeleton line'></span><span class='skeleton line'></span><span class='skeleton line'></span><span class='skeleton line short'></span>";
	about.appendChild(summary);

	const toggle = document.createElement("button");
	toggle.className = "text-button hidden";
	toggle.textContent = App.translations.ShowMore;
	toggle.addEventListener("click", function () {
		const clamped = summary.classList.toggle("clamped");
		toggle.textContent = clamped ? App.translations.ShowMore : App.translations.ShowLess;
	});
	about.appendChild(toggle);

	const readMore = document.createElement("a");
	readMore.className = "sheet-link hidden";
	readMore.target = "_blank";
	readMore.rel = "noopener";
	readMore.innerHTML = App.translations.ReadOnWikipedia + " <span class='material-symbols-outlined'>open_in_new</span>";
	about.appendChild(readMore);

	// Photos: placeholders of the same size until the real ones arrive.
	const photos = sheet.section(App.translations.Photos);
	const strip = photoStrip(photos);

	// Sounds from xeno-canto; hidden when there are none.
	const sounds = sheet.section(App.translations.Sounds);
	sounds.classList.add("hidden");
	const soundList = document.createElement("div");
	soundList.className = "sound-list";
	sounds.appendChild(soundList);

	sheet.addAction(Ui.actionButton(App.translations.SubmitSighting, "add_circle")).addEventListener("click", function () {
		openReportPopUp(species, list, function (sa: boolean) {
			speciesAdded = speciesAdded || sa;
		});
	});

	const r = new HttpRequest("GET", "/v1/species/" + species.uuid);
	r.process(function (status: any, data: any) {
		const wikipedia = data.wikipedia || {};
		sheet.setHeroImage(wikipediaHeroUrl(wikipedia));

		summary.textContent = wikipedia.summary ? wikipedia.summary : App.translations.NoProvidedInfo;
		// Long articles start folded so the photos and sounds stay in view.
		summary.classList.add("clamped");
		if (summary.scrollHeight > summary.clientHeight + 2) {
			toggle.classList.remove("hidden");
		} else {
			summary.classList.remove("clamped");
		}
		const articleUrl = Utils.safeUrl(wikipedia.url);
		if (articleUrl) {
			readMore.href = articleUrl;
			readMore.classList.remove("hidden");
		}

		fillPhotoStrip(photos, strip, (wikipedia.images || []).filter(isPhotoUrl));

		const recordings = (data.recordings || []) as Array<Models.Recording>;
		recordings.forEach(e => {
			soundList.appendChild(soundCard(species, e));
		});
		if (recordings.length > 0) {
			sounds.classList.remove("hidden");
			const credit = document.createElement("p");
			credit.className = "credit";
			credit.innerHTML = "xeno-canto · <a target='_blank' rel='noopener' href='https://xeno-canto.org/'>" + App.translations.GoToWebsite + "</a>";
			sounds.appendChild(credit);
		}
	}, function () {
		sheet.setHeroImage("");
		summary.textContent = App.translations.NoProvidedInfo;
		photos.classList.add("hidden");
	});

	sheet.open(function () {
		closeCallback(speciesAdded);
	});
}

// A strip of photo placeholders in the section, replaced by fillPhotoStrip().
function photoStrip(section: HTMLElement): HTMLDivElement {
	const strip = document.createElement("div");
	strip.className = "photo-strip";
	for (let i = 0; i < 4; ++i) {
		const placeholder = document.createElement("div");
		placeholder.className = "photo skeleton";
		strip.appendChild(placeholder);
	}
	section.appendChild(strip);
	return strip;
}

function fillPhotoStrip(section: HTMLElement, strip: HTMLDivElement, photoUrls: Array<string>): void {
	strip.innerHTML = "";
	photoUrls.forEach(url => {
		const photo = document.createElement("button");
		photo.className = "photo";
		photo.setAttribute("aria-label", App.translations.Photos);
		const img = document.createElement("img");
		img.alt = "";
		img.loading = "lazy";
		img.src = getThumbnailFromUrl(url, "330px");
		img.addEventListener("load", function () {
			photo.classList.add("loaded");
		});
		img.addEventListener("error", function () {
			photo.remove();
			if (strip.childElementCount === 0) {
				section.classList.add("hidden");
			}
		});
		photo.appendChild(img);
		photo.dataset.url = url;
		photo.addEventListener("click", function () {
			// Photos that failed to load were removed from the strip; go through those left.
			const urls = Array.from(strip.querySelectorAll<HTMLElement>("button.photo")).map(b => b.dataset.url as string);
			openImageViewer(urls, urls.indexOf(url));
		});
		strip.appendChild(photo);
	});
	if (photoUrls.length === 0) {
		section.classList.add("hidden");
	}
}

// The article's lead photo in a size for a sheet header, or "" when it has none.
function wikipediaHeroUrl(wikipedia: any): string {
	if (wikipedia.thumbnail) {
		return resizeThumbnail(wikipedia.thumbnail, "960px");
	}
	const photos = (wikipedia.images || []).filter(isPhotoUrl);
	return photos.length > 0 ? getThumbnailFromUrl(photos[0], "960px") : "";
}

// The recording playing now: starting another stops it.
let playingSound: HTMLAudioElement | null = null;

// A recording in the app's own player: the browser's controls are wider than a phone's sheet and
// look different in every browser.
function soundCard(species: Models.Species, recording: Models.Recording): HTMLElement {
	const card = document.createElement("div");
	card.className = "sound-card";

	const head = document.createElement("div");
	head.className = "sound-head";
	card.appendChild(head);

	const audio = new Audio();
	audio.preload = "none";
	audio.src = recording.url;

	const play = document.createElement("button");
	play.className = "sound-play";
	play.setAttribute("aria-label", App.translations.Play);
	play.innerHTML = "<span class='material-symbols-outlined'>play_arrow</span>";
	head.appendChild(play);

	const text = document.createElement("div");
	head.appendChild(text);

	const type = document.createElement("p");
	type.className = "sound-type";
	type.textContent = recording.type.charAt(0).toUpperCase() + recording.type.slice(1);
	text.appendChild(type);

	const info = document.createElement("p");
	info.className = "details";
	info.textContent = [recording.recorder, recording.location, recording.country].filter(v => v).join(" · ");
	text.appendChild(info);

	const more = document.createElement("button");
	more.className = "icon-button";
	more.setAttribute("aria-label", App.translations.Details);
	more.innerHTML = "<span class='material-symbols-outlined'>map</span>";
	more.addEventListener("click", function () {
		audio.pause();
		openRecordingPopUp(species, recording);
	});
	head.appendChild(more);

	// Where the recording is, tapped or dragged to move in it, and the time.
	const bar = document.createElement("div");
	bar.className = "sound-bar";
	const track = document.createElement("div");
	track.className = "sound-track";
	track.setAttribute("role", "slider");
	track.setAttribute("aria-label", type.textContent);
	track.setAttribute("aria-valuemin", "0");
	track.tabIndex = 0;
	const fill = document.createElement("div");
	fill.className = "sound-fill";
	track.appendChild(fill);
	bar.appendChild(track);
	const time = document.createElement("span");
	time.className = "sound-time";
	time.textContent = "0:00";
	bar.appendChild(time);
	card.appendChild(bar);

	const clock = function (seconds: number): string {
		const s = Math.floor(seconds || 0);
		return Math.floor(s / 60) + ":" + String(s % 60).padStart(2, "0");
	};
	const show = function (): void {
		const length = isFinite(audio.duration) ? audio.duration : 0;
		fill.style.width = (length > 0 ? audio.currentTime / length * 100 : 0) + "%";
		time.textContent = length > 0 ? clock(audio.currentTime) + " / " + clock(length) : clock(audio.currentTime);
		track.setAttribute("aria-valuemax", String(Math.round(length)));
		track.setAttribute("aria-valuenow", String(Math.round(audio.currentTime)));
	};
	const playing = function (on: boolean): void {
		card.classList.toggle("playing", on);
		(play.firstChild as HTMLElement).textContent = on ? "pause" : "play_arrow";
		play.setAttribute("aria-label", on ? App.translations.Pause : App.translations.Play);
	};
	audio.addEventListener("timeupdate", show);
	audio.addEventListener("loadedmetadata", show);
	audio.addEventListener("play", function () {
		if (playingSound && playingSound !== audio) {
			playingSound.pause();
		}
		playingSound = audio;
		playing(true);
	});
	audio.addEventListener("pause", function () {
		playing(false);
	});
	audio.addEventListener("ended", function () {
		audio.currentTime = 0;
		playing(false);
		show();
	});
	audio.addEventListener("waiting", function () {
		card.classList.add("loading");
	});
	audio.addEventListener("playing", function () {
		card.classList.remove("loading");
	});

	play.addEventListener("click", function () {
		if (audio.paused) {
			audio.play().catch(function () {
				playing(false);
			});
		} else {
			audio.pause();
		}
	});

	const seek = function (clientX: number): void {
		if (!isFinite(audio.duration) || audio.duration <= 0) {
			return;
		}
		const r = track.getBoundingClientRect();
		audio.currentTime = Math.min(1, Math.max(0, (clientX - r.left) / r.width)) * audio.duration;
		show();
	};
	track.addEventListener("pointerdown", function (e) {
		seek(e.clientX);
		track.setPointerCapture(e.pointerId);
	});
	track.addEventListener("pointermove", function (e) {
		if (track.hasPointerCapture(e.pointerId)) {
			seek(e.clientX);
		}
	});
	track.addEventListener("keydown", function (e) {
		if (e.key === "ArrowLeft" || e.key === "ArrowRight") {
			e.preventDefault();
			audio.currentTime = Math.max(0, audio.currentTime + (e.key === "ArrowLeft" ? -5 : 5));
			show();
		}
	});

	// Closing the sheet takes the card away; the sound must not play on unseen.
	const stopWhenGone = new MutationObserver(function () {
		if (!card.isConnected) {
			audio.pause();
			stopWhenGone.disconnect();
		}
	});
	requestAnimationFrame(function () {
		stopWhenGone.observe(document.body, { childList: true, subtree: true });
	});

	return card;
}

// Wikimedia only serves thumbnails in standard widths (250px, 330px, 500px, ...) and now appends
// tracking parameters to image URLs, which must not end up inside the thumbnail path.
function getThumbnailFromUrl(url: string, width: string): string {
	const parts = url.split(/[?#]/)[0].split("/");
	const filename = parts.pop();
	const lastHash = parts.pop();
	const secondLastHash = parts.pop();
	const newUrl = `${parts.join('/')}/thumb/${secondLastHash}/${lastHash}/${filename}/${width}-${filename}`
	return newUrl;
}

// Changes the width of a URL that already points at a thumbnail.
function resizeThumbnail(url: string, width: string): string {
	return url.split(/[?#]/)[0].replace(/\/\d+px-([^/]+)$/, "/" + width + "-$1");
}

// Article images also include icons, logos, maps (SVG) and sound or video files; only photos are shown.
function isPhotoUrl(url: string): boolean {
	return /\.(jpe?g|png|webp)$/i.test(url.split(/[?#]/)[0]);
}
