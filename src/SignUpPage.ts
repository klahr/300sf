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
		logo.src = "/img/99249498-452f-4b0b-a1f5-45902f7ffe48.png";
		logo.alt = "TODO";
		logo.style.width = "192px";
		logo.style.borderRadius = "50%";
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
		this.attach(acceptTerms);

		const acceptTermsLbl = document.createElement("label");
		acceptTermsLbl.setAttribute("for", acceptTerms.container.getAttribute("name")!);
		acceptTermsLbl.innerHTML = App.translations.AcceptTerms;
		this.container.appendChild(acceptTermsLbl);

		const help = document.createElement("p");
		help.className = "details";
		help.innerHTML = App.translations.MarkSignInCheckbox;
		this.container.appendChild(help);

		const signUp = document.createElement("button");
//		signUp.disabled = true;
		signUp.innerHTML = App.translations.SignUp;
		signUp.addEventListener("click", function() {
			const termsAccepted = (acceptTerms.container as HTMLInputElement).checked;
			const content = {
				email: email.value,
				display_name: displayName.value,
				password: password.value,
				icon_uuid: "",
				accepted_terms: termsAccepted,
			};

			const r = new HttpRequest("POST", "/v1/users", JSON.stringify(content));
			r.process(function(status: any, data: any) {
				password.value = "",
				self.appListener.showPage("home", {"email": email.value, "display_name": displayName.value});
			}, function(status: any, data: Responses.Error, error: any) {
				switch (data.code) {
					case Responses.ErrorCode.ErrorSignUpInvalidDisplayName:
						new Ui.MessageBox("Oh, no!", "Your <b>Display name</b> appears to be invalid.", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function() {});
						break;
					case Responses.ErrorCode.ErrorSignUpInvalidEmail:
						new Ui.MessageBox("Oh, no!", "Your <b>Email</b> appears to be invalid.", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function() {});
						break;
					case Responses.ErrorCode.ErrorSignUpInvalidPassword:
						new Ui.MessageBox("Oh, no!", "Your <b>Password</b> appears to be invalid.", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function() {});
						break;
					case Responses.ErrorCode.ErrorSignUpTermsNotAccepted:
						new Ui.MessageBox("Oh, no!", "You must accept the <b>terms</b>.", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function() {});
						break;
					default:
						new Ui.MessageBox("Oh, no!", "It appears that the <b>email</b> or <b>password</b> you entered is invalid.", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function() {});
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
		alreadyHaveAccount.addEventListener("click", function() {
			self.appListener.showPage("home", null);
		});
		this.container.appendChild(alreadyHaveAccount);
    }

	public get id(): string {
		return "sign-up";
	}
}
