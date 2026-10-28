/// <reference path="ui/PopUp.ts"/>

// Full screen viewer for the Wikipedia photos of a species or order. Swipe, the arrow keys or the
// side buttons go between photos; pinch, double-tap or the mouse wheel zoom and a drag pans a zoomed
// photo; a swipe down, a tap beside the photo, Escape or back close it. Each photo is credited with
// its author and license, as the photos' licenses require.
class ImageViewer extends Ui.PopUp {
	private static readonly MAX_ZOOM = 4;
	private static sCredits: { [url: string]: Promise<{ artist: string, license: string, page: string } | null> } = {};

	private mUrls: Array<string>;
	private mIndex: number;
	private mStage: HTMLDivElement;
	private mImg: HTMLImageElement;
	private mCounter: HTMLParagraphElement;
	private mCredit: HTMLParagraphElement;
	private mPrev: HTMLButtonElement;
	private mNext: HTMLButtonElement;
	private mArrowListener: (event: KeyboardEvent) => void;

	// Zoom and pan of the photo, and the swipe in progress when it is not zoomed.
	private mScale: number;
	private mX: number;
	private mY: number;
	private mSwipeX: number;
	private mSwipeY: number;

	constructor(urls: Array<string>, index: number) {
		const stage = document.createElement("div");
		super(stage, () => { });

		const self = this;
		this.mUrls = urls;
		this.mIndex = index;
		this.mStage = stage;
		this.mScale = 1;
		this.mX = 0;
		this.mY = 0;
		this.mSwipeX = 0;
		this.mSwipeY = 0;

		this.addClass("image-viewer");
		stage.className = "viewer-stage is-loading";

		this.mImg = document.createElement("img");
		this.mImg.alt = "";
		this.mImg.draggable = false;
		this.mImg.addEventListener("load", function () {
			stage.classList.remove("is-loading", "is-broken");
		});
		this.mImg.addEventListener("error", function () {
			// A thumbnail wider than the photo itself does not exist; show the original then.
			const original = self.mUrls[self.mIndex];
			if (self.mImg.src !== original) {
				self.mImg.src = original;
			} else {
				stage.classList.remove("is-loading");
				stage.classList.add("is-broken");
			}
		});
		stage.appendChild(this.mImg);

		const broken = document.createElement("span");
		broken.className = "material-symbols-outlined viewer-broken";
		broken.textContent = "broken_image";
		stage.appendChild(broken);

		this.mCounter = document.createElement("p");
		this.mCounter.className = "viewer-counter";
		stage.appendChild(this.mCounter);

		this.mPrev = this.navButton("chevron_left", App.translations.PreviousPhoto, -1);
		this.mNext = this.navButton("chevron_right", App.translations.NextPhoto, 1);

		this.mCredit = document.createElement("p");
		this.mCredit.className = "viewer-credit";
		stage.appendChild(this.mCredit);

		this.mArrowListener = function (event: KeyboardEvent) {
			if (event.key === "ArrowLeft") {
				self.go(-1);
			} else if (event.key === "ArrowRight") {
				self.go(1);
			}
		};
		document.addEventListener("keydown", this.mArrowListener);

		this.handleGestures();
		this.show();
	}

	public close(fromHistory: boolean = false): void {
		document.removeEventListener("keydown", this.mArrowListener);
		super.close(fromHistory);
	}

	private navButton(icon: string, label: string, step: number): HTMLButtonElement {
		const self = this;
		const button = document.createElement("button");
		button.className = "viewer-nav " + (step < 0 ? "prev" : "next");
		button.setAttribute("aria-label", label);
		button.innerHTML = "<span class='material-symbols-outlined'>" + icon + "</span>";
		button.addEventListener("click", function () {
			self.go(step);
		});
		this.mStage.appendChild(button);
		return button;
	}

	private static sized(url: string): string {
		// Wikimedia only serves thumbnails in its standard widths.
		const width = window.innerWidth * (window.devicePixelRatio || 1) > 1280 ? "1920px" : "1280px";
		return getThumbnailFromUrl(url, width);
	}

	private show(): void {
		const url = this.mUrls[this.mIndex];
		this.mStage.classList.remove("is-broken");
		if (!this.mImg.complete || this.mImg.src !== ImageViewer.sized(url)) {
			this.mStage.classList.add("is-loading");
		}
		this.mImg.src = ImageViewer.sized(url);

		const many = this.mUrls.length > 1;
		this.mCounter.textContent = many ? (this.mIndex + 1) + " / " + this.mUrls.length : "";
		this.mPrev.hidden = !many;
		this.mNext.hidden = !many;
		this.mPrev.disabled = this.mIndex === 0;
		this.mNext.disabled = this.mIndex === this.mUrls.length - 1;

		// The neighbours, so going to them shows them at once.
		[this.mIndex - 1, this.mIndex + 1].forEach(i => {
			if (i >= 0 && i < this.mUrls.length) {
				new Image().src = ImageViewer.sized(this.mUrls[i]);
			}
		});

		this.showCredit(url);
	}

	private showCredit(url: string): void {
		const self = this;
		this.mCredit.textContent = "";
		ImageViewer.credit(url).then(function (credit) {
			if (!credit || self.mUrls[self.mIndex] !== url) {
				return;
			}
			self.mCredit.textContent = App.translations.PhotoBy + " " + credit.artist + (credit.license ? " · " + credit.license + " · " : " · ");
			const link = document.createElement("a");
			link.href = credit.page;
			link.target = "_blank";
			link.rel = "noopener";
			link.textContent = "Wikimedia";
			self.mCredit.appendChild(link);
		});
	}

	// The author and license of a photo on Wikimedia Commons or a Wikipedia, or null when unknown.
	private static credit(url: string): Promise<{ artist: string, license: string, page: string } | null> {
		if (!ImageViewer.sCredits[url]) {
			const m = url.split(/[?#]/)[0].match(/\/wikipedia\/([a-z-]+)\/(?:thumb\/)?[0-9a-f]\/[0-9a-f]{2}\/([^/]+)/);
			if (!m) {
				return Promise.resolve(null);
			}
			const host = m[1] === "commons" ? "commons.wikimedia.org" : m[1] + ".wikipedia.org";
			const title = "File:" + decodeURIComponent(m[2]);
			const api = "https://" + host + "/w/api.php?action=query&format=json&origin=*&prop=imageinfo&iiprop=extmetadata|url&iiextmetadatafilter=Artist|LicenseShortName&titles=" + encodeURIComponent(title);
			const text = function (html: string): string {
				return (new DOMParser().parseFromString(html || "", "text/html").body.textContent || "").trim();
			};
			ImageViewer.sCredits[url] = fetch(api).then(r => r.json()).then(function (data) {
				const pages = data.query && data.query.pages ? Object.keys(data.query.pages).map(k => data.query.pages[k]) : [];
				const info = pages.length > 0 && pages[0].imageinfo ? pages[0].imageinfo[0] : null;
				if (!info || !info.extmetadata || !/^https:\/\//.test(info.descriptionurl || "")) {
					return null;
				}
				const meta = info.extmetadata;
				const artist = text(meta.Artist ? meta.Artist.value : "");
				if (artist === "") {
					return null;
				}
				return { artist: artist, license: text(meta.LicenseShortName ? meta.LicenseShortName.value : ""), page: info.descriptionurl };
			}).catch(() => null);
		}
		return ImageViewer.sCredits[url];
	}

	private go(step: number): void {
		const index = this.mIndex + step;
		if (index < 0 || index >= this.mUrls.length) {
			this.settle();
			return;
		}
		const self = this;
		const img = this.mImg;
		// Out to the side it goes towards, then the next photo in from the other side.
		img.style.transition = "transform 160ms var(--ease-in), opacity 160ms";
		img.style.transform = "translateX(" + (-step * 40) + "%)";
		img.style.opacity = "0";
		setTimeout(function () {
			self.mIndex = index;
			self.mScale = 1;
			self.mX = 0;
			self.mY = 0;
			self.mSwipeX = 0;
			img.style.transition = "none";
			img.style.transform = "translateX(" + (step * 12) + "%)";
			self.show();
			requestAnimationFrame(function () {
				requestAnimationFrame(function () {
					img.style.transition = "";
					img.style.opacity = "";
					self.apply();
				});
			});
		}, 160);
	}

	// Puts the photo where the zoom, pan and swipe say.
	private apply(): void {
		const x = this.mX + this.mSwipeX;
		const y = this.mY + this.mSwipeY;
		this.mImg.style.transform = "translate(" + x + "px, " + y + "px) scale(" + this.mScale + ")";
		this.mStage.classList.toggle("zoomed", this.mScale > 1);
		// Dragging down fades the dark background to show it will close.
		this.mStage.style.setProperty("--dim", String(1 - Math.min(1, Math.abs(this.mSwipeY) / 400)));
	}

	// Back to a valid position after a gesture, animated.
	private settle(): void {
		this.mImg.style.transition = "";
		this.mSwipeX = 0;
		this.mSwipeY = 0;
		if (this.mScale <= 1) {
			this.mScale = 1;
			this.mX = 0;
			this.mY = 0;
		} else {
			const maxX = Math.max(0, (this.mImg.offsetWidth * this.mScale - this.mStage.clientWidth) / 2);
			const maxY = Math.max(0, (this.mImg.offsetHeight * this.mScale - this.mStage.clientHeight) / 2);
			this.mX = Math.max(-maxX, Math.min(maxX, this.mX));
			this.mY = Math.max(-maxY, Math.min(maxY, this.mY));
		}
		this.apply();
	}

	// Zooms to the scale keeping the screen point (x, y) where it is.
	private zoomAt(scale: number, x: number, y: number): void {
		scale = Math.max(1, Math.min(ImageViewer.MAX_ZOOM, scale));
		const box = this.mStage.getBoundingClientRect();
		const cx = x - (box.left + box.width / 2);
		const cy = y - (box.top + box.height / 2);
		const ratio = scale / this.mScale;
		this.mX = cx - ratio * (cx - this.mX);
		this.mY = cy - ratio * (cy - this.mY);
		this.mScale = scale;
	}

	private handleGestures(): void {
		const self = this;
		const stage = this.mStage;
		const img = this.mImg;
		const pointers: { [id: number]: { x: number, y: number } } = {};
		let start = { x: 0, y: 0, time: 0 };
		let moved = false;
		let axis: "" | "x" | "y" = "";
		let pinch = { distance: 0, scale: 1 };
		let lastTap = { x: 0, y: 0, time: 0 };
		let tapTimer: number | null = null;

		const ids = () => Object.keys(pointers).map(Number);
		const distance = function () {
			const [a, b] = ids().map(id => pointers[id]);
			return Math.hypot(a.x - b.x, a.y - b.y);
		};

		stage.addEventListener("pointerdown", function (event) {
			if ((event.target as HTMLElement).closest("button, a")) {
				return;
			}
			try {
				stage.setPointerCapture(event.pointerId);
			} catch (e) {
				// The pointer is already gone.
			}
			pointers[event.pointerId] = { x: event.clientX, y: event.clientY };
			img.style.transition = "none";
			if (ids().length === 1) {
				start = { x: event.clientX, y: event.clientY, time: Date.now() };
				moved = false;
				axis = "";
			} else if (ids().length === 2) {
				pinch = { distance: distance(), scale: self.mScale };
				self.mSwipeX = 0;
				self.mSwipeY = 0;
				moved = true;
			}
		});

		stage.addEventListener("pointermove", function (event) {
			const p = pointers[event.pointerId];
			if (!p) {
				return;
			}
			const dx = event.clientX - p.x;
			const dy = event.clientY - p.y;
			p.x = event.clientX;
			p.y = event.clientY;

			if (ids().length === 2) {
				const [a, b] = ids().map(id => pointers[id]);
				self.zoomAt(pinch.scale * distance() / pinch.distance, (a.x + b.x) / 2, (a.y + b.y) / 2);
				self.apply();
				return;
			}
			if (!moved && Math.hypot(event.clientX - start.x, event.clientY - start.y) < 8) {
				return;
			}
			moved = true;
			if (self.mScale > 1) {
				self.mX += dx;
				self.mY += dy;
			} else {
				if (axis === "") {
					axis = Math.abs(event.clientX - start.x) > Math.abs(event.clientY - start.y) ? "x" : "y";
				}
				if (axis === "x") {
					// Resist past the first and last photo.
					const atEnd = (self.mSwipeX > 0 && self.mIndex === 0) || (self.mSwipeX < 0 && self.mIndex === self.mUrls.length - 1);
					self.mSwipeX += atEnd ? dx / 3 : dx;
				} else {
					self.mSwipeY += dy;
				}
			}
			self.apply();
		});

		const end = function (event: PointerEvent) {
			if (!pointers[event.pointerId]) {
				return;
			}
			delete pointers[event.pointerId];
			if (ids().length > 0) {
				// One finger of a pinch lifted: the other one goes on panning from here.
				const rest = pointers[ids()[0]];
				start = { x: rest.x, y: rest.y, time: Date.now() };
				return;
			}

			if (!moved && event.type === "pointerup") {
				const now = Date.now();
				if (now - lastTap.time < 300 && Math.hypot(event.clientX - lastTap.x, event.clientY - lastTap.y) < 30) {
					// Double tap: zoom in there, or back out.
					if (tapTimer !== null) {
						clearTimeout(tapTimer);
						tapTimer = null;
					}
					lastTap.time = 0;
					img.style.transition = "";
					if (self.mScale > 1) {
						self.mScale = 1;
					} else {
						self.zoomAt(2.5, event.clientX, event.clientY);
					}
					self.settle();
					return;
				}
				lastTap = { x: event.clientX, y: event.clientY, time: now };
				// The stage captures the pointer, so the target is always the stage.
				const onPhoto = document.elementFromPoint(event.clientX, event.clientY) === img;
				// A single tap beside the photo closes the viewer, once it is clear it is no double tap.
				tapTimer = window.setTimeout(function () {
					tapTimer = null;
					if (!onPhoto && self.mScale === 1) {
						self.close();
					}
				}, 300);
				return;
			}

			const elapsed = Math.max(1, Date.now() - start.time);
			if (axis === "x" && self.mScale === 1) {
				const fast = Math.abs(self.mSwipeX) / elapsed > 0.5;
				if (Math.abs(self.mSwipeX) > stage.clientWidth / 4 || (fast && Math.abs(self.mSwipeX) > 30)) {
					self.go(self.mSwipeX < 0 ? 1 : -1);
					return;
				}
			} else if (axis === "y" && self.mScale === 1) {
				if (Math.abs(self.mSwipeY) > 120 || (Math.abs(self.mSwipeY) / elapsed > 0.5 && Math.abs(self.mSwipeY) > 40)) {
					self.close();
					return;
				}
			}
			self.settle();
		};
		stage.addEventListener("pointerup", end);
		stage.addEventListener("pointercancel", end);

		stage.addEventListener("wheel", function (event) {
			event.preventDefault();
			img.style.transition = "none";
			self.zoomAt(self.mScale * Math.exp(-event.deltaY / 300), event.clientX, event.clientY);
			self.settle();
			img.style.transition = "none";
		}, { passive: false });
	}
}

function openImageViewer(urls: Array<string>, index: number): void {
	new ImageViewer(urls, index);
}
