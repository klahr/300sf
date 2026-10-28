/// <reference path="HttpRequest.ts"/>
/// <reference path="ui/MessageBox.ts"/>

class SignUpPage extends Page {
	private mDisplayName: Ui.Input;

	constructor(appListener: IAppListener) {
		super(appListener);
		this.addClass("auth-page");

		const content = new Ui.Div();
		content.style.position = "relative";
		content.style.width = "100%";
		content.style.textAlign = "center";
		this.attach(content);
		const logo = document.createElement("img");
		logo.src = "/img/birds-192.webp";
		logo.alt = "";
		logo.width = 192;
		logo.height = 192;
		logo.style.borderRadius = "50%";
		logo.classList.add("shadow");
		content.container.appendChild(logo);

		const title = document.createElement("h1");
		title.innerHTML = App.translations.SignUp;
		this.container.appendChild(title);

		this.mDisplayName = new Ui.Input("text");
		this.mDisplayName.style.position = "relative";
		this.mDisplayName.setPlaceHolder(App.translations.DisplayName);
		this.attach(this.mDisplayName);

		const email = new Ui.Input("email");
		email.style.position = "relative";
		email.setPlaceHolder(App.translations.Email);
		this.attach(email);

		const password = new Ui.Input("password");
		password.style.position = "relative";
		password.setPlaceHolder(App.translations.Password);
		this.attach(password);

		const termsRow = document.createElement("div");
		termsRow.className = "check-row";
		this.container.appendChild(termsRow);

		const acceptTerms = new Ui.Input("checkbox");
		acceptTerms.container.setAttribute("name", "accept_terms");
		acceptTerms.container.id = "accept_terms";
		this.mChildren.push(acceptTerms);
		termsRow.appendChild(acceptTerms.container);

		const acceptTermsLbl = document.createElement("label");
		acceptTermsLbl.setAttribute("for", acceptTerms.container.id);
		acceptTermsLbl.innerHTML = App.translations.AcceptTerms + " (<a target='_blank' href='/terms'>" + App.translations.TermsOfService.toLowerCase() + "</a>)";
		termsRow.appendChild(acceptTermsLbl);

		const signUp = document.createElement("button");
		signUp.disabled = false;
		signUp.innerHTML = App.translations.SignUp;
		signUp.addEventListener("click", function () {
			const termsAccepted = (acceptTerms.container as HTMLInputElement).checked;
			const content = {
				email: email.value,
				display_name: self.mDisplayName.value,
				username: self.getUsername(),
				password: password.value,
				portrait: "3.webp",
				accepted_terms: termsAccepted,
			};

			Ui.setBusy(signUp, true);
			const r = new HttpRequest("POST", "/v1/users", content);
			r.process(function (status: any, data: any) {
				// Signed up: sign in right away instead of asking for the password again.
				const auth = new HttpRequest("POST", "/v1/auth", {
					grant_type: "password",
					email: content.email,
					password: content.password,
					scopes: ["user"],
					stay_signed_in: true
				});
				auth.setNoRetry();
				auth.process(function (status: any, data: Responses.AuthenticationResponse) {
					Session.store(data);
					window.open("/home", "_self");
				}, function () {
					Ui.setBusy(signUp, false);
					password.value = "";
					self.appListener.showPage("home", { "email": content.email, "display_name": content.display_name });
				});
			}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
				Ui.setBusy(signUp, false);
				switch (data ? data.code : 0) {
					case Responses.ErrorCode.ErrorSignUpInvalidDisplayName:
						new Ui.MessageBox(App.translations.Error, App.translations.InvalidDisplayName, "/img/messagebox-256.webp", "var(--error-color)", function () { });
						break;
					case Responses.ErrorCode.ErrorSignUpInvalidEmail:
						new Ui.MessageBox(App.translations.Error, App.translations.InvalidEmail, "/img/messagebox-256.webp", "var(--error-color)", function () { });
						break;
					case Responses.ErrorCode.ErrorSignUpInvalidPassword:
						new Ui.MessageBox(App.translations.Error, App.translations.InvalidPassword, "/img/messagebox-256.webp", "var(--error-color)", function () { });
						break;
					case Responses.ErrorCode.ErrorSignUpTermsNotAccepted:
						new Ui.MessageBox(App.translations.Error, App.translations.InvalidAcceptTerms, "/img/messagebox-256.webp", "var(--error-color)", function () { });
						break;
					default:
						new Ui.MessageBox(App.translations.Error, App.translations.SomethingWentWrong, "/img/messagebox-256.webp", "var(--error-color)", function () { });
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

	// The server only accepts a-z, 0-9 and ".#-" in usernames, so "Åsa Ödman" becomes "asa-odman#123".
	private getUsername(): string {
		let username = this.mDisplayName.value.toLowerCase().normalize("NFD").replace(/[\u0300-\u036f]/g, "");
		username = username.replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "").substring(0, 24);
		if (username.length < 3) {
			username = "birder";
		}
		const randomNumber = Math.floor(Math.random() * (999 - 100 + 1)) + 100;
		return username + "#" + randomNumber;
	}

	public get id(): string {
		return "sign-up";
	}
}
