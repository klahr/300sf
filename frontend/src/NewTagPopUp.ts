/// <reference path="ui/Div.ts"/>
/// <reference path="HttpRequest.ts"/>

class NewTagPopUp {
	constructor(closeCallback: (tag: Models.Tag | null) => void) {
		const self = this;
		const content = new Ui.Div();
		content.style.textAlign = "left";
		content.style.width = "100%";
		content.style.position = "relative";

		const title = document.createElement("h1");
		title.innerHTML = App.translations.NewLabel;
		title.style.margin = "unset";
		content.container.appendChild(title);

		const name = new Ui.Input("text");
		name.setPlaceHolder(App.translations.Name);
		name.container.addEventListener("input", function () {
			if (name.value === "") {
				createButton.disabled = true;
				return;
			}
			createButton.disabled = false;
		});
		content.attach(name);

		const colorLbl = document.createElement("h3");
		colorLbl.className = "title";
		colorLbl.innerHTML = App.translations.Color;
		content.container.appendChild(colorLbl);

		const colorSelect = new Ui.DivSelect();
		colorSelect.addOption("", "#8227ac").style.backgroundColor = "#8227ac";
		colorSelect.addOption("", "#0d81b4").style.backgroundColor = "#0d81b4";
		colorSelect.addOption("", "#75b98e").style.backgroundColor = "#75b98e";
		colorSelect.addOption("", "#68e060").style.backgroundColor = "#68e060";
		colorSelect.addOption("", "#e06868").style.backgroundColor = "#e06868";
		colorSelect.addOption("", "#ac7827").style.backgroundColor = "#ac7827";
		content.attach(colorSelect);
		colorSelect.select("#0d81b4");

		const createButton = document.createElement("button");
		createButton.innerHTML = App.translations.CreateTag;
		createButton.style.width = "auto";
		createButton.style.cssFloat = "right";
		createButton.style.borderRadius = "8px";
		createButton.style.backgroundColor = "var(--success-color)";
		createButton.disabled = true;
		content.container.appendChild(createButton);
		createButton.addEventListener("click", function () {
			const r = new Requests.Tags();
			r.name =name.value;
			r.color = colorSelect.value!;
			const req = new HttpRequest("POST", "/v1/tags", r);
			req.process(function (status: any, data: any) {
				closeCallback(data);
				popUp.close();
			});
		});

		const popUp = new Ui.PopUp(content.container, function () {
			closeCallback(null);
		});
	}
}
