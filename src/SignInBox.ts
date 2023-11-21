/// <reference path="ui/MessageBox.ts"/>
/// <reference path="ui/Input.ts"/>
/// <reference path="App.ts"/>

class SignInBox extends Ui.MessageBox {
	constructor(title: string, message: string, image: string | null, color: string, yesCallback:() => void, closeCallback: () => void, yesText: string, cancelText: string) {
		super(title, message, image, color, closeCallback);

		const email = new Ui.Input("email");
		email.setPlaceHolder(App.translations.Email);
		email.container.addEventListener("click", (event) => {
			event.stopPropagation();
		});
		email.container.addEventListener("keydown", function(e) {
			if (e.key == "Enter") {
				e.preventDefault();
				password.select();
			}
		});
		this.popUp().attach(email);
		const password = new Ui.Input("password");
		password.setPlaceHolder(App.translations.Password);
		password.container.addEventListener("click", (event) => {
			event.stopPropagation();
		});
		password.container.addEventListener("keydown", function(e) {
			if (e.key == "Enter") {
				e.preventDefault();
				signInButton.click();
			}
		});
		this.popUp().attach(password);

		const d = document.createElement("div");
		d.style.width = "100%";
		d.style.textAlign = "left";
		this.popUp().container.appendChild(d);

		const staySignedIn = new Ui.Input("checkbox");
		staySignedIn.style.position = "relative";
		staySignedIn.container.setAttribute("name", "stay_signed_in");
		staySignedIn.container.id = "stay_signed_in";
		d.addEventListener("click", (event) => {
			event.stopPropagation();
		});
		d.appendChild(staySignedIn.container);

		const staySignedInLbl = document.createElement("label");
		staySignedInLbl.setAttribute("for", staySignedIn.container.id);
		staySignedInLbl.innerHTML = App.translations.StaySignedIn;
		d.appendChild(staySignedInLbl);

		const help = document.createElement("p");
		help.className = "details";
		help.innerHTML = App.translations.MarkSignInCheckbox;
		d.appendChild(help);

		const cancelButton = document.createElement("button");
		cancelButton.innerHTML = cancelText;
		cancelButton.style.width = "calc(50% - var(--small-distance)*2)";
		cancelButton.style.backgroundColor = "transparent";
		cancelButton.style.boxShadow = "0 0 0 1px inset white";
		this.popUp().container.appendChild(cancelButton);

		const self = this;
		const signInButton = document.createElement("button");
		signInButton.innerHTML = yesText;
		signInButton.style.width =  "calc(50% - var(--small-distance)*2)";
		signInButton.addEventListener("click", function(event) {
			event.stopPropagation();
			const authorization = 'Basic ' + btoa(email.value + ':' + password.value);
			const stayLoggedIn = (staySignedIn.container as HTMLInputElement).checked;

			const r = new HttpRequest("PUT", "/v1/auth");
			r.setHeader("Authorization", authorization);
			r.process(function(status: any, data: Responses.SingleItemDataResponse) {
				const content = data.data as Responses.AuthenticationResponse;
				const ks = new LocalStorage();
				ks.set("x300sf_at", content.access_token, stayLoggedIn ? 0 : content.expires_in);
				ks.set("lang", content.lang, stayLoggedIn ? 0 : content.expires_in);
				self.container.click();
			}, function(status: any, data: Responses.GeneralErrorResponse, err: any) {
				if (data.code === Responses.ErrorCode.ErrorAccountDeactivated) {
					new Ui.MessageBox(App.translations.Error, App.translations.AccountDeactivated, "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function() {});
				} else {
					new Ui.MessageBox(App.translations.Error, App.translations.InvalidEmailOrPassword, "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function() {});
				}
			});
		});
		this.popUp().container.appendChild(signInButton);
	}
}