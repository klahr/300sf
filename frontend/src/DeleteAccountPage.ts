/// <reference path="Page.ts"/>
/// <reference path="App.ts"/>

class DeleteAccountPage extends Page {
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
		logo.classList.add("shadow");
		logo.width = 192;
		logo.height = 192;
		logo.style.borderRadius = "50%";
		content.container.appendChild(logo);

		const title = document.createElement("h1");
		title.innerHTML = App.translations.DeleteAccount;
		this.container.appendChild(title);

		const password = new Ui.Input("password");
		password.style.position = "relative";
		password.setPlaceHolder(App.translations.Password);
		this.attach(password);

		const self = this;
		const set = document.createElement("button");
		set.className = "danger";
		set.innerHTML = App.translations.DeleteAccount;
		set.addEventListener("click", function () {
			password.setDisabled(true);
			const r = new HttpRequest("DELETE", "/v1/users/confirm_delete/" + self.data.token, {
				password: password.value
			});
			r.process(function (status: any, data: any) {
				password.setDisabled(false);
				password.value = ""
				const ks = new LocalStorage();
				ks.clear();
				new Ui.MessageBox(App.translations.AccountDeleted, App.translations.AccountDeletedDescription, "/img/messagebox-256.webp", "var(--success-color)", function () {
					appListener.logOut();
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
		return "delete-account";
	}

	public setData(data: any): void {
		super.setData(data);
	}
}