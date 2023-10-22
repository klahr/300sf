class DeleteAccountPage extends Page {
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
		title.innerHTML = "Delete Account";
		this.container.appendChild(title);

		const password = new Ui.Input("password");
		password.style.position = "relative";
		password.setPlaceHolder(App.translations.Password);
		this.attach(password);

		const self = this;
		const set = document.createElement("button");
		set.innerHTML = "Delete Account";
		set.addEventListener("click", function() {
			password.setDisabled(true);
			const r = new HttpRequest("DELETE", "/v1/users/confirm_delete/" + self.data.token, JSON.stringify({
				password: password.value
			}));
			r.process(function(status: any, data: any) {
				password.setDisabled(false);
				password.value = ""
				setCookie("x300sf_at", "", 0);
				setCookie("lang", "", 0);
				new Ui.MessageBox("Account Deleted", "Your account has been inactivated or will be permanently deleted in 7 days. If you change your mind you must contact support@300sf.se within a week.", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--success-color)", function() {
					appListener.logOut();
				});
			});
		});
		password.container.addEventListener("keydown", function(e) {
			if (e.key == "Enter") {
				e.preventDefault();
				set.click();
			}
		});
		this.container.appendChild(set);
	}

	public get id(): string {
		return "delete-account";
	}

	public setData(data: any): void {
		super.setData(data);
	}
}