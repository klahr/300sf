/// <reference path="App.ts"/>

class ForgotPage extends Page {
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
		title.innerHTML = App.translations.ForgotPassword;
		this.container.appendChild(title);

		const description = document.createElement("p");
		description.innerHTML = App.translations.ForgotPasswordDescription;
		this.container.appendChild(description);

		const email = new Ui.Input("email");
		email.style.position = "relative";
		email.setPlaceHolder(App.translations.Email);
		this.attach(email);

		const reset = document.createElement("button");
		reset.innerHTML = App.translations.SendResetLink;
		reset.addEventListener("click", function() {
			email.setDisabled(true);
			const r = new HttpRequest("POST", "/v1/users/reset_password", {
				email: email.value,
			});
			r.process(function(status: any, data: any) {
				email.setDisabled(false);
				email.value = "";
				new Ui.MessageBox("Reset Password", "Check your email inbox for further instructions.", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--success-color)", function() {
					appListener.showPage("home", null);
				});
			});
		});
		email.container.addEventListener("keydown", function(e) {
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
