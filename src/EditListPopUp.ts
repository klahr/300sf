/// <reference path="responses/Reports.ts"/>
/// <reference path="responses/Lists.ts"/>
/// <reference path="models/List.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="utils/Sort.ts"/>

class EditListPopUp {
	private mArchiveButton: HTMLButtonElement;
	private mPopUp: Ui.PopUp;

	constructor(list: Models.List, closeCallback: () => void) {
		const self = this;

		const content = new Ui.Div();
		content.style.textAlign = "left";
		content.style.width = "100%";
		content.style.position = "relative";

		const name = document.createElement("h1");
		name.innerHTML = App.translations.Checklist;
		name.style.margin = "unset";
		content.container.appendChild(name);

		const speciesList = new Ui.Input("text")
		speciesList.value = App.getTranslation(list.name);
		speciesList.setDisabled(true);
		content.attach(speciesList);

		this.mArchiveButton = document.createElement("button");
		this.mArchiveButton.innerHTML = App.translations.ArchiveChecklist;
		this.mArchiveButton.style.width = "auto";
		this.mArchiveButton.style.cssFloat = "right";
		this.mArchiveButton.style.borderRadius = "8px";
		this.mArchiveButton.style.backgroundColor = "#a11f1f";
		this.mArchiveButton.addEventListener("click", function() {
			new Ui.ConfirmBox(App.translations.ArchiveChecklist, App.translations.ConfirmArchiveChecklist, null, "var(--success-color)", function() {
				const r = new HttpRequest("POST", "/v1/lists/" + list.uuid + "/archive");
				r.process(function(status: any, data: any) {
					self.mPopUp.close();
					setTimeout(() => {
						location.reload();
					}, 150);
				});
			}, function() {
			}, App.translations.ArchiveChecklist, App.translations.Cancel);
		});
		content.container.appendChild(this.mArchiveButton);

		this.mPopUp = new Ui.PopUp(content.container, function() {
			closeCallback();
		});
	}
}

function openEditListPopUp(list: Models.List, closeCallback: () => void = function() {}): void {
	new EditListPopUp(list, closeCallback);
}
