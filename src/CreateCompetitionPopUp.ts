/// <reference path="ui/Div.ts"/>
/// <reference path="ui/MultiInput.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="models/Competition.ts"/>

class CreateCompetitionPopUp {
	constructor(closeCallback: (competition: Models.Competition | null) => void) {
		const content = new Ui.Div();
		content.style.textAlign = "left";
		content.style.width = "100%";
		content.style.position = "relative";

		const title = document.createElement("h1");
		title.innerHTML = "New competition";
		title.style.margin = "unset";
		content.container.appendChild(title);

		const nameLbl = document.createElement("h2");
		nameLbl.innerHTML = App.translations.Name;
		content.container.appendChild(nameLbl);

		const name = new Ui.MultiInput("text");
		content.attach(name);
		name.addOption(App.translations.English, App.translations.Name, "en");
		name.addOption(App.translations.Swedish, App.translations.Name, "sv");
		name.addOption(App.translations.German, App.translations.Name, "de");
		name.setSelected(App.language);

		const popUp = new Ui.PopUp(content.container, function () {
			closeCallback(null);
		});
	}
}