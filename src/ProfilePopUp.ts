/// <reference path="ui/PopUp.ts"/>
/// <reference path="ui/DivSelect.ts"/>
/// <reference path="ui/Input.ts"/>
/// <reference path="ui/Select.ts"/>
/// <reference path="HttpRequest.ts"/>

class ProfilePopUp {
	private mPopUp: Ui.PopUp;

	constructor(closeCallback: () => void) {
		const content = new Ui.Div();
		content.style.textAlign = "center";
		content.style.width = "100%";
		content.style.position = "relative";

		const portraitSelect = new Ui.DivSelect((value) => {
			const r = new HttpRequest("POST", "/v1/users/portrait", JSON.stringify({"portrait": value}));
			r.process(function(status: any, data: any) {
				App.user.portrait = value!;
				location.reload();
			});
		});
		portraitSelect.addOption("", "1.png").style.backgroundImage = "url(/img/portraits/1.png)";
		portraitSelect.addOption("", "2.png").style.backgroundImage = "url(/img/portraits/2.png)";
		portraitSelect.addOption("", "3.png").style.backgroundImage = "url(/img/portraits/3.png)";
		portraitSelect.addOption("", "4.png").style.backgroundImage = "url(/img/portraits/4.png)";
		portraitSelect.addOption("", "5.png").style.backgroundImage = "url(/img/portraits/5.png)";
		portraitSelect.addOption("", "6.png").style.backgroundImage = "url(/img/portraits/6.png)";
		portraitSelect.addOption("", "7.png").style.backgroundImage = "url(/img/portraits/7.png)";
		portraitSelect.addOption("", "8.png").style.backgroundImage = "url(/img/portraits/8.png)";
		portraitSelect.addOption("", "9.png").style.backgroundImage = "url(/img/portraits/9.png)";
		portraitSelect.addOption("", "10.png").style.backgroundImage = "url(/img/portraits/10.png)";
		content.attach(portraitSelect);

		const displayName = document.createElement("h1");
		displayName.innerHTML = App.translations.PleaseWait + "...";
		content.container.appendChild(displayName);

		const uuidLbl = document.createElement("h2");
		uuidLbl.innerHTML = App.translations.InternalID;
		uuidLbl.style.textAlign = "left";
		content.container.appendChild(uuidLbl);

		const uuid = new Ui.Input("text");
		uuid.container.setAttribute("disabled", "");
		content.attach(uuid);

		const languageLbl = document.createElement("h2");
		languageLbl.innerHTML = App.translations.Language;
		languageLbl.style.textAlign = "left";
		content.container.appendChild(languageLbl);

		let lang;
		switch (App.language) {
			case "sv":
				lang = "Svenska";
				break;
			case "de":
				lang = "Deutch";
				break;
			default:
				lang = "English";
				break;
		}

		const languageList = new Ui.Select(lang, App.translations.Search, function(value: string | null) {
			if (value) {
				const d = {
					"language_code": value
				};
				const r = new HttpRequest("POST", "/v1/users/language", JSON.stringify(d));
				r.process(function(status: any, data: any) {
					App.setLanguage(value);
					setCookie("lang", value);
					location.reload();
				});
			}
		});
		languageList.style.textAlign = "left";
		languageList.addOption("Deutch", "de");
		languageList.addOption("English", "en");
		languageList.addOption("Svenska", "sv");
		content.attach(languageList);

		const me = new HttpRequest("GET", "/v1/users/me");
		me.process(function(status: any, data: any) {
			const d = data as Responses.UsersMe;
			displayName.innerHTML = d.user.display_name;
			uuid.value = d.user.uuid;
			portraitSelect.select(d.user.portrait);
		});

		this.mPopUp = new Ui.PopUp(content.container, function() {
			closeCallback();
		});
	}
}

function openProfilePopUp(closeCallback: () => void = function() {}): void {
	new ProfilePopUp(closeCallback);
}
