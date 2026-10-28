/// <reference path="Widget.ts"/>

namespace Ui {
	// A plain search field for species. Results appear under it as you type: species on your own lists
	// first, then Swedish species, then the rest; within each, matches on the start of the name come
	// before matches anywhere in it, and Latin names match too.
	export class SpeciesSearch extends Widget {
		private mInput: HTMLInputElement;
		private mResults: HTMLDivElement;
		private mSpecies: Array<Models.Species> = [];
		private mPreferred = new Set<string>();
		private mSwedish = new Set<string>();
		private mActive = -1;
		private mPick: (species: Models.Species) => void;
		private mInline: boolean;
		private mHint: string;

		// Inline: results stay in the flow below the field (in a sheet) instead of a dropdown, and the
		// hint shows while the field is empty.
		// With identify, a microphone button in the field starts identification by sound.
		constructor(placeholder: string, pick: (species: Models.Species) => void, inline: boolean = false, hint: string = "", identify: (() => void) | null = null) {
			super();
			const self = this;
			this.mPick = pick;
			this.mInline = inline;
			this.mHint = hint;

			this.mContainer = document.createElement("div");
			this.mContainer.className = inline ? "species-search inline" : "species-search";

			const icon = document.createElement("span");
			icon.className = "material-symbols-outlined";
			icon.textContent = "search";
			this.mContainer.appendChild(icon);

			this.mInput = document.createElement("input");
			this.mInput.type = "search";
			this.mInput.placeholder = placeholder;
			this.mInput.setAttribute("aria-label", placeholder);
			this.mInput.setAttribute("autocomplete", "off");
			this.mInput.setAttribute("role", "combobox");
			this.mInput.setAttribute("aria-expanded", "false");
			this.mContainer.appendChild(this.mInput);

			if (identify) {
				const mic = document.createElement("button");
				mic.className = "icon-button species-search-mic";
				mic.setAttribute("aria-label", App.translations.IdentifyBySound);
				mic.title = App.translations.IdentifyBySound;
				mic.innerHTML = "<span class='material-symbols-outlined'>mic</span>";
				mic.addEventListener("click", function () {
					identify();
				});
				this.mContainer.appendChild(mic);
				this.mContainer.classList.add("with-mic");
			}

			this.mResults = document.createElement("div");
			this.mResults.className = "species-results hidden";
			this.mResults.setAttribute("role", "listbox");
			this.mContainer.appendChild(this.mResults);

			this.mInput.addEventListener("input", function () {
				self.update();
			});
			this.mInput.addEventListener("focus", function () {
				self.update();
			});
			this.mInput.addEventListener("keydown", function (e) {
				const items = self.mResults.querySelectorAll("button");
				if (e.key === "ArrowDown" || e.key === "ArrowUp") {
					e.preventDefault();
					if (items.length === 0) {
						return;
					}
					self.mActive = (self.mActive + (e.key === "ArrowDown" ? 1 : -1) + items.length) % items.length;
					items.forEach((b, i) => b.classList.toggle("active", i === self.mActive));
					(items[self.mActive] as HTMLElement).scrollIntoView({ block: "nearest" });
				} else if (e.key === "Enter") {
					const item = items[Math.max(0, self.mActive)] as HTMLButtonElement | undefined;
					if (item) {
						e.preventDefault();
						item.click();
					}
				} else if (e.key === "Escape") {
					e.stopPropagation();
					self.hide();
				}
			});
			if (!inline) {
				document.addEventListener("pointerdown", function (e) {
					if (!self.mContainer.contains(e.target as Node)) {
						self.hide();
					}
				});
			}
			this.update();
		}

		public setSpecies(species: Array<Models.Species>, preferred: Set<string>, swedish: Set<string> = this.mSwedish): void {
			this.mSpecies = species;
			this.mPreferred = preferred;
			this.mSwedish = swedish;
			if (document.activeElement === this.mInput) {
				this.update();
			}
		}

		public focus(): void {
			this.mInput.focus();
		}

		private hide(): void {
			if (this.mInline) {
				return;
			}
			this.mResults.classList.add("hidden");
			this.mInput.setAttribute("aria-expanded", "false");
		}

		private update(): void {
			const self = this;
			const query = this.mInput.value.trim().toLowerCase();
			this.mResults.innerHTML = "";
			this.mActive = -1;
			if (query === "") {
				if (this.mInline && this.mHint !== "") {
					const hint = document.createElement("p");
					hint.className = "species-results-empty";
					hint.textContent = this.mHint;
					this.mResults.appendChild(hint);
					this.mResults.classList.remove("hidden");
				}
				this.hide();
				return;
			}

			const scored = Array<{ species: Models.Species, score: number }>();
			this.mSpecies.forEach(s => {
				const name = App.getTranslation(s.name).toLowerCase();
				const latin = (s.name.la || "").toLowerCase();
				let score = -1;
				if (name.startsWith(query)) {
					score = 0;
				} else if (name.split(/[\s-]/).some(w => w.startsWith(query))) {
					score = 1;
				} else if (name.indexOf(query) >= 0 || latin.startsWith(query)) {
					score = 2;
				} else if (latin.indexOf(query) >= 0) {
					score = 3;
				}
				if (score >= 0) {
					const tier = this.mPreferred.has(s.uuid) ? 0 : this.mSwedish.has(s.uuid) ? 10 : 20;
					scored.push({ species: s, score: score + tier });
				}
			});
			scored.sort((a, b) => a.score - b.score || App.getTranslation(a.species.name).localeCompare(App.getTranslation(b.species.name)));

			const shown = scored.slice(0, this.mInline ? 15 : 8);
			if (shown.length === 0) {
				const none = document.createElement("p");
				none.className = "species-results-empty";
				none.textContent = App.translations.NoResults;
				this.mResults.appendChild(none);
			}
			shown.forEach(r => {
				const b = document.createElement("button");
				b.setAttribute("role", "option");
				const name = document.createElement("span");
				name.className = "result-name";
				name.textContent = App.getTranslation(r.species.name);
				b.appendChild(name);
				const latin = document.createElement("span");
				latin.className = "result-latin";
				latin.textContent = r.species.name.la;
				b.appendChild(latin);
				b.addEventListener("click", function () {
					self.hide();
					self.mInput.blur();
					self.mPick(r.species);
				});
				self.mResults.appendChild(b);
			});
			this.mResults.classList.remove("hidden");
			this.mInput.setAttribute("aria-expanded", "true");
		}
	}
}
