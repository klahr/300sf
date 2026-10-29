/// <reference path="AdminCommon.ts"/>

// Every stored name: species, checklists, zones and the catalog.
function openAdminTranslationsPopUp(closeCallback: () => void = function () { }): void {
	const sheet = new Ui.Sheet({ title: App.translations.Translations, size: "tall" });
	const section = sheet.section();
	let translations: Array<Models.Translation> = [];
	const text = (t: Models.Translation) => [t.en, t.sv, t.de, t.la].join(" ");

	const search = Admin.search(sheet, section, "admin-translations-search", function (query) {
		render(query);
	});
	const rows = Admin.list(section);
	const note = Admin.note(section);

	const render = function (query: string) {
		rows.innerHTML = "";
		const shown = Admin.filtered(translations, query, text);
		shown.forEach(t => {
			Admin.row(rows, t.en || "–", t.la || t.sv || "", () => Admin.editName(t, true, function () {
				render(search.value);
			}));
		});
		note.textContent = shown.length < translations.length && query.trim() === "" ? App.translations.ShowingOf.replace("{0}", String(shown.length)).replace("{1}", String(translations.length)) : "";
	};

	new HttpRequest("GET", "/v1/translations/list", null).process(function (status: any, data: any) {
		translations = data.translations;
		translations.sort((a, b) => (a.en || "").localeCompare(b.en || ""));
		render(search.value);
	}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
		Admin.showError(error);
	});

	sheet.open(closeCallback);
}
