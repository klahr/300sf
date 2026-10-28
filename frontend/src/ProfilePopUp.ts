/// <reference path="ui/PopUp.ts"/>
/// <reference path="ui/Sheet.ts"/>
/// <reference path="ui/DivSelect.ts"/>
/// <reference path="ui/Input.ts"/>
/// <reference path="ui/Select.ts"/>
/// <reference path="HttpRequest.ts"/>

// The signed-in user: portrait, account details and language.
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
		const username = addDetail(App.translations.Username, "alternate_email");
		const uuid = addDetail(App.translations.InternalID, "fingerprint");

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
			username.textContent = d.user.username;
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
