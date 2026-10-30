/// <reference path="AdminCommon.ts"/>
/// <reference path="models/ApiKey.ts"/>
/// <reference path="LocalStorage.ts"/>

// The user's own API keys, for scripts and other clients of the API.
function openApiKeysPopUp(closeCallback: () => void = function () { }): void {
	const sheet = new Ui.Sheet({ title: App.translations.ApiKeys, size: "tall" });
	const keys = Admin.list(sheet.section());

	const created = sheet.section();
	created.style.display = "none";

	const add = sheet.section(App.translations.CreateApiKey);
	const name = Admin.input(sheet, add, "api-key-name", App.translations.Name);
	// Admins may give a key their admin scope too.
	const scope = new LocalStorage().get("scope");
	const isAdmin = !!scope && scope.includes("admin");
	const admin = new Ui.Input("checkbox");
	if (isAdmin) {
		const row = document.createElement("div");
		row.className = "check-row";
		admin.container.id = "api-key-admin";
		row.appendChild(admin.container);
		const lbl = document.createElement("label");
		lbl.setAttribute("for", admin.container.id);
		lbl.textContent = App.translations.ApiKeyAdminScope;
		row.appendChild(lbl);
		add.appendChild(row);
	}
	const create = sheet.addAction(Ui.actionButton(App.translations.CreateApiKey, "key", "success"));
	create.disabled = true;
	name.container.addEventListener("input", function () {
		create.disabled = name.value.trim() === "";
	});

	const refresh = function () {
		new HttpRequest("GET", "/v1/api_keys/list", null).process(function (status: any, data: any) {
			keys.innerHTML = "";
			(data.api_keys as Array<Models.ApiKey>).forEach(k => {
				const revoked = Utils.isValidDate(k.revoked_at as any);
				const meta = revoked ? App.translations.Revoked + " " + Utils.toYYYY_MM_DD(k.revoked_at as any) : k.website ? App.translations.UsedByWebsite : App.translations.ClientId + " " + k.client_id + " · " + k.scopes;
				const r = Admin.row(keys, k.name + " (" + k.prefix + ")", meta, null);
				if (revoked) {
					r.classList.add("archived");
					return;
				}
				if (k.website) {
					return;
				}
				const revoke = document.createElement("button");
				revoke.className = "icon-button delete";
				revoke.title = App.translations.Revoke;
				revoke.setAttribute("aria-label", App.translations.Revoke + ": " + k.name);
				revoke.innerHTML = "<span class='material-symbols-outlined'>key_off</span>";
				r.appendChild(revoke);
				revoke.addEventListener("click", function () {
					new Ui.ConfirmBox(App.translations.AreYouSure, App.translations.ConfirmRevokeApiKey, null, "var(--error-color)", function () {
						new HttpRequest("POST", "/v1/api_keys/" + encodeURIComponent(k.prefix) + "/revoke", {}).process(refresh, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
							Admin.showError(error);
						});
					}, function () { }, App.translations.Revoke, App.translations.Cancel);
				});
			});
		}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
			Admin.showError(error);
		});
	};

	create.addEventListener("click", function () {
		create.disabled = true;
		const scopes = (admin.container as HTMLInputElement).checked ? ["user", "admin"] : ["user"];
		new HttpRequest("POST", "/v1/api_keys", { name: name.value.trim(), scopes: scopes }).process(function (status: any, data: any) {
			name.value = "";
			(admin.container as HTMLInputElement).checked = false;
			// The secret is only ever returned here; a client signs in with it and the client ID.
			created.innerHTML = "";
			created.style.display = "";
			const hint = document.createElement("p");
			hint.textContent = App.translations.ApiKeyCreatedHint;
			created.appendChild(hint);
			[[App.translations.ClientId, data.client_id], [App.translations.ApiKeySecret, data.api_key]].forEach(function (pair) {
				const field = document.createElement("div");
				created.appendChild(field);
				const lbl = document.createElement("label");
				lbl.textContent = pair[0];
				field.appendChild(lbl);
				const value = new Ui.Input("text");
				value.value = pair[1];
				(value.container as HTMLInputElement).readOnly = true;
				field.appendChild(value.container);
				const copy = Ui.actionButton(App.translations.Copy, "content_copy", "inline");
				field.appendChild(copy);
				copy.addEventListener("click", function () {
					navigator.clipboard.writeText(pair[1]).then(function () {
						new Ui.Toast(App.translations.Copied, "", null, "var(--success-color)", function () { });
					});
				});
			});
			refresh();
		}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
			create.disabled = false;
			Admin.showError(error);
		});
	});

	refresh();
	sheet.open(closeCallback);
}
