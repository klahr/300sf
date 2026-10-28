/// <reference path="ui/Div.ts"/>
/// <reference path="ui/Sheet.ts"/>
/// <reference path="HttpRequest.ts"/>

// Creates a label from inside the report form.
class NewTagPopUp {
	constructor(closeCallback: (tag: Models.Tag | null) => void) {
		const sheet = new Ui.Sheet({ title: App.translations.NewLabel, size: "compact" });

		const form = sheet.section();

		const name = new Ui.Input("text");
		name.container.id = "new-tag-name";
		name.setPlaceHolder(App.translations.Name);
		name.container.addEventListener("input", function () {
			createButton.disabled = name.value === "";
		});
		sheet.field(form, App.translations.Name, name.container);

		const colorSelect = new Ui.DivSelect();
		colorSelect.addClass("swatches");
		colorSelect.addOption("", "#8227ac").style.backgroundColor = "#8227ac";
		colorSelect.addOption("", "#0d81b4").style.backgroundColor = "#0d81b4";
		colorSelect.addOption("", "#75b98e").style.backgroundColor = "#75b98e";
		colorSelect.addOption("", "#68e060").style.backgroundColor = "#68e060";
		colorSelect.addOption("", "#e06868").style.backgroundColor = "#e06868";
		colorSelect.addOption("", "#ac7827").style.backgroundColor = "#ac7827";
		colorSelect.select("#0d81b4");
		sheet.field(form, App.translations.Color, colorSelect.container);

		const createButton = sheet.addAction(Ui.actionButton(App.translations.CreateTag, "add", "success"));
		createButton.disabled = true;
		createButton.addEventListener("click", function () {
			const r = new Requests.Tags();
			r.name = name.value;
			r.color = colorSelect.value!;
			Ui.setBusy(createButton, true);
			const req = new HttpRequest("POST", "/v1/tags", r);
			req.process(function (status: any, data: any) {
				closeCallback(data);
				sheet.close();
			}, function () {
				new Ui.MessageBox(App.translations.Error, App.translations.SomethingWentWrong, "/img/messagebox-256.webp", "var(--error-color)", function () { });
			}, function () {
				Ui.setBusy(createButton, false);
				createButton.disabled = name.value === "";
			});
		});

		sheet.open(function () {
			closeCallback(null);
		});
	}
}
