/// <reference path="ui/Widget.ts"/>
/// <reference path="ui/Image.ts"/>
/// <reference path="ui/Input.ts"/>
/// <reference path="Page.ts"/>
/// <reference path="App.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="SpeciesPopUp.ts"/>

class HomePage extends Page {
	constructor(appListener: IAppListener) {
		super(appListener);
	}

	public get id(): string {
		return "home-page";
	}

	protected onCreateLoggedIn(): void {
		const self = this;

		const profile = new Ui.Image("img/99249498-452f-4b0b-a1f5-45902f7ffe48.png", "48px", "48px");
		profile.style.left = "unset";
		profile.style.right = "var(--medium-distance)";
		profile.style.top = "var(--small-distance)";
		profile.style.borderRadius = "100%";
		profile.container.addEventListener("click", function() {
			const p = document.createElement("p");
			p.innerHTML = "hej";
			new Ui.PopUp(p, function() {
			});
		});
		this.attach(profile);

		const title = document.createElement("h1");
		title.innerHTML = App.translations.WelcomeMessage;
		this.container.appendChild(title);

		const reportButton = document.createElement("button");
		reportButton.innerHTML = "Submit";
		reportButton.style.width = "auto";
		reportButton.style.cssFloat = "right";
		reportButton.style.borderRadius = "inherit";
		reportButton.style.backgroundColor = "var(--success-color)";
		reportButton.addEventListener("click", function() {
			openReportPopup(null, null);
		});
		this.container.appendChild(reportButton);

		const table = new Ui.Table();
		this.attach(table);

		const r = new HttpRequest("GET", "/v1/reports/list");
		r.process(function(status: any, data: any) {
			data.result.forEach(e => {
				const d = e as Models.Report;
				const species = e.species as Models.Species;
				const entry = document.createElement("div");
				entry.className = "list-entry";
	
				const name = document.createElement("h3");
				name.innerHTML = App.getTranslation(species.name);
				name.style.cursor = "pointer";
				entry.appendChild(name);
				name.addEventListener("click", function() {
				});
	
				const subName = document.createElement("p");
				subName.innerHTML = species.name.la;
				entry.appendChild(subName);

				const observationDate = document.createElement("p");
				observationDate.innerHTML = new Date(d.from).toLocaleString();
				entry.appendChild(observationDate);

				table.addRow([entry]);
			});
		});
	}
	
	protected onCreateLoggedOut(): void {
		const content = new Ui.Div();
		content.style.position = "relative";
		content.style.width = "100%";
		content.style.textAlign = "center";
		this.attach(content);
		const logo = document.createElement("img");
		logo.src = "img/99249498-452f-4b0b-a1f5-45902f7ffe48.png";
		logo.alt = "TODO";
		logo.style.width = "192px";
		logo.style.borderRadius = "50%";
		content.container.appendChild(logo);

		const title = document.createElement("h1");
		title.innerHTML = App.translations.SignIn;
		this.container.appendChild(title);

		const email = new Ui.Input("email");
		const password = new Ui.Input("password");
		const signIn = document.createElement("button");

		email.style.position = "relative";
		email.setPlaceHolder(App.translations.Email);
		email.container.addEventListener("keydown", function(e) {
			if (e.key == "Enter") {
				e.preventDefault();
				password.select();
			}
		});
		this.attach(email);

		password.style.position = "relative";
		password.setPlaceHolder(App.translations.Password);
		password.container.addEventListener("keydown", function(e) {
			if (e.key == "Enter") {
				e.preventDefault();
				signIn.click();
			}
		});
		this.attach(password);

		const staySignedIn = new Ui.Input("checkbox");
		staySignedIn.style.position = "relative";
		staySignedIn.container.setAttribute("name", "stay_signed_in");
		this.attach(staySignedIn);

		const staySignedInLbl = document.createElement("label");
		staySignedInLbl.setAttribute("for", staySignedIn.container.getAttribute("name")!);
		staySignedInLbl.innerHTML = App.translations.StaySignedIn;
		this.container.appendChild(staySignedInLbl);

		const help = document.createElement("p");
		help.className = "details";
		help.innerHTML = App.translations.MarkSignInCheckbox;
		this.container.appendChild(help);

		signIn.innerHTML = App.translations.SignIn;
		signIn.addEventListener("click", function() {
			const authorization = 'Basic ' + btoa(email.value + ':' + password.value);
			const stayLoggedIn = (staySignedIn.container as HTMLInputElement).checked;
			const content = {
				'refresh-token': ''
			};

			const r = new HttpRequest("POST", "/v1/auth", JSON.stringify(content));
			r.setHeader("Authorization", authorization);
			r.process(function(status: any, data: any) {
				// TODO:
				setCookie("x300sf_at", data.access_token, stayLoggedIn ? null : data.expires_at);
				//setCookie("x300sf_rt", data.refresh_token, 3600 * 24 * 60); // 60 days.
				location.reload();
			});
		});

		this.container.appendChild(signIn);

		const hr = document.createElement("hr");
		this.container.appendChild(hr);

		const self = this;
		const noAccountYet = document.createElement("p");
		noAccountYet.innerHTML = App.translations.NoAccountYet;
		noAccountYet.addEventListener("click", function() {
			self.appListener.showPage("sign-up-page", null);
		});
		this.container.appendChild(noAccountYet);

		const forgot = document.createElement("a");
		forgot.innerHTML = App.translations.ForgotYourPassword;
		forgot.addEventListener("click", function() {
			self.appListener.showPage("forgot-page", null);
		});
		this.container.appendChild(forgot);
	}
}
