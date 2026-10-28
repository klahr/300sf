/// <reference path="Div.ts"/>
/// <reference path="PopUp.ts"/>

namespace Ui {
	export interface SheetOptions {
		title: string;
		// Small line above the title, such as the kind of thing shown.
		eyebrow?: string;
		subtitle?: string;
		// Set the subtitle in the italic face used for Latin names.
		latin?: boolean;
		// A photo header instead of a plain one; set the photo with setHeroImage().
		hero?: boolean;
		// "tall" has a fixed height so loading content never moves the footer, "auto" fits its
		// content and "compact" is also narrow, for a field or two.
		size?: "tall" | "auto" | "compact";
	}

	// The layout every dialog shares: a header, sections with small headings, and a footer pinned to
	// the bottom for the actions.
	export class Sheet {
		public readonly root: Ui.Div;
		public readonly header: HTMLDivElement;
		public readonly names: HTMLDivElement;
		public readonly body: HTMLDivElement;
		public readonly footer: HTMLDivElement;
		private mHeroImage: HTMLImageElement | null = null;
		private mPopUp: Ui.PopUp | null = null;

		constructor(options: SheetOptions) {
			this.root = new Ui.Div();
			this.root.addClass("sheet");
			this.root.addClass("sheet-" + (options.size || "tall"));

			this.header = document.createElement("div");
			this.header.className = options.hero ? "sheet-hero is-loading" : "sheet-header";
			this.root.container.appendChild(this.header);

			if (options.hero) {
				const self = this;
				const img = document.createElement("img");
				img.alt = "";
				img.addEventListener("load", function () {
					self.header.classList.remove("is-loading");
					self.header.classList.add("loaded");
				});
				img.addEventListener("error", function () {
					self.header.classList.remove("is-loading");
				});
				this.header.appendChild(img);
				this.mHeroImage = img;
			}

			this.names = document.createElement("div");
			this.names.className = "sheet-names";
			this.header.appendChild(this.names);

			if (options.eyebrow) {
				const eyebrow = document.createElement("p");
				eyebrow.className = "sheet-eyebrow";
				eyebrow.textContent = options.eyebrow;
				this.names.appendChild(eyebrow);
			}

			const title = document.createElement("h1");
			title.textContent = options.title;
			this.names.appendChild(title);

			if (options.subtitle) {
				const subtitle = document.createElement("p");
				subtitle.className = options.latin ? "sheet-subtitle latin" : "sheet-subtitle";
				subtitle.textContent = options.subtitle;
				this.names.appendChild(subtitle);
			}

			this.body = document.createElement("div");
			this.body.className = "sheet-body";
			this.root.container.appendChild(this.body);

			this.footer = document.createElement("div");
			this.footer.className = "sheet-footer hidden";
			this.root.container.appendChild(this.footer);
		}

		public get title(): HTMLHeadingElement {
			return this.names.querySelector("h1") as HTMLHeadingElement;
		}

		// A section with an optional heading; returns the element to fill.
		public section(title: string = ""): HTMLElement {
			const s = document.createElement("section");
			s.className = "sheet-section";
			if (title !== "") {
				const h = document.createElement("h2");
				h.textContent = title;
				s.appendChild(h);
			}
			this.body.appendChild(s);
			return s;
		}

		// A labelled form field inside a section.
		public field(parent: HTMLElement, label: string, control: HTMLElement): HTMLElement {
			const f = document.createElement("div");
			f.className = "sheet-field";
			if (label !== "") {
				const l = document.createElement("label");
				l.textContent = label;
				if (control.id) {
					l.htmlFor = control.id;
				}
				f.appendChild(l);
			}
			f.appendChild(control);
			parent.appendChild(f);
			return f;
		}

		public addAction(button: HTMLButtonElement): HTMLButtonElement {
			this.footer.classList.remove("hidden");
			this.footer.appendChild(button);
			return button;
		}

		public setHeroImage(url: string): void {
			if (!this.mHeroImage) {
				return;
			}
			if (url === "") {
				this.header.classList.remove("is-loading");
				return;
			}
			this.mHeroImage.src = url;
		}

		// A sheet around content built for a plain dialog.
		public static wrap(content: HTMLElement, title: string): Sheet {
			const sheet = new Sheet({ title: title, size: "auto" });
			sheet.section().appendChild(content);
			return sheet;
		}

		public open(closeCallback: () => void = function () { }): Sheet {
			this.mPopUp = new Ui.PopUp(this.root.container, closeCallback);
			return this;
		}

		public close(): void {
			if (this.mPopUp) {
				this.mPopUp.close();
			}
		}
	}

	export function actionButton(label: string, icon: string = "", className: string = ""): HTMLButtonElement {
		const b = document.createElement("button");
		b.className = className;
		b.innerHTML = (icon !== "" ? "<span class='material-symbols-outlined'>" + icon + "</span>" : "") + label;
		return b;
	}
}
