/// <reference path="models/Tag.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="ui/Sheet.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="utils/Sort.ts"/>

// A label in the management list, and archiving it.
class EditTagPopUp {
	private mSheet: Ui.Sheet;

	constructor(tag: Models.Tag, closeCallback: () => void) {
		const self = this;

		const sheet = new Ui.Sheet({ eyebrow: App.translations.Label, title: tag.name, size: "compact" });
		this.mSheet = sheet;

		const dot = document.createElement("span");
		dot.className = "list-dot";
		dot.style.setProperty("--dot", tag.color !== "" ? tag.color : "var(--blue)");
		sheet.names.querySelector("p.sheet-eyebrow")?.prepend(dot);

		sheet.addAction(Ui.actionButton(App.translations.ArchiveLabel, "archive", "danger")).addEventListener("click", function () {
			new Ui.ConfirmBox(App.translations.ArchiveLabel, App.translations.ConfirmArchiveLabel, "/img/messagebox-256.webp", "var(--error-color)", function () {
				const r = new HttpRequest("POST", "/v1/tags/" + tag.uuid + "/archive");
				r.process(function (status: any, data: any) {
					self.mSheet.close();
					setTimeout(() => {
						location.reload();
					}, 150);
				});
			}, function () {
			}, App.translations.ArchiveLabel, App.translations.Cancel);
		});

		sheet.open(closeCallback);
	}
}

function openEditTagPopUp(tag: Models.Tag, closeCallback: () => void = function () { }): void {
	new EditTagPopUp(tag, closeCallback);
}
