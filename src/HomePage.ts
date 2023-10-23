/// <reference path="ui/Widget.ts"/>
/// <reference path="ui/Image.ts"/>
/// <reference path="ui/Input.ts"/>
/// <reference path="utils/Report.ts"/>
/// <reference path="Page.ts"/>
/// <reference path="App.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="SpeciesPopUp.ts"/>
/// <reference path="ReportPopUp.ts"/>
/// <reference path="ProfilePopUp.ts"/>
/// <reference path="EditReportPopUp.ts"/>
/// <reference path="responses/Users.ts"/>

class HomePage extends Page {
	private mTable: Ui.Table;
	private mEmail: Ui.Input;
	private mPassword: Ui.Input;
	private mProgressBoxes: Ui.Div;

	public get id(): string {
		return "home";
	}

	protected onCreateLoggedIn(): void {
		const self = this;

		this.addProfile();

		const title = document.createElement("h1");
		title.innerHTML = App.translations.WelcomeMessage;
		this.container.appendChild(title);

		const reportButton = document.createElement("button");
		reportButton.innerHTML = App.translations.SubmitSighting;
		reportButton.style.width = "auto";
		reportButton.style.cssFloat = "right";
		reportButton.style.borderRadius = "inherit";
		reportButton.style.backgroundColor = "var(--primary-color)";
		reportButton.addEventListener("click", function() {
			openReportPopUp(null, null, function() {
				self.refresh();
			});
		});
		this.container.appendChild(reportButton);

		this.mProgressBoxes = new Ui.Div();
		this.mProgressBoxes.style.position = "relative";
		this.attach(this.mProgressBoxes);

		const lastReportsLbl = document.createElement("h2");
		lastReportsLbl.innerHTML = App.translations.LastSightings;
		this.container.appendChild(lastReportsLbl);

		this.mTable = new Ui.Table();
		this.attach(this.mTable);

		//this.refresh();
	}

	public setVisible(visible: boolean): void {
		super.setVisible(visible);
		if (App.isLoggedIn()) {
			this.refresh();
		}
	}

	private refresh(): void {
		const self = this;

		const me = new HttpRequest("GET", "/v1/users/me");
		me.process(function(status: any, data: any) {
			const d = data as Responses.UsersMe;
			App.setUser(d.user);
			self.mProfile.setImage("/img/portraits/" + d.user.portrait);
			self.mProgressBoxes.clear();
			self.mProgressBoxes.container.innerHTML = "";

			d.list_reports.forEach(e => {
				const d = new Ui.Div();
				d.container.className = "progress-box";
				d.container.style.cursor = "pointer";
				d.container.addEventListener("click", function() {
					self.appListener.showPage("list", e.list);
				});
				self.mProgressBoxes.attach(d);
				if (e.list.color !== "") {
					d.container.style.background = "linear-gradient(" + e.list.color + ", " + e.list.color + "bf)";
				}

				const title = document.createElement("p");
				title.innerHTML = App.getTranslation(e.list.name).substring(0, 25);
				d.container.appendChild(title);

				const pd = new Ui.Div();
				pd.container.style.bottom = "4px";
				pd.container.style.top = "unset";
				pd.container.style.left = "0";
				pd.container.style.width = "100%";
				pd.container.style.padding = "0 4px";
				pd.container.style.boxSizing = "border-box";
				d.attach(pd);

				const uniqueSpeciesInReports = Utils.numberOfUniqueSpeciesInReports(e.reports);
				const pbt = document.createElement("p");
				pbt.innerHTML = uniqueSpeciesInReports.toString() + "/" + e.list.species.length.toString();
				pbt.style.textShadow = "1px 1px #0000005e";
				pbt.style.fontSize = "0.8rem";
				pbt.style.fontWeight = "bold";
				pbt.style.cssFloat = "right";
				pbt.style.lineHeight = "unset";
				pd.container.appendChild(pbt);

				const pb = new Ui.ProgressBar();
				pb.style.position = "relative";
				pb.setProgress(uniqueSpeciesInReports / e.list.species.length * 100);
				pd.attach(pb);
			});

			self.mTable.clear();
			d.last_reports.forEach(e => {
				const d = e as Models.Report;
				const species = e.species as Models.Species;
				const entry = document.createElement("div");
				entry.addEventListener("click", function() {
					openEditReportPopUp(e);
				});
				entry.className = "list-entry";

				const name = document.createElement("h3");
				name.innerHTML = App.getTranslation(species.name);
				name.style.cursor = "pointer";
				name.style.color = "var(--success-color)";
				name.style.width = "fit-content";
				entry.appendChild(name);
				name.addEventListener("click", function(event) {
					event.stopPropagation();
					openSpeciesPopUp(e.species);
				});

				const subName = document.createElement("p");
				subName.innerHTML = species.name.la;
				entry.appendChild(subName);

				const dateInterval = document.createElement("p");
				if (Utils.isValidDate(d.from)) {
					dateInterval.innerHTML = Utils.toYYYY_MM_DD_hh_mm(d.from);
				}
				if (Utils.isValidDate(d.until)) {
					dateInterval.innerHTML += " - " + Utils.toYYYY_MM_DD_hh_mm(d.until);
				}
				dateInterval.style.fontStyle = "unset";
				entry.appendChild(dateInterval);

				const location = document.createElement("p");
				if (d.zone) {
					location.innerHTML = App.getTranslation(d.zone.name);
				}
				location.style.fontStyle = "unset";
				entry.appendChild(location);

				if (d.comment.text !== "") {
					const comment = document.createElement("p");
					comment.style.backgroundColor = "#1d3944";
					comment.style.borderRadius = "0 12px 0 12px";
					comment.style.lineHeight = "unset";
					comment.style.padding = "8px";
					comment.style.width = "fit-content";
					comment.innerHTML = d.comment.text;
					entry.appendChild(comment);
				}

				if (e.tags.length > 0) {
					const tagsContainer = document.createElement("div");
					tagsContainer.style.position = "relative";
					entry.appendChild(tagsContainer);

					e.tags.forEach(e => {
						const tag = document.createElement("div");
						tag.className = "tag";
						if (e.color !== "") {
							tag.style.backgroundColor = e.color;
						}
						const c = document.createElement("p");
						c.innerHTML = e.name;
						tag.appendChild(c);
						tagsContainer.appendChild(tag);
					});
				}

				self.mTable.addRow([entry]);
			});

			if (d.last_reports.length === 0) {
				const noRecent = document.createElement("p");
				noRecent.innerHTML = App.translations.NoRecentSightings;
				self.mTable.addRow([noRecent]);
			}

		}, function(status: any, data: Responses.Error, error: any) {
			console.log(data);
		});
	}

	public setData(data: any): void {
		super.setData(data);

		if (data && data.email) {
			if (this.mEmail) {
				this.mEmail.value = data.email;
				new Ui.MessageBox("Welcome " + data.display_name, "Your account has been successfully created, and you can now proceed to sign in!", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--success-color)", function() {});
			}
		}
	}

	protected onCreateLoggedOut(): void {
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
		logo.classList.add("shadow");
		content.container.appendChild(logo);

		const title = document.createElement("h1");
		title.innerHTML = App.translations.SignIn;
		this.container.appendChild(title);

		this.mEmail = new Ui.Input("email");
		this.mPassword = new Ui.Input("password");
		const signIn = document.createElement("button");

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
				signIn.click();
			}
		});
		this.attach(this.mPassword);

		const staySignedIn = new Ui.Input("checkbox");
		staySignedIn.style.position = "relative";
		staySignedIn.container.setAttribute("name", "stay_signed_in");
		staySignedIn.container.id = "stay_signed_in";
		this.attach(staySignedIn);

		const staySignedInLbl = document.createElement("label");
		staySignedInLbl.setAttribute("for", staySignedIn.container.id);
		staySignedInLbl.innerHTML = App.translations.StaySignedIn;
		this.container.appendChild(staySignedInLbl);

		const help = document.createElement("p");
		help.className = "details";
		help.innerHTML = App.translations.MarkSignInCheckbox;
		this.container.appendChild(help);

		signIn.innerHTML = App.translations.SignIn;
		signIn.addEventListener("click", function() {
			const authorization = 'Basic ' + btoa(self.mEmail.value + ':' + self.mPassword.value);
			const stayLoggedIn = (staySignedIn.container as HTMLInputElement).checked;
			const content = {
				'refresh-token': ''
			};

			const r = new HttpRequest("POST", "/v1/auth", content);
			r.setHeader("Authorization", authorization);
			r.process(function(status: any, data: any) {
				setCookie("x300sf_at", data.access_token, stayLoggedIn ? null : data.expires_at);
				//setCookie("x300sf_rt", data.refresh_token, 3600 * 24 * 60); // 60 days.
				const url = new URL(window.location.href);
				const redirectParam = url.searchParams.get("redirect");
				if (redirectParam) {
					window.open(redirectParam, "_self");
				} else {
					location.reload();
				}
			}, function(status: any, data: Responses.Error, err: any) {
				if (data.code === Responses.ErrorCode.ErrorAccountDeactivated) {
					new Ui.MessageBox("Oh, no!", "Your account has been deactivated.", "img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function() {});
				} else {
					new Ui.MessageBox("Oh, no!", "It appears that the <b>email</b> or <b>password</b> you entered is invalid.", "img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function() {});
				}
			});
		});

		this.container.appendChild(signIn);

		const hr = document.createElement("hr");
		this.container.appendChild(hr);

		const self = this;
		const noAccountYet = document.createElement("p");
		noAccountYet.innerHTML = App.translations.NoAccountYet;
		noAccountYet.addEventListener("click", function() {
			self.appListener.showPage("sign-up", null);
		});
		this.container.appendChild(noAccountYet);

		const forgot = document.createElement("a");
		forgot.innerHTML = App.translations.ForgotYourPassword;
		forgot.style.margin = "var(--small-distance)";
		forgot.addEventListener("click", function() {
			self.appListener.showPage("forgot", null);
		});
		this.container.appendChild(forgot);
	}
}
