/// <reference path="AdminCommon.ts"/>

const zoneTypes: Array<[number, () => string]> = [
	[2, () => App.translations.Country],
	[4, () => App.translations.County],
	[7, () => App.translations.Municipality],
	[10, () => App.translations.Habitat],
];

function zoneTypeName(level: number): string {
	const t = zoneTypes.find(([l]) => l === level);
	return t ? t[1]() : String(level);
}

function openAdminZonesPopUp(closeCallback: () => void = function () { }): void {
	const sheet = new Ui.Sheet({ title: App.translations.Zones, size: "tall" });
	const rows = Admin.list(sheet.section());

	const add = sheet.section(App.translations.AddZone);
	const addButton = sheet.addAction(Ui.actionButton(App.translations.AddZone, "add", "success"));
	let polygon = "";
	const verify = function () {
		addButton.disabled = name.english === "" || type.value === null || polygon === "";
	};
	const name = new Admin.NameFields(sheet, add, "new-zone", false, null, verify);

	const type = new Ui.Select(App.translations.ZoneType, App.translations.Search, verify);
	zoneTypes.forEach(([level, label]) => type.addOption(label(), String(level)));
	sheet.field(add, App.translations.ZoneType, type.container);

	const parent = new Ui.Select(App.translations.None, App.translations.Search);
	sheet.field(add, App.translations.ParentZone, parent.container);

	const file = new Ui.Input("file");
	file.container.id = "new-zone-polygon";
	(file.container as HTMLInputElement).accept = ".json,.geojson,application/geo+json,application/json";
	sheet.field(add, App.translations.PolygonFile, file.container);
	file.container.addEventListener("change", function () {
		const f = (file.container as HTMLInputElement).files;
		polygon = "";
		verify();
		if (f && f.length > 0) {
			f[0].text().then(function (t) {
				polygon = t;
				verify();
			});
		}
	});
	verify();

	const refresh = function () {
		new HttpRequest("GET", "/v1/zones/list", null).process(function (status: any, data: any) {
			rows.innerHTML = "";
			parent.clear();
			parent.addOption(App.translations.None, "");
			const zones = data.result as Array<Models.Zone>;
			zones.sort(Utils.sortByNameOnCurrentLanguage);
			zones.forEach(z => {
				const where = z.parent_zone ? " · " + App.getTranslation(z.parent_zone.name) : "";
				Admin.row(rows, App.getTranslation(z.name), zoneTypeName(z.admin_level) + where, () => Admin.editName(z.name, false, refresh));
				parent.addOption(App.getTranslation(z.name), z.uuid);
			});
		}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
			Admin.showError(error);
		});
	};

	addButton.addEventListener("click", function () {
		addButton.disabled = true;
		name.setDisabled(true);
		const d = {
			name: name.value,
			admin_level: Number(type.value),
			parent_zone_uuid: parent.value || "",
			polygon: polygon,
		};
		new HttpRequest("POST", "/v1/zones", d).process(function () {
			name.clear();
			(file.container as HTMLInputElement).value = "";
			polygon = "";
			refresh();
		}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
			Admin.showError(error);
		}, function () {
			name.setDisabled(false);
			verify();
		});
	});

	refresh();
	sheet.open(closeCallback);
}
