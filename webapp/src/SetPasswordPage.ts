/// <reference path="App.ts"/>

class SetPasswordPage extends Page {
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
		content.container.appendChild(logo);

		const title = document.createElement("h1");
		title.innerHTML = "New password";
		this.container.appendChild(title);

		const password = new Ui.Input("password");
		password.style.position = "relative";
		password.setPlaceHolder(App.translations.Password);
		this.attach(password);

		const self = this;
		let inFlight = false;
		const set = document.createElement("button");
		set.innerHTML = "Set Password";
		set.addEventListener("click", function () {
			// The token is single use, a second submission would report the link as
			// expired right after the first one succeeded.
			if (inFlight) {
				return;
			}
			const token = self.data ? self.data.token : null;
			if (!token) {
				new Ui.MessageBox(App.translations.Error, App.translations.ResetLinkInvalid, "/img/messagebox-256.webp", "var(--error-color)", function () {
					appListener.showPage("forgot", null);
				});
				return;
			}
			inFlight = true;
			password.setDisabled(true);
			const r = new HttpRequest("POST", "/v1/users/set_password/" + token, {
				password: password.value,
			});
			r.process(function (status: any, data: any) {
				password.value = "";
				new Ui.MessageBox(App.translations.NewPasswordSet, App.translations.SuccessfullyResetPassword, "/img/messagebox-256.webp", "var(--success-color)", function () {
					appListener.showPage("home", null);
				});
			}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
				switch (data ? data.code : 0) {
					case Responses.ErrorCode.ErrorLinkExpired:
						new Ui.MessageBox(App.translations.Error, App.translations.ResetLinkExpired, "/img/messagebox-256.webp", "var(--error-color)", function () {
							appListener.showPage("forgot", null);
						});
						break;
					case Responses.ErrorCode.ErrorRecordNotFound:
						new Ui.MessageBox(App.translations.Error, App.translations.ResetLinkInvalid, "/img/messagebox-256.webp", "var(--error-color)", function () {
							appListener.showPage("forgot", null);
						});
						break;
					case Responses.ErrorCode.ErrorSignUpInvalidPassword:
						new Ui.MessageBox(App.translations.Error, App.translations.InvalidPassword, "/img/messagebox-256.webp", "var(--error-color)", function () { });
						break;
					default:
						new Ui.MessageBox(App.translations.Error, App.translations.SomethingWentWrong, "/img/messagebox-256.webp", "var(--error-color)", function () { });
						break;
				}
			}, function () {
				// Runs on success and on failure alike. Without this the input stays
				// greyed out after an error and the button looks dead.
				inFlight = false;
				password.setDisabled(false);
			});
		});
		password.container.addEventListener("keydown", function (e) {
			if (e.key == "Enter") {
				e.preventDefault();
				set.click();
			}
		});
		this.container.appendChild(set);
	}

	public get id(): string {
		return "set-password";
	}

	public setData(data: any): void {
		super.setData(data);
	}
}