/// <reference path="models/Tag.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="utils/Sort.ts"/>

class EditTagPopUp {
	private mArchiveButton: HTMLButtonElement;
	private mPopUp: Ui.PopUp;

	constructor(tag: Models.Tag, closeCallback: () => void) {
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
		speciesList.value = tag.name;
		speciesList.setDisabled(true);
		content.attach(speciesList);

		this.mArchiveButton = document.createElement("button");
		this.mArchiveButton.innerHTML = App.translations.ArchiveLabel;
		this.mArchiveButton.style.width = "auto";
		this.mArchiveButton.style.cssFloat = "right";
		this.mArchiveButton.style.borderRadius = "8px";
		this.mArchiveButton.style.backgroundColor = "var(--danger)";
		this.mArchiveButton.addEventListener("click", function () {
			new Ui.ConfirmBox(App.translations.ArchiveLabel, App.translations.ConfirmArchiveLabel, "/img/messagebox-256.webp", "var(--success-color)", function () {
				const r = new HttpRequest("POST", "/v1/tags/" + tag.uuid + "/archive");
				r.process(function (status: any, data: any) {
					self.mPopUp.close();
					setTimeout(() => {
						location.reload();
					}, 150);
				});
			}, function () {
			}, App.translations.ArchiveLabel, App.translations.Cancel);
		});
		content.container.appendChild(this.mArchiveButton);

		this.mPopUp = new Ui.PopUp(content.container, function () {
			closeCallback();
		});
	}
}

function openEditTagPopUp(tag: Models.Tag, closeCallback: () => void = function () { }): void {
	new EditTagPopUp(tag, closeCallback);
}
