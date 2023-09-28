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
		content.container.appendChild(logo);

		const displayName = document.createElement("h1");
		displayName.innerHTML = "Please wait...";
		content.container.appendChild(displayName);

		const uuidLbl = document.createElement("h2");
		uuidLbl.innerHTML = "Internal ID";
		uuidLbl.style.textAlign = "left";
		content.container.appendChild(uuidLbl);

		const uuid = new Ui.Input("text");
		uuid.container.setAttribute("disabled", "");
		content.attach(uuid);

		const languageLbl = document.createElement("h2");
		languageLbl.innerHTML = "Language";
		languageLbl.style.textAlign = "left";
		content.container.appendChild(languageLbl);

		const languageList = new Ui.Select("English", function(value: string | null) {

		});
		languageList.style.textAlign = "left";
		languageList.addOption("Svenska", "sv");
		languageList.addOption("English", "en");
		content.attach(languageList);

		const self = this;
		const me = new HttpRequest("GET", "/v1/users/me");
		me.process(function(status: any, data: any) {
			const d = data as Responses.UsersMe;
			console.log(d);
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
