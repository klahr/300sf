/// <reference path="AdminCommon.ts"/>

function openAdminUsersPopUp(closeCallback: () => void = function () { }): void {
	const sheet = new Ui.Sheet({ title: App.translations.Users, size: "tall" });
	const section = sheet.section();
	let users: Array<Models.User> = [];
	const rows = Admin.list(section);
	const note = Admin.note(section);

	const render = function (query: string) {
		rows.innerHTML = "";
		const shown = Admin.filtered(users, query, u => u.display_name + " " + u.username);
		shown.forEach(u => {
			const seen = Utils.isValidDate(u.last_seen_at as any) ? Utils.toYYYY_MM_DD(u.last_seen_at as any) : App.translations.Never;
			const label = u.display_name + ((u.scopes || "").split(" ").indexOf("admin") > -1 ? " · " + App.translations.Admin : "");
			Admin.row(rows, label, seen, () => openAdminUserPopUp(u));
		});
		note.textContent = shown.length < users.length ? App.translations.ShowingOf.replace("{0}", String(shown.length)).replace("{1}", String(users.length)) : "";
	};
	const search = Admin.search(sheet, section, "admin-users-search", render);
	section.insertBefore(search.container, rows);

	new HttpRequest("GET", "/v1/users/list", null).process(function (status: any, data: Array<Models.User>) {
		users = data;
		// Most recently active first.
		users.sort((a, b) => String(b.last_seen_at || "").localeCompare(String(a.last_seen_at || "")));
		render(search.value);
	}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
		Admin.showError(error);
	});

	sheet.open(closeCallback);
}

function openAdminUserPopUp(user: Models.User): void {
	const sheet = new Ui.Sheet({ title: user.display_name, subtitle: "@" + user.username, size: "compact" });
	const rows = Admin.list(sheet.section());
	new HttpRequest("GET", "/v1/users/" + user.uuid, null).process(function (status: any, data: any) {
		const u = data.user as Models.User;
		Admin.row(rows, App.translations.Access, u.scopes || "", null);
		Admin.row(rows, App.translations.LastSeen, Utils.isValidDate(u.last_seen_at as any) ? Utils.toYYYY_MM_DD_hh_mm(u.last_seen_at as any) : App.translations.Never, null);
		Admin.row(rows, App.translations.Sightings, String((data.reports || []).length), null);
		if (u.language) {
			Admin.row(rows, App.translations.Language, u.language.code, null);
		}
	}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
		Admin.showError(error);
	});
	sheet.open();
}
