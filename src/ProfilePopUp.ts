class ProfilePopUp {
	private mPopUp: Ui.PopUp;

	constructor(closeCallback: () => void) {
		const content = new Ui.Div();
		content.style.textAlign = "center";
		content.style.width = "100%";
		content.style.position = "relative";

		const logo = document.createElement("img");
		logo.src = "/img/99249498-452f-4b0b-a1f5-45902f7ffe48.png";
		logo.alt = "TODO";
		logo.style.width = "192px";
		logo.style.borderRadius = "50%";
		logo.style.textAlign = "center";
		logo.classList.add("shadow");
		content.container.appendChild(logo);

		const displayName = document.createElement("h1");
		displayName.innerHTML = "Please wait...";
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
		switch (App.getLanguage()) {
			case "sv":
				lang = "Svenska";
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
		languageList.addOption("Svenska", "sv");
		languageList.addOption("English", "en");
		content.attach(languageList);

		const me = new HttpRequest("GET", "/v1/users/me");
		me.process(function(status: any, data: any) {
			const d = data as Responses.UsersMe;
			displayName.innerHTML = d.user.display_name;
			uuid.value = d.user.uuid;
		});

		this.mPopUp = new Ui.PopUp(content.container, function() {
			closeCallback();
		});
	}
}

function openProfilePopUp(closeCallback: () => void = function() {}): void {
	new ProfilePopUp(closeCallback);
}
