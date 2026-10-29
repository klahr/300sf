/// <reference path="models/ListTemplate.ts"/>
/// <reference path="models/Visibility.ts"/>
/// <reference path="requests/Lists.ts"/>
/// <reference path="ui/Sheet.ts"/>
/// <reference path="ui/Select.ts"/>
/// <reference path="ui/DivSelect.ts"/>
/// <reference path="ui/SpeciesSearch.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="SpeciesCatalog.ts"/>
/// <reference path="Colors.ts"/>

// The user's own species list: which species it holds, picked one by one or copied from another list.
// New, it also gets a name and colour and becomes a checklist; opened again, its species change on
// every checklist made from it. Admins can share it with everyone.
class SpeciesListPopUp {
	private mSheet: Ui.Sheet;
	private mSpecies = new Map<string, Models.Species>();
	private mChips: HTMLDivElement;
	private mCount: HTMLHeadingElement;
	private mChanged: () => void;

	// With template, its species are edited; without, a new list is made and onCreated gets the checklist.
	constructor(template: Models.ListTemplate | null, onCreated: (listUuid: string) => void) {
		const self = this;
		const isAdmin = (new LocalStorage().get("scope") || "").includes("admin");

		const sheet = new Ui.Sheet({ eyebrow: template ? App.translations.EditSpecies : App.translations.NewChecklist, title: template ? App.getTranslation(template.name) : App.translations.OwnSpeciesList, subtitle: template ? "" : App.translations.OwnSpeciesListHint, size: "tall" });
		this.mSheet = sheet;

		const name = new Ui.Input("text");
		const colorSelect = new Ui.DivSelect();
		if (!template) {
			const about = sheet.section();
			name.setPlaceHolder(App.translations.Name);
			sheet.field(about, App.translations.Name, name.container);
			name.container.addEventListener("input", function () {
				self.mChanged();
			});

			const from = new Ui.Select(App.translations.StartEmpty, App.translations.Search, function (value: string | null) {
				self.mSpecies.clear();
				if (!value) {
					self.render();
					return;
				}
				new HttpRequest("GET", "/v1/list_templates/" + value).process(function (status: any, data: any) {
					(data.result.species || []).forEach(sp => self.mSpecies.set(sp.uuid, sp));
					self.render();
				});
			});
			from.addOption(App.translations.StartEmpty, "");
			sheet.field(about, App.translations.StartFrom, from.container);
			new HttpRequest("GET", "/v1/list_templates/list").process(function (status: any, data: any) {
				(data.result || []).forEach(t => from.addOption(App.getTranslation(t.name) + " · " + App.translations.NumberOfSpecies.replace("{0}", String(t.species.length)), t.uuid));
			});

			colorSelect.addClass("swatches");
			Colors.palette.forEach(c => colorSelect.addOption("", c).style.backgroundColor = c);
			colorSelect.select(Colors.fallback);
			sheet.field(about, App.translations.Color, colorSelect.container);
		}

		const speciesSection = sheet.section();
		this.mCount = document.createElement("h3");
		this.mCount.className = "title";
		speciesSection.appendChild(this.mCount);
		const search = new Ui.SpeciesSearch(App.translations.AddSpecies, function (species: Models.Species) {
			self.mSpecies.set(species.uuid, species);
			const input = search.container.querySelector("input") as HTMLInputElement;
			input.value = "";
			input.dispatchEvent(new Event("input"));
			self.render();
		}, true);
		speciesSection.appendChild(search.container);
		SpeciesCatalog.get(function (c) {
			search.setSpecies(c.species, c.preferred, c.swedish);
		});
		this.mChips = document.createElement("div");
		this.mChips.className = "chips";
		speciesSection.appendChild(this.mChips);

		const share = new Ui.Input("checkbox");
		if (isAdmin) {
			const row = document.createElement("div");
			row.className = "check-row";
			share.container.id = "species-list-share";
			(share.container as HTMLInputElement).checked = template !== null && template.visibility === Models.Visibility.Public;
			share.container.addEventListener("change", function () {
				self.mChanged();
			});
			row.appendChild(share.container);
			const lbl = document.createElement("label");
			lbl.setAttribute("for", share.container.id);
			lbl.textContent = App.translations.ShareWithEveryone;
			row.appendChild(lbl);
			const shareSection = sheet.section();
			shareSection.appendChild(row);
			const hint = document.createElement("p");
			hint.className = "details";
			hint.textContent = template && template.visibility === Models.Visibility.Public ? App.translations.SharedListEditWarning : App.translations.ShareWithEveryoneHint;
			shareSection.appendChild(hint);
		}
		const shared = function (): boolean {
			return (share.container as HTMLInputElement).checked;
		};

		const original = new Set<string>();
		if (template) {
			new HttpRequest("GET", "/v1/list_templates/" + template.uuid).process(function (status: any, data: any) {
				(data.result.species || []).forEach(sp => {
					self.mSpecies.set(sp.uuid, sp);
					original.add(sp.uuid);
				});
				self.render();
			});
		}

		const saveButton = sheet.addAction(Ui.actionButton(template ? App.translations.SaveChanges : App.translations.CreateChecklist, "check", "success"));
		saveButton.disabled = true;
		this.mChanged = function () {
			const speciesChanged = self.mSpecies.size !== original.size || Array.from(self.mSpecies.keys()).some(u => !original.has(u));
			const shareChanged = template !== null && shared() !== (template.visibility === Models.Visibility.Public);
			saveButton.disabled = self.mSpecies.size === 0 || (template ? !speciesChanged && !shareChanged : name.value.trim() === "");
		};

		const failed = function () {
			Ui.setBusy(saveButton, false);
			new Ui.MessageBox(App.translations.Error, App.translations.SomethingWentWrong, "/img/messagebox-256.webp", "var(--error-color)", function () { });
		};
		saveButton.addEventListener("click", function () {
			Ui.setBusy(saveButton, true);
			const species = Array.from(self.mSpecies.keys());
			if (template) {
				const requests: Array<HttpRequest> = [new HttpRequest("POST", "/v1/list_templates/" + template.uuid + "/species", { species_uuids: species })];
				if (isAdmin && shared() !== (template.visibility === Models.Visibility.Public)) {
					requests.push(new HttpRequest("POST", "/v1/list_templates/" + template.uuid + "/visibility", { visibility: shared() ? Models.Visibility.Public : Models.Visibility.Private }));
				}
				let pending = requests.length;
				requests.forEach(r => r.process(function () {
					if (--pending === 0) {
						self.mSheet.close();
						setTimeout(() => {
							location.reload();
						}, 150);
					}
				}, failed));
				return;
			}

			const n = new Models.Translation();
			n.en = n.sv = n.de = name.value.trim();
			new HttpRequest("POST", "/v1/list_templates", { name: n, description: new Models.Translation(), visibility: shared() ? Models.Visibility.Public : Models.Visibility.Private, species_uuids: species }).process(function (status: any, data: any) {
				const r = new Requests.Lists();
				r.name = n;
				r.template = data.uuid;
				r.visibility = Models.Visibility.Hidden;
				r.color = colorSelect.value || Colors.fallback;
				new HttpRequest("POST", "/v1/lists", r).process(function (status: any, list: any) {
					window.dispatchEvent(new Event("listschanged"));
					self.mSheet.close();
					onCreated(list.uuid);
				}, failed);
			}, failed);
		});

		this.render();
		sheet.open();
	}

	// The chosen species as chips, by name, each with a button to take it off.
	private render(): void {
		const self = this;
		this.mCount.textContent = App.translations.SpeciesOnList + " · " + this.mSpecies.size;
		this.mChips.innerHTML = "";
		if (this.mSpecies.size === 0) {
			const none = document.createElement("p");
			none.className = "details";
			none.textContent = App.translations.NoSpeciesYet;
			this.mChips.appendChild(none);
		}
		Array.from(this.mSpecies.values())
			.sort((a, b) => App.getTranslation(a.name).localeCompare(App.getTranslation(b.name), App.language))
			.forEach(sp => {
				const chip = document.createElement("span");
				chip.className = "chip removable";
				chip.textContent = App.getTranslation(sp.name);
				const remove = document.createElement("button");
				remove.className = "icon-button";
				remove.setAttribute("aria-label", App.translations.RemoveSpecies.replace("{0}", App.getTranslation(sp.name)));
				remove.innerHTML = "<span class='material-symbols-outlined'>close</span>";
				remove.addEventListener("click", function () {
					self.mSpecies.delete(sp.uuid);
					self.render();
				});
				chip.appendChild(remove);
				this.mChips.appendChild(chip);
			});
		this.mChanged();
	}
}

function openSpeciesListPopUp(template: Models.ListTemplate | null, onCreated: (listUuid: string) => void = function () { }): void {
	new SpeciesListPopUp(template, onCreated);
}
