/// <reference path="models/List.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="ui/Sheet.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="utils/Sort.ts"/>

// A checklist in the management list: what it holds, and archiving it.
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

		sheet.open(closeCallback);
	}
}

function openEditListPopUp(list: Models.List, closeCallback: () => void = function () { }): void {
	new EditListPopUp(list, closeCallback);
}
