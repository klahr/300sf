/// <reference path="ui/PopUp.ts"/>
/// <reference path="ui/Sheet.ts"/>
/// <reference path="ui/DivSelect.ts"/>
/// <reference path="ui/Input.ts"/>
/// <reference path="ui/Select.ts"/>
/// <reference path="HttpRequest.ts"/>

// The signed-in user: portrait, name, username, email and language.
class ProfilePopUp {
	constructor(closeCallback: () => void) {
		const sheet = new Ui.Sheet({ eyebrow: App.translations.Profile, title: App.user ? App.user.display_name : App.translations.PleaseWait + "...", subtitle: App.user ? "@" + App.user.username : " ", size: "auto" });

		sheet.header.classList.add("with-portrait");
		const portrait = document.createElement("img");
		portrait.alt = "";
		if (App.user) {
			portrait.src = "/img/portraits/" + App.user.portrait;
		}
		sheet.header.insertBefore(portrait, sheet.names);

		const portraits = sheet.section(App.translations.Portrait);
		const portraitSelect = new Ui.DivSelect((value) => {
			const r = new HttpRequest("POST", "/v1/users/portrait", { "portrait": value });
			r.process(function (status: any, data: any) {
				App.user.portrait = value!;
				location.reload();
			});
		});
		for (let i = 1; i <= 10; ++i) {
			portraitSelect.addOption("", i + ".webp").style.backgroundImage = "url(/img/portraits/" + i + ".webp)";
		}
		portraitSelect.addClass("portraits");
		portraits.appendChild(portraitSelect.container);

		const account = sheet.section(App.translations.Account);
		const details = document.createElement("div");
		details.className = "detail-grid";
		details.style.gridTemplateColumns = "1fr";
		account.appendChild(details);
		const addDetail = function (label: string, icon: string): HTMLParagraphElement {
			const item = document.createElement("div");
			const l = document.createElement("p");
			l.className = "detail-label";
			l.innerHTML = "<span class='material-symbols-outlined'>" + icon + "</span>" + label;
			item.appendChild(l);
			const v = document.createElement("p");
			v.className = "detail-value";
			v.innerHTML = "<span class='skeleton line short'></span>";
			item.appendChild(v);
			details.appendChild(item);
			return v;
		};
		const uuid = addDetail(App.translations.InternalID, "fingerprint");

		const error = function (message: string) {
			new Ui.MessageBox(App.translations.Error, message, "/img/messagebox-256.webp", "var(--error-color)", function () { });
		};
		const errors: { [code: number]: string } = {
			4001: App.translations.InvalidDisplayName,
			4002: App.translations.InvalidNewEmail,
			4005: App.translations.InvalidUsername,
			4006: App.translations.UsernameTaken,
			7002: App.translations.TooManyWrongPasswords,
			[-30002]: App.translations.WrongPassword,
		};
		// A field with its own save button, enabled once the value changes.
		const editable = function (label: string, input: Ui.Input, save: (done: () => void) => void): HTMLButtonElement {
			const row = document.createElement("div");
			row.className = "inline-field";
			row.appendChild(input.container);
			const button = Ui.actionButton(App.translations.Save, "check", "secondary");
			button.disabled = true;
			row.appendChild(button);
			sheet.field(account, label, row);
			button.addEventListener("click", function () {
				Ui.setBusy(button, true);
				save(function () {
					Ui.setBusy(button, false);
				});
			});
			return button;
		};
		const post = function (path: string, body: any, done: () => void, success: () => void) {
			new HttpRequest("POST", path, body).process(success, function (status: any, data: Responses.GeneralErrorResponse) {
				error((data && errors[data.code]) || App.translations.SomethingWentWrong);
			}, done);
		};

		let current = { display_name: "", username: "" };
		const displayName = new Ui.Input("text");
		displayName.container.id = "profile-display-name";
		const saveDisplayName = editable(App.translations.DisplayName, displayName, function (done) {
			post("/v1/users/display_name", { display_name: displayName.value.trim() }, done, function () {
				location.reload();
			});
		});
		displayName.container.addEventListener("input", function () {
			saveDisplayName.disabled = displayName.value.trim() === "" || displayName.value.trim() === current.display_name;
		});

		const username = new Ui.Input("text");
		username.container.id = "profile-username";
		const saveUsername = editable(App.translations.Username, username, function (done) {
			post("/v1/users/username", { username: username.value.trim() }, done, function () {
				location.reload();
			});
		});
		username.container.addEventListener("input", function () {
			saveUsername.disabled = username.value.trim() === "" || username.value.trim() === current.username;
		});

		// A new email takes the password, and the old address is told.
		const emailSection = sheet.section(App.translations.ChangeEmail);
		const newEmail = new Ui.Input("email");
		newEmail.container.id = "profile-new-email";
		sheet.field(emailSection, App.translations.NewEmail, newEmail.container);
		const password = new Ui.Input("password");
		password.container.id = "profile-email-password";
		password.container.setAttribute("autocomplete", "current-password");
		sheet.field(emailSection, App.translations.CurrentPassword, password.container);
		const changeEmail = Ui.actionButton(App.translations.ChangeEmail, "mail", "secondary");
		changeEmail.disabled = true;
		emailSection.appendChild(changeEmail);
		const emailReady = function () {
			changeEmail.disabled = !newEmail.value.includes("@") || password.value === "";
		};
		newEmail.container.addEventListener("input", emailReady);
		password.container.addEventListener("input", emailReady);
		changeEmail.addEventListener("click", function () {
			Ui.setBusy(changeEmail, true);
			post("/v1/users/email", { email: newEmail.value.trim(), password: password.value }, function () {
				Ui.setBusy(changeEmail, false);
				password.value = "";
				emailReady();
			}, function () {
				newEmail.value = "";
				new Ui.MessageBox(App.translations.Success, App.translations.EmailChanged, "/img/messagebox-256.webp", "var(--success-color)", function () { });
			});
		});

		const language = sheet.section(App.translations.Language);
		let lang;
		switch (App.language) {
			case "sv":
				lang = "Svenska";
				break;
			case "de":
				lang = "Deutsch";
				break;
			default:
				lang = "English";
				break;
		}
		const languageList = new Ui.Select(lang, App.translations.Search, function (value: string | null) {
			if (value) {
				const d = {
					"language_code": value
				};
				const r = new HttpRequest("POST", "/v1/users/language", d);
				r.process(function (status: any, data: any) {
					// TODO: This is a bug. lang should have an expire time.
					App.setLanguage(value);
					const ks = new LocalStorage();
					ks.set("lang", value, 0);
					location.reload();
				});
			}
		});
		languageList.addOption("Deutsch", "de");
		languageList.addOption("English", "en");
		languageList.addOption("Svenska", "sv");
		language.appendChild(languageList.container);

		const me = new HttpRequest("GET", "/v1/users/me");
		me.process(function (status: any, data: any) {
			const d = data as Responses.UsersMe;
			sheet.title.textContent = d.user.display_name;
			const handle = sheet.names.querySelector("p.sheet-subtitle");
			if (handle) {
				handle.textContent = "@" + d.user.username;
			}
			current = { display_name: d.user.display_name, username: d.user.username };
			displayName.value = d.user.display_name;
			username.value = d.user.username;
			uuid.textContent = d.user.uuid;
			portrait.src = "/img/portraits/" + d.user.portrait;
			portraitSelect.select(d.user.portrait);
		});

		sheet.open(closeCallback);
	}
}

function openProfilePopUp(closeCallback: () => void = function () { }): void {
	new ProfilePopUp(closeCallback);
}
