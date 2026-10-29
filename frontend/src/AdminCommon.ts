/// <reference path="ui/Sheet.ts"/>
/// <reference path="ui/Input.ts"/>
/// <reference path="App.ts"/>

// Building blocks shared by the admin sheets.
namespace Admin {
	export function list(parent: HTMLElement): HTMLDivElement {
		const l = document.createElement("div");
		l.className = "manage-list";
		parent.appendChild(l);
		return l;
	}

	// A row with a label, a quieter detail on the right and, when it opens something, a chevron.
	export function row(list: HTMLElement, label: string, meta: string, open: (() => void) | null): HTMLDivElement {
		const r = document.createElement("div");
		r.className = "manage-row";
		const l = document.createElement("p");
		l.textContent = label;
		r.appendChild(l);
		if (meta !== "") {
			const m = document.createElement("p");
			m.className = "meta";
			m.textContent = meta;
			r.appendChild(m);
		}
		if (open) {
			r.setAttribute("role", "button");
			r.tabIndex = 0;
			const chevron = document.createElement("span");
			chevron.className = "material-symbols-outlined";
			chevron.textContent = "chevron_right";
			r.appendChild(chevron);
			r.addEventListener("click", open);
			r.addEventListener("keydown", function (e) {
				if (e.key === "Enter") {
					open();
				}
			});
		} else {
			r.style.cursor = "default";
		}
		list.appendChild(r);
		return r;
	}

	// A note under a list, such as how much of it is shown.
	export function note(parent: HTMLElement): HTMLParagraphElement {
		const p = document.createElement("p");
		p.className = "details";
		parent.appendChild(p);
		return p;
	}

	export function input(sheet: Ui.Sheet, parent: HTMLElement, id: string, label: string, type: string = "text"): Ui.Input {
		const i = new Ui.Input(type);
		i.container.id = id;
		sheet.field(parent, label, i.container);
		return i;
	}

	export function showError(error: any): void {
		new Ui.MessageBox(App.translations.Error, error, "/img/messagebox-256.webp", "var(--error-color)", function () { });
	}

	const languages: Array<[keyof Models.Translation, () => string]> = [
		["en", () => App.translations.English],
		["sv", () => App.translations.Swedish],
		["de", () => App.translations.German],
		["la", () => App.translations.Latin],
	];

	// One input per language for a name. Latin is only for species.
	export class NameFields {
		private mInputs = new Map<string, Ui.Input>();

		constructor(sheet: Ui.Sheet, parent: HTMLElement, idPrefix: string, withLatin: boolean, initial: Models.Translation | null, changed: () => void = function () { }) {
			languages.forEach(([code, label]) => {
				if (code === "la" && !withLatin) {
					return;
				}
				const i = input(sheet, parent, idPrefix + "-" + code, label());
				if (initial) {
					i.value = (initial[code] as string) || "";
				}
				i.container.addEventListener("input", changed);
				this.mInputs.set(code as string, i);
			});
		}

		public get value(): any {
			const v: any = {};
			this.mInputs.forEach((i, code) => v[code] = i.value.trim());
			return v;
		}

		public get english(): string {
			return this.mInputs.get("en")!.value.trim();
		}

		public setDisabled(disabled: boolean): void {
			this.mInputs.forEach(i => i.setDisabled(disabled));
		}

		public clear(): void {
			this.mInputs.forEach(i => i.value = "");
		}
	}

	// Edits a name in place, in every language.
	export function editName(name: Models.Translation, withLatin: boolean, saved: () => void): void {
		const sheet = new Ui.Sheet({ title: App.translations.EditName, subtitle: name.en, size: "auto" });
		const fields = new NameFields(sheet, sheet.section(), "edit-name", withLatin, name);
		const save = sheet.addAction(Ui.actionButton(App.translations.Save, "check", "success"));
		save.addEventListener("click", function () {
			save.disabled = true;
			fields.setDisabled(true);
			const d = fields.value;
			d.uuid = name.uuid;
			new HttpRequest("POST", "/v1/translations/update", d).process(function () {
				Object.assign(name, fields.value);
				new Ui.Toast(App.translations.Saved, "", null, "var(--success-color)", function () { });
				sheet.close();
				saved();
			}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
				showError(error);
			}, function () {
				save.disabled = false;
				fields.setDisabled(false);
			});
		});
		sheet.open();
	}

	// Shows the first rows that match a search, so long lists stay quick to draw.
	export function filtered<T>(items: Array<T>, query: string, text: (item: T) => string, limit: number = 100): Array<T> {
		const q = query.trim().toLowerCase();
		const out: Array<T> = [];
		for (const item of items) {
			if (q === "" || text(item).toLowerCase().indexOf(q) > -1) {
				out.push(item);
				if (out.length === limit) {
					break;
				}
			}
		}
		return out;
	}

	export function search(sheet: Ui.Sheet, parent: HTMLElement, id: string, changed: (query: string) => void): Ui.Input {
		const i = new Ui.Input("search");
		i.container.id = id;
		i.setPlaceHolder(App.translations.Search + "...");
		i.style.marginBottom = "12px";
		parent.appendChild(i.container);
		i.container.addEventListener("input", function () {
			changed(i.value);
		});
		return i;
	}
}
