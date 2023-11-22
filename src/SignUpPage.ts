/// <reference path="HttpRequest.ts"/>
/// <reference path="ui/MessageBox.ts"/>

class SignUpPage extends Page {
	constructor(appListener: IAppListener) {
		super(appListener);

		const content = new Ui.Div();
		content.style.position = "relative";
		content.style.width = "100%";
		content.style.textAlign = "center";
		this.attach(content);
		const logo = document.createElement("img");
		logo.src = "/img/birds-192.webp";
		logo.alt = "TODO";
		logo.width = 192;
		logo.height = 192;
		logo.style.borderRadius = "50%";
		logo.classList.add("shadow");
		content.container.appendChild(logo);

		const title = document.createElement("h1");
		title.innerHTML = App.translations.SignUp;
		this.container.appendChild(title);

		const displayName = new Ui.Input("display_name");
		displayName.style.position = "relative";
		displayName.setPlaceHolder(App.translations.DisplayName);
		this.attach(displayName);

		const email = new Ui.Input("email");
		email.style.position = "relative";
		email.setPlaceHolder(App.translations.Email);
		this.attach(email);

		const password = new Ui.Input("password");
		password.style.position = "relative";
		password.setPlaceHolder(App.translations.Password);
		this.attach(password);

		const acceptTerms = new Ui.Input("checkbox");
		acceptTerms.style.position = "relative";
		acceptTerms.container.setAttribute("name", "accept_terms");
		acceptTerms.container.id = "accept_terms";
		this.attach(acceptTerms);

		const acceptTermsLbl = document.createElement("label");
		acceptTermsLbl.setAttribute("for", acceptTerms.container.id);
		acceptTermsLbl.innerHTML = App.translations.AcceptTerms;
		this.container.appendChild(acceptTermsLbl);

		this.container.appendChild(document.createElement("br"));
		const help = document.createElement("a");
		help.style.fontSize = "0.8rem";
		help.style.margin = "var(--small-distance)";
		help.style.display = "block";
		help.target = "_blank";
		help.innerHTML = App.translations.TermsOfService;
		help.href = "/terms";
		this.container.appendChild(help);

		const signUp = document.createElement("button");
		signUp.disabled = false;
		signUp.innerHTML = App.translations.SignUp;
		signUp.addEventListener("click", function () {
			const termsAccepted = (acceptTerms.container as HTMLInputElement).checked;
			const content = {
				email: email.value,
				display_name: displayName.value,
				password: password.value,
				portrait: "3.webp",
				accepted_terms: termsAccepted,
			};

			const r = new HttpRequest("POST", "/v1/users", content);
			r.process(function (status: any, data: any) {
				password.value = "",
					self.appListener.showPage("home", { "email": email.value, "display_name": displayName.value });
			}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
				switch (data.code) {
					case Responses.ErrorCode.ErrorSignUpInvalidDisplayName:
						new Ui.MessageBox(App.translations.Error, App.translations.InvalidDisplayName, "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function () { });
						break;
					case Responses.ErrorCode.ErrorSignUpInvalidEmail:
						new Ui.MessageBox(App.translations.Error, App.translations.InvalidEmail, "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function () { });
						break;
					case Responses.ErrorCode.ErrorSignUpInvalidPassword:
						new Ui.MessageBox(App.translations.Error, App.translations.InvalidPassword, "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function () { });
						break;
					case Responses.ErrorCode.ErrorSignUpTermsNotAccepted:
						new Ui.MessageBox(App.translations.Error, App.translations.InvalidAcceptTerms, "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function () { });
						break;
					default:
						new Ui.MessageBox(App.translations.Error, App.translations.SomethingWentWrong, "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function () { });
						break;
				}
			});
		});

		this.container.appendChild(signUp);

		const hr = document.createElement("hr");
		this.container.appendChild(hr);

		const self = this;

		const alreadyHaveAccount = document.createElement("p");
		alreadyHaveAccount.innerHTML = App.translations.AlreadyHaveAnAccount;
		this.container.appendChild(alreadyHaveAccount);
	}

	public get id(): string {
		return "sign-up";
	}
}
