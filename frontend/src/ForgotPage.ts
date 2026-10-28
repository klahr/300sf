/// <reference path="App.ts"/>

class ForgotPage extends Page {
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
		content.container.appendChild(logo);

		const title = document.createElement("h1");
		title.innerHTML = App.translations.ForgotPassword;
		this.container.appendChild(title);

		const description = document.createElement("p");
		description.innerHTML = App.translations.ForgotPasswordDescription;
		this.container.appendChild(description);

		const email = new Ui.Input("email");
		email.style.position = "relative";
		email.setPlaceHolder(App.translations.Email);
		this.attach(email);

		let inFlight = false;
		const reset = document.createElement("button");
		reset.innerHTML = App.translations.SendResetLink;
		reset.addEventListener("click", function () {
			if (inFlight) {
				return;
			}
			inFlight = true;
			email.setDisabled(true);
			const r = new HttpRequest("POST", "/v1/users/reset_password", {
				email: email.value,
			});
			r.process(function (status: any, data: any) {
				email.value = "";
				new Ui.MessageBox(App.translations.ResetPassword, App.translations.CheckEmailInstructions, "/img/messagebox-256.webp", "var(--success-color)", function () {
					appListener.showPage("home", null);
				});
			}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
				// Never claim an email was sent when it was not.
				switch (data ? data.code : 0) {
					case Responses.ErrorCode.ErrorRecordNotFound:
						new Ui.MessageBox(App.translations.Error, App.translations.EmailNotFound, "/img/messagebox-256.webp", "var(--error-color)", function () { });
						break;
					case Responses.ErrorCode.ErrorSignUpInvalidEmail:
						new Ui.MessageBox(App.translations.Error, App.translations.InvalidEmail, "/img/messagebox-256.webp", "var(--error-color)", function () { });
						break;
					default:
						new Ui.MessageBox(App.translations.Error, App.translations.SomethingWentWrong, "/img/messagebox-256.webp", "var(--error-color)", function () { });
						break;
				}
			}, function () {
				inFlight = false;
				email.setDisabled(false);
			});
		});
		email.container.addEventListener("keydown", function (e) {
			if (e.key == "Enter") {
				e.preventDefault();
				reset.click();
			}
		});

		this.container.appendChild(reset);
	}

	public get id(): string {
		return "forgot";
	}
}
