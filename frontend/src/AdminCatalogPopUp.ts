/// <reference path="AdminCommon.ts"/>

// The fixed values a sighting is described with: orders, genders, ages, methods and activities.
function openAdminCatalogPopUp(closeCallback: () => void = function () { }): void {
	const sheet = new Ui.Sheet({ title: App.translations.Catalog, size: "auto" });
	const kinds = Admin.list(sheet.section());
	const entries: Array<[string, string]> = [
		[App.translations.Orders, "orders"],
		[App.translations.Genders, "genders"],
		[App.translations.Ages, "ages"],
		[App.translations.Methods, "methods"],
		[App.translations.Activities, "activities"],
	];
	entries.forEach(([title, endpoint]) => Admin.row(kinds, title, "", () => openAdminNamedRecordsPopUp(title, endpoint)));
	sheet.open(closeCallback);
}

function openAdminNamedRecordsPopUp(title: string, endpoint: string): void {
	const sheet = new Ui.Sheet({ title: title, size: "tall" });
	const rows = Admin.list(sheet.section());

	const add = sheet.section(App.translations.Add);
	const addButton = sheet.addAction(Ui.actionButton(App.translations.Add, "add", "success"));
	addButton.disabled = true;
	const name = new Admin.NameFields(sheet, add, "new-" + endpoint, false, null, function () {
		addButton.disabled = name.english === "";
	});

	const refresh = function () {
		new HttpRequest("GET", "/v1/" + endpoint + "/list", null).process(function (status: any, data: any) {
			rows.innerHTML = "";
			const records = data.result as Array<{ uuid: string, name: Models.Translation }>;
			records.sort(Utils.sortByNameOnCurrentLanguage);
			records.forEach(r => Admin.row(rows, App.getTranslation(r.name), "", () => Admin.editName(r.name, false, refresh)));
		}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
			Admin.showError(error);
		});
	};

	addButton.addEventListener("click", function () {
		addButton.disabled = true;
		name.setDisabled(true);
		new HttpRequest("POST", "/v1/" + endpoint, { name: name.value }).process(function () {
			name.clear();
			refresh();
		}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
			addButton.disabled = false;
			Admin.showError(error);
		}, function () {
			name.setDisabled(false);
		});
	});

	refresh();
	sheet.open();
}
