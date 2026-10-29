/// <reference path="models/List.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="ui/Sheet.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="utils/Sort.ts"/>
/// <reference path="ui/DivSelect.ts"/>
/// <reference path="Colors.ts"/>
/// <reference path="SpeciesListPopUp.ts"/>

// A checklist in the management list: what it holds, renaming, recolouring and archiving it.
class EditListPopUp {
	private mSheet: Ui.Sheet;

	constructor(list: Models.List, closeCallback: () => void) {
		const self = this;

		const sheet = new Ui.Sheet({ eyebrow: App.translations.Checklist, title: App.getTranslation(list.name), subtitle: App.translations.SpeciesCount.replace("{0}", list.species.length.toString()), size: "compact" });
		this.mSheet = sheet;

		const dot = document.createElement("span");
		dot.className = "list-dot";
		dot.style.setProperty("--dot", list.color !== "" ? list.color : "var(--blue)");
		sheet.names.querySelector("p.sheet-eyebrow")?.prepend(dot);

		const form = sheet.section();

		const name = new Ui.Input("text");
		name.value = App.getTranslation(list.name);
		name.setPlaceHolder(App.translations.Name);
		sheet.field(form, App.translations.Name, name.container);

		const colorSelect = new Ui.DivSelect();
		colorSelect.addClass("swatches");
		Colors.palette.forEach(c => colorSelect.addOption("", c).style.backgroundColor = c);
		const color = list.color !== "" ? list.color : Colors.fallback;
		colorSelect.select(color);
		sheet.field(form, App.translations.Color, colorSelect.container);

		// Lists made from the user's own species list can have their species changed.
		if (list.template) {
			new HttpRequest("GET", "/v1/list_templates/list/me").process(function (status: any, data: any) {
				const template = ((data.result || []) as Array<Models.ListTemplate>).find(t => t.uuid === list.template);
				if (!template) {
					return;
				}
				const edit = Ui.actionButton(App.translations.EditSpecies, "playlist_add", "secondary");
				edit.addEventListener("click", function () {
					self.mSheet.close();
					Ui.PopUp.whenHistorySettled(function () {
						openSpeciesListPopUp(template);
					});
				});
				sheet.footer.insertBefore(edit, sheet.footer.firstChild);
			});
		}

		sheet.addAction(Ui.actionButton(App.translations.ArchiveChecklist, "archive", "danger")).addEventListener("click", function () {
			new Ui.ConfirmBox(App.translations.ArchiveChecklist, App.translations.ConfirmArchiveChecklist, "/img/messagebox-256.webp", "var(--error-color)", function () {
				const r = new HttpRequest("POST", "/v1/lists/" + list.uuid + "/archive");
				r.process(function (status: any, data: any) {
					self.mSheet.close();
					setTimeout(() => {
						location.reload();
					}, 150);
				});
			}, function () {
			}, App.translations.ArchiveChecklist, App.translations.Cancel);
		});

		const saveButton = sheet.addAction(Ui.actionButton(App.translations.SaveChanges, "check", "success"));
		const changed = function (): boolean {
			return name.value.trim() !== App.getTranslation(list.name) || (colorSelect.value !== null && colorSelect.value !== color);
		};
		saveButton.disabled = true;
		name.container.addEventListener("input", function () {
			saveButton.disabled = name.value.trim() === "" || !changed();
		});
		colorSelect.container.addEventListener("click", function () {
			saveButton.disabled = name.value.trim() === "" || !changed();
		});
		saveButton.addEventListener("click", function () {
			const requests: Array<HttpRequest> = [];
			if (name.value.trim() !== App.getTranslation(list.name)) {
				const n = new Models.Translation();
				n.en = n.sv = n.de = name.value.trim();
				requests.push(new HttpRequest("POST", "/v1/lists/" + list.uuid + "/rename", { name: n }));
			}
			if (colorSelect.value !== null && colorSelect.value !== color) {
				requests.push(new HttpRequest("POST", "/v1/lists/" + list.uuid + "/color", { color: colorSelect.value }));
			}
			Ui.setBusy(saveButton, true);
			let pending = requests.length;
			requests.forEach(r => r.process(function () {
				if (--pending === 0) {
					self.mSheet.close();
					setTimeout(() => {
						location.reload();
					}, 150);
				}
			}, function () {
				Ui.setBusy(saveButton, false);
				new Ui.MessageBox(App.translations.Error, App.translations.SomethingWentWrong, "/img/messagebox-256.webp", "var(--error-color)", function () { });
			}));
		});

		sheet.open(closeCallback);
	}
}

function openEditListPopUp(list: Models.List, closeCallback: () => void = function () { }): void {
	new EditListPopUp(list, closeCallback);
}
