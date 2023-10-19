/// <reference path="utils/Date.ts"/>
/// <reference path="Page.ts"/>
/// <reference path="ui/Input.ts"/>

class SettingsPage extends Page {
	private mEmail: Ui.Input;
	private mPassword: Ui.Input;
	private mImportButton: HTMLButtonElement;

	constructor(appListener: IAppListener) {
		super(appListener);

		const self = this;

		this.addProfile();

		const title = document.createElement("h1");
		title.innerHTML = App.translations.Settings;
		this.container.appendChild(title);

		const importSubmissionsLbl = document.createElement("h2");
		importSubmissionsLbl.innerHTML = App.translations.ImportSightings;
		this.container.appendChild(importSubmissionsLbl);

		const importSubmissionsDescription = document.createElement("p");
		importSubmissionsDescription.innerHTML = App.translations.ImportSightingsDescription;
		this.container.appendChild(importSubmissionsDescription);

		this.mEmail = new Ui.Input("email");
		this.mPassword = new Ui.Input("password");
		this.mImportButton = document.createElement("button");

		this.mEmail.style.position = "relative";
		this.mEmail.setPlaceHolder(App.translations.Email);
		this.mEmail.container.addEventListener("keydown", function(e) {
			if (e.key == "Enter") {
				e.preventDefault();
				self.mPassword.select();
			}
		});
		this.attach(this.mEmail);

		this.mPassword.style.position = "relative";
		this.mPassword.setPlaceHolder(App.translations.Password);
		this.mPassword.container.addEventListener("keydown", function(e) {
			if (e.key == "Enter") {
				e.preventDefault();
				self.mImportButton.click();
			}
		});
		this.attach(this.mPassword);

		this.mImportButton.innerHTML = "Import";
		this.mImportButton.addEventListener("click", function() {
			const authorization = 'Basic ' + btoa(self.mEmail.value + ':' + self.mPassword.value);
			const d = {
				old_password: self.mPassword.value,
				old_username: self.mEmail.value
			};

			self.mImportButton.disabled = true;
			self.mImportButton.innerHTML = "Please wait...";
			self.mEmail.container.setAttribute("disabled", "");
			self.mPassword.container.setAttribute("disabled", "");
			const r = new HttpRequest("POST", "/v1/users/migrate", JSON.stringify(d));
			r.setHeader("Authorization", authorization);
			r.process(function(status: any, data: any) {
				self.refresh();
				new Ui.MessageBox("Success", "All your submissions have successfully been imported!", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--success-color)", function() {});
				self.mPassword.container.innerHTML = "";
			}, function() {
				self.mImportButton.disabled = false;
				self.mImportButton.innerHTML = "Import";
				self.mEmail.container.removeAttribute("disabled");
				self.mPassword.container.removeAttribute("disabled");
				new Ui.MessageBox("Oh, no!", "It appears that the <b>email</b> or <b>password</b> you entered is invalid.", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function() {});
			});
		});

		this.container.appendChild(this.mImportButton);

		const deleteAccountLbl = document.createElement("h2");
		deleteAccountLbl.innerHTML = "Delete Account";
		this.container.appendChild(deleteAccountLbl);

		const deleteAccountDescription = document.createElement("p");
		deleteAccountDescription.innerHTML = "Leaving us?";
		this.container.appendChild(deleteAccountDescription);

		const deleteAccountPassword = new Ui.Input("password");
		const deleteAccountButton = document.createElement("button");

		deleteAccountPassword.style.position = "relative";
		deleteAccountPassword.setPlaceHolder(App.translations.Password);
		deleteAccountPassword.container.addEventListener("keydown", function(e) {
			if (e.key == "Enter") {
				e.preventDefault();
				deleteAccountButton.click();
			}
		});
		this.attach(deleteAccountPassword);

		deleteAccountButton.innerHTML = "Delete Account";
		deleteAccountButton.style.backgroundColor = "red";
		deleteAccountButton.addEventListener("click", function() {
			deleteAccountButton.disabled = true;
			deleteAccountPassword.container.setAttribute("disabled", "");
			const d = {
				password: deleteAccountPassword.value
			}
			const r = new HttpRequest("POST", "/v1/users/request_delete", JSON.stringify(d));
			r.process(function(status: any, data: any) {
				self.refresh();
				new Ui.MessageBox("All right", "Check your email inbox for further instructions.", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--success-color)", function() {});
				deleteAccountPassword.container.innerHTML = "";
			}, function() {
				deleteAccountButton.disabled = false;
				deleteAccountPassword.container.removeAttribute("disabled");
				new Ui.MessageBox("Oh, no!", "Something went wrong.", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function() {});
			});
		});

		this.container.appendChild(deleteAccountButton);
	}

	public get id(): string {
		return "settings";
	}

	public onBecomeVisible(): void {
		super.onBecomeVisible();
		this.refresh();
	}

	private refresh(): void {
		const self = this;
		this.mProfile.setImage("/img/portraits/" + App.user.portrait);
		const me = new HttpRequest("GET", "/v1/users/me");
		me.process(function(status: any, data: any) {
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