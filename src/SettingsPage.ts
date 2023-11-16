/// <reference path="utils/Date.ts"/>
/// <reference path="Page.ts"/>
/// <reference path="DownloadFile.ts"/>
/// <reference path="ui/Input.ts"/>
/// <reference path="ui/MessageBox.ts"/>
/// <reference path="App.ts"/>

class SettingsPage extends Page {
	private mEmail: Ui.Input;
	private mPassword: Ui.Input;
	private mImportButton: HTMLButtonElement;

	constructor(appListener: IAppListener) {
		super(appListener);

		const self = this;

		this.addProfile();

		const title = document.createElement("h1");
		title.innerHTML = App.translations.Account;
		this.container.appendChild(title);

		const importSubmissionsLbl = document.createElement("h2");
		importSubmissionsLbl.innerHTML = App.translations.ImportData;
		this.container.appendChild(importSubmissionsLbl);

		const importSubmissionsDescription = document.createElement("p");
		importSubmissionsDescription.innerHTML = App.translations.ImportDataDescription;
		this.container.appendChild(importSubmissionsDescription);

		this.mEmail = new Ui.Input("email");
		this.mPassword = new Ui.Input("password");
		this.mImportButton = document.createElement("button");

		this.mEmail.style.position = "relative";
		this.mEmail.setPlaceHolder(App.translations.Email);
		this.mEmail.container.addEventListener("keydown", function (e) {
			if (e.key == "Enter") {
				e.preventDefault();
				self.mPassword.select();
			}
		});
		this.attach(this.mEmail);

		this.mPassword.style.position = "relative";
		this.mPassword.setPlaceHolder(App.translations.Password);
		this.mPassword.container.addEventListener("keydown", function (e) {
			if (e.key == "Enter") {
				e.preventDefault();
				self.mImportButton.click();
			}
		});
		this.attach(this.mPassword);

		this.mImportButton.innerHTML = "Import";
		this.mImportButton.addEventListener("click", function () {
			const d = {
				old_password: self.mPassword.value,
				old_username: self.mEmail.value
			};

			self.mImportButton.disabled = true;
			self.mImportButton.innerHTML = "Please wait...";
			self.mEmail.container.setAttribute("disabled", "");
			self.mPassword.container.setAttribute("disabled", "");
			const r = new HttpRequest("POST", "/v1/users/migrate", d);
			r.process(function (status: any, data: any) {
				self.refresh();
				new Ui.MessageBox(App.translations.Success, App.translations.SightingsImportSuccess, "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--success-color)", function () { });
				self.mPassword.container.innerHTML = "";
			}, function () {
				self.mImportButton.disabled = false;
				self.mImportButton.innerHTML = "Import";
				self.mEmail.container.removeAttribute("disabled");
				self.mPassword.container.removeAttribute("disabled");
				new Ui.MessageBox(App.translations.Error, App.translations.WrongImportCredentials, "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function () { });
			});
		});

		this.container.appendChild(this.mImportButton);

		const exportDataLbl = document.createElement("h2");
		exportDataLbl.innerHTML = App.translations.ExportMyData;
		this.container.appendChild(exportDataLbl);

		const exportDataDescription = document.createElement("p");
		exportDataDescription.innerHTML = App.translations.ExportMyDataDescription;
		this.container.appendChild(exportDataDescription);

		const exportDataButton = document.createElement("button");
		exportDataButton.innerHTML = App.translations.ExportMyData;
		exportDataButton.addEventListener("click", function () {
			exportDataButton.disabled = true;
			const r = new HttpRequest("GET", "/v1/users/export");
			r.process(function (status: any, data: any) {
				DownloadFile("data.json", JSON.stringify(data));
			}, function () {
				new Ui.MessageBox(App.translations.Error, App.translations.SomethingWentWrong, "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function () { });
			}, function () {
				exportDataButton.disabled = false;
			});
		});

		this.container.appendChild(exportDataButton);

		const deleteAccountLbl = document.createElement("h2");
		deleteAccountLbl.innerHTML = App.translations.PermanentlyDeleteAccount;
		this.container.appendChild(deleteAccountLbl);

		const deleteAccountDescription = document.createElement("p");
		deleteAccountDescription.innerHTML = App.translations.DeleteAccountDescription;
		this.container.appendChild(deleteAccountDescription);

		const deleteAccountPassword = new Ui.Input("password");
		const deleteAccountButton = document.createElement("button");

		deleteAccountPassword.style.position = "relative";
		deleteAccountPassword.setPlaceHolder(App.translations.Password);
		deleteAccountPassword.container.addEventListener("keydown", function (e) {
			if (e.key == "Enter") {
				e.preventDefault();
				deleteAccountButton.click();
			}
		});
		this.attach(deleteAccountPassword);

		deleteAccountButton.innerHTML = App.translations.DeleteAccount;
		deleteAccountButton.style.backgroundColor = "red";
		deleteAccountButton.addEventListener("click", function () {
			deleteAccountButton.disabled = true;
			deleteAccountPassword.container.setAttribute("disabled", "");
			const d = {
				password: deleteAccountPassword.value
			}
			const r = new HttpRequest("POST", "/v1/users/request_delete", d);
			r.process(function (status: any, data: any) {
				self.refresh();
				new Ui.MessageBox(App.translations.Success, App.translations.CheckEmailInstructions, "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--success-color)", function () { });
				deleteAccountPassword.container.innerHTML = "";
			}, function () {
				deleteAccountButton.disabled = false;
				deleteAccountPassword.container.removeAttribute("disabled");
				new Ui.MessageBox(App.translations.Error, App.translations.SomethingWentWrong, "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function () { });
			});
		});

		this.container.appendChild(deleteAccountButton);
	}

	public get id(): string {
		return "settings";
	}

	public requireSignIn(): boolean {
		return true;
	}

	public onBecomeVisible(): void {
		super.onBecomeVisible();
		this.refresh();
	}

	private refresh(): void {
		const self = this;
		this.mProfile.setImage("/img/portraits/" + App.user.portrait);
		const me = new HttpRequest("GET", "/v1/users/me");
		me.process(function (status: any, data: any) {
			const d = data as Responses.UsersMe;
			if (Utils.isValidDate(d.user.imported_at)) {
				self.mImportButton.disabled = true;
				self.mImportButton.innerHTML = "(" + App.translations.AlreadyImported + ")";
				self.mEmail.container.setAttribute("disabled", "");
				self.mPassword.container.setAttribute("disabled", "");
			}
		});
	}
}