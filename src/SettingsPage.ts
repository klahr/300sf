/// <reference path="utils/Date.ts"/>

class SettingsPage extends Page {
    private mContent: Ui.Div;
    private mEmail: Ui.Input;
    private mPassword: Ui.Input;
    private mImportButton: HTMLButtonElement;

    constructor(appListener: IAppListener) {
        super(appListener);

        const self = this;
    
		this.mContent = new Ui.Div();
		this.mContent.style.position = "relative";
		this.mContent.style.width = "100%";
		this.attach(this.mContent);

		const profile = new Ui.Image("/img/99249498-452f-4b0b-a1f5-45902f7ffe48.png", "48px", "48px");
        profile.addClass("portrait");
		profile.container.addEventListener("click", function() {
		});
        this.attach(profile);

        const title = document.createElement("h1");
		title.innerHTML = App.translations.Settings;
		this.container.appendChild(title);

        const importSubmissionsLbl = document.createElement("h2");
		importSubmissionsLbl.innerHTML = "Import submissions";
		this.container.appendChild(importSubmissionsLbl);

        const importSubmissionsDescription = document.createElement("p");
		importSubmissionsDescription.innerHTML = "If you possess any submittals within the previous version of the 300 Swedish Birds application, you have the ability to import them.";
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
    }

    public get id(): string {
        return "settings";
    }

    public setVisible(visible: boolean): void {
        super.setVisible(visible);
        if (visible) {
            this.refresh();
        }
    }

    private refresh(): void {
        const self = this;
        const me = new HttpRequest("GET", "/v1/users/me");
		me.process(function(status: any, data: any) {
			const d = data as Responses.UsersMe;
            if (Utils.isValidDate(d.user.imported_at)) {
                self.mImportButton.disabled = true;
                self.mImportButton.innerHTML = "(Already imported)";
                self.mEmail.container.setAttribute("disabled", "");
                self.mPassword.container.setAttribute("disabled", "");
            }
        });
    }
}