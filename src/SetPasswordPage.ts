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
		logo.src = "/img/birds-256.png";
		logo.alt = "TODO";
		logo.width = 192;
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
		const set = document.createElement("button");
		set.innerHTML = "Set Password";
		set.addEventListener("click", function () {
			password.setDisabled(true);
			const r = new HttpRequest("POST", "/v1/users/set_password/" + self.data.token, {
				password: password.value,
			});
			r.process(function (status: any, data: any) {
				password.setDisabled(false);
				password.value = "";
				new Ui.MessageBox(App.translations.NewPasswordSet, App.translations.SuccessfullyResetPassword, "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--success-color)", function () {
					appListener.showPage("home", null);
				});
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