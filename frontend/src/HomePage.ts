/// <reference path="ui/Widget.ts"/>
/// <reference path="ui/Image.ts"/>
/// <reference path="Overview.ts"/>
/// <reference path="CompetitionSheet.ts"/>
/// <reference path="Onboarding.ts"/>
/// <reference path="AnalyzePopUp.ts"/>
/// <reference path="ui/Input.ts"/>
/// <reference path="ui/SpeciesSearch.ts"/>
/// <reference path="utils/Report.ts"/>
/// <reference path="Page.ts"/>
/// <reference path="App.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="SpeciesPopUp.ts"/>
/// <reference path="ReportPopUp.ts"/>
/// <reference path="ProfilePopUp.ts"/>
/// <reference path="EditReportPopUp.ts"/>
/// <reference path="responses/Users.ts"/>
/// <reference path="responses/Auth.ts"/>
/// <reference path="responses/Response.ts"/>

class HomePage extends Page {
	private mTable: Ui.Table;
	private mEmail: Ui.Input;
	private mPassword: Ui.Input;
	private mProgressBoxes: Ui.Div;
	private mSpecies: Array<Models.Species>;
	private mSpeciesList: Ui.SpeciesSearch;
	private mPreferred = new Set<string>();
	private mSwedish = new Set<string>();
	private mAnnouncements: Ui.Div;
	private mTitle: HTMLHeadingElement;
	private mStats: HTMLDivElement;
	private mCompetitions: HTMLDivElement;

	public get id(): string {
		return "home";
	}

	protected onCreateLoggedIn(): void {
		const self = this;
		this.addClass("wide");

		this.addProfile();

		const hero = document.createElement("div");
		hero.className = "hero";
		this.container.appendChild(hero);

		this.mTitle = document.createElement("h1");
		this.mTitle.innerHTML = App.translations.WelcomeMessage;
		hero.appendChild(this.mTitle);

		const today = document.createElement("p");
		today.className = "hero-date";
		const now = new Date();
		today.textContent = now.toLocaleDateString(App.language, { weekday: "long", day: "numeric", month: "long" });
		hero.appendChild(today);

		this.mAnnouncements = new Ui.Div();
		this.mAnnouncements.addClass("announcements");
		this.attach(this.mAnnouncements);

		this.mSpeciesList = new Ui.SpeciesSearch(App.translations.SearchSpecies, (species: Models.Species) => {
			openSpeciesPopUp(species, null, () => { });
		}, false, "", () => {
			openIdentifyBySound((species: Models.Species) => {
				openSpeciesPopUp(species, null, () => { });
			});
		});
		this.attach(this.mSpeciesList);

		this.mStats = document.createElement("div");
		this.mStats.className = "stats";
		this.container.appendChild(this.mStats);
		const statsPlaceholder = document.createElement("div");
		statsPlaceholder.className = "stat-row";
		for (let i = 0; i < 3; ++i) {
			const placeholder = document.createElement("div");
			placeholder.className = "stat-tile skeleton";
			statsPlaceholder.appendChild(placeholder);
		}
		this.mStats.appendChild(statsPlaceholder);

		const listsLbl = document.createElement("h2");
		listsLbl.innerHTML = App.translations.YourChecklists;
		this.container.appendChild(listsLbl);

		this.mProgressBoxes = new Ui.Div();
		this.mProgressBoxes.addClass("list-tiles");
		this.attach(this.mProgressBoxes);
		for (let i = 0; i < 3; ++i) {
			const placeholder = document.createElement("div");
			placeholder.className = "list-tile skeleton";
			this.mProgressBoxes.container.appendChild(placeholder);
		}

		// Running competitions and invitations; hidden when there are none.
		this.mCompetitions = document.createElement("div");
		this.mCompetitions.className = "home-competitions hidden";
		this.container.appendChild(this.mCompetitions);

		const lastReportsLbl = document.createElement("h2");
		lastReportsLbl.innerHTML = App.translations.LastSightings;
		this.container.appendChild(lastReportsLbl);

		this.mTable = new Ui.Table();
		this.mTable.addClass("sightings");
		this.attach(this.mTable);
	}

	public setVisible(visible: boolean): void {
		super.setVisible(visible);
		if (visible && App.isLoggedIn()) {
			this.refresh();
		}
	}

	private refreshCompetitions(): void {
		const self = this;
		new HttpRequest("GET", "/v1/competitions/list/me").process(function (status: any, data: any) {
			const current = ((data.result || []) as Array<Models.Competition>).filter(c => c.status !== "finished" && (c.joined || c.invited)).slice(0, 3);
			self.mCompetitions.innerHTML = "";
			self.mCompetitions.classList.toggle("hidden", current.length === 0);
			if (current.length === 0) {
				return;
			}
			const h = document.createElement("h2");
			h.textContent = App.translations.YourCompetitions;
			self.mCompetitions.appendChild(h);
			current.forEach(c => {
				const row = document.createElement("button");
				row.className = "home-competition";
				row.style.setProperty("--competition-color", c.color || "var(--primary)");
				const icon = document.createElement("span");
				icon.className = "material-symbols-outlined";
				icon.textContent = c.invited ? "mail" : "emoji_events";
				const text = document.createElement("span");
				text.className = "home-competition-text";
				const name = document.createElement("b");
				name.textContent = App.getTranslation(c.name);
				const detail = document.createElement("small");
				detail.textContent = c.invited ? App.translations.Invitation : Competitions.statusLabel(c) + " · " + Competitions.period(c);
				text.appendChild(name);
				text.appendChild(detail);
				const place = document.createElement("span");
				place.className = "home-competition-place";
				place.textContent = c.joined && c.my_count > 0 ? "#" + c.my_rank + " · " + App.translations.NumberOfSpecies.replace("{0}", String(c.my_count)) : "";
				row.appendChild(icon);
				row.appendChild(text);
				row.appendChild(place);
				row.addEventListener("click", function () {
					Competitions.open(c.uuid, () => self.refreshCompetitions());
				});
				self.mCompetitions.appendChild(row);
			});
		}, function () { });
	}

	private refresh(): void {
		const self = this;
		this.refreshCompetitions();

		const me = new HttpRequest("GET", "/v1/users/me");
		me.process(function (status: any, data: any) {
			const d = data as Responses.UsersMe;
			// The server sends null rather than [] when there is nothing yet.
			d.list_reports = d.list_reports || [];
			d.last_reports = d.last_reports || [];
			App.setUser(d.user);
			self.mProfile.setImage("/img/portraits/" + d.user.portrait);
			self.mProgressBoxes.clear();
			self.mProgressBoxes.container.innerHTML = "";

			// Species on the user's lists come first in search.
			self.mPreferred = new Set<string>();
			d.list_reports.forEach(e => {
				e.list.species.forEach(s => self.mPreferred.add(s.uuid));
			});
			if (self.mSpecies) {
				self.mSpeciesList.setSpecies(self.mSpecies, self.mPreferred, self.mSwedish);
			}
			d.list_reports.forEach(e => {
				self.mProgressBoxes.container.appendChild(Overview.listTile(e, function () {
					self.appListener.showPage("list", e.list);
				}));
			});
			if (d.list_reports.length === 0) {
				self.mProgressBoxes.container.appendChild(Onboarding.starterCard(function (uuid: string) {
					self.appListener.showPage("list", { uuid: uuid });
				}));
			}

			self.mStats.innerHTML = "";
			self.mStats.appendChild(Overview.statRow(d.list_reports, d.last_reports));
			self.mTitle.textContent = App.translations.Greeting.replace("{0}", d.user.display_name);

			self.mTable.clear();
			d.last_reports.forEach(e => {
				const d = e as Models.Report;
				const species = e.species as Models.Species;
				const entry = document.createElement("div");
				entry.addEventListener("click", function () {
					openEditReportPopUp(e);
				});
				entry.className = "list-entry";

				const name = document.createElement("h3");
				name.textContent = App.getTranslation(species.name);
				name.classList.add("species-name");
				entry.appendChild(name);
				name.addEventListener("click", function (event) {
					event.stopPropagation();
					openSpeciesPopUp(e.species, null, () => { });
				});

				const subName = document.createElement("p");
				subName.textContent = species.name.la;
				entry.appendChild(subName);

				const dateInterval = document.createElement("p");
				dateInterval.className = "meta date";
				if (Utils.isValidDate(d.from)) {
					dateInterval.innerHTML = Utils.toYYYY_MM_DD_hh_mm(d.from);
				}
				if (Utils.isValidDate(d.until)) {
					dateInterval.innerHTML += " - " + Utils.toYYYY_MM_DD_hh_mm(d.until);
				}
				entry.appendChild(dateInterval);

				const location = document.createElement("p");
				location.className = "meta place";
				if (d.zone) {
					location.textContent = App.getTranslation(d.zone.name);
				}
				entry.appendChild(location);

				if (d.comment.text !== "") {
					const comment = document.createElement("p");
					comment.classList.add("comment");
					comment.textContent = d.comment.text;
					entry.appendChild(comment);
				}

				if (e.tags.length > 0) {
					const tagsContainer = document.createElement("div");
					entry.appendChild(tagsContainer);

					e.tags.forEach(e => {
						const tag = document.createElement("div");
						tag.className = "tag";
						if (e.color !== "") {
							tag.style.setProperty("--tag", e.color);
						}
						const c = document.createElement("p");
						c.textContent = e.name;
						tag.appendChild(c);
						tagsContainer.appendChild(tag);
					});
				}

				self.mTable.addRow([entry]);
			});

			if (d.last_reports.length === 0) {
				const noRecent = document.createElement("div");
				noRecent.className = "empty-state";
				noRecent.innerHTML = "<span class='material-symbols-outlined'>raven</span>";
				const text = document.createElement("p");
				text.innerHTML = App.translations.NoRecentSightings;
				noRecent.appendChild(text);
				self.mTable.addRow([noRecent]);
			}
		}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
		});

		// Every species on a list template is a Swedish one; those rank above the rest in search.
		if (self.mSwedish.size === 0) {
			new HttpRequest("GET", "/v1/list_templates/list").process(function (status: any, data: any) {
				(data.result || []).forEach(t => (t.species || []).forEach(s => self.mSwedish.add(s.uuid)));
				if (self.mSpecies) {
					self.mSpeciesList.setSpecies(self.mSpecies, self.mPreferred, self.mSwedish);
				}
			});
		}

		const species = new HttpRequest("GET", "/v1/species/list", null);
		species.process((status: any, data: any) => {
			self.mSpecies = data.result || [];
			self.mSpeciesList.setSpecies(self.mSpecies, self.mPreferred, self.mSwedish);
		}, (status: any, data: Responses.GeneralErrorResponse, error: any) => {
		}, undefined);

		const announcements = new HttpRequest("GET", "/v1/announcements/list", null);
		announcements.process((status: any, data: any) => {
			const an = data.data as Array<Models.Announcement>;
			self.mAnnouncements.clear();
			an.forEach(e => {
				const a = document.createElement("div");
				a.className = "announcement";
				self.mAnnouncements.container.appendChild(a);
				const img = document.createElement("p");
				img.className = "announcement-icon";
				img.innerHTML = "<span class='material-symbols-outlined'>" + Utils.escapeHtml(e.icon) + "</span>";
				if (e.color) {
					a.style.setProperty("--tone", e.color);
				}

				a.appendChild(img);

				const close = document.createElement("span");
				close.className = "material-symbols-outlined close";
				close.setAttribute("role", "button");
				close.setAttribute("aria-label", App.translations.Close);
				close.innerHTML = "close";
				a.appendChild(close);
				close.addEventListener("click", function () {
					const r = new HttpRequest("PUT", "/v1/announcements/" + e.uuid + "/read");
					r.process((status: any, data: any) => {
						a.classList.add("fade-out");
						setTimeout(function () {
							a.remove();
						}, 500);
					});
				});

				const c = document.createElement("p");
				c.innerHTML = App.getTranslation(e.content);
				a.appendChild(c);
			});
		}, (status: any, data: Responses.GeneralErrorResponse, error: any) => {
		}, undefined);
	}

	public setData(data: any): void {
		super.setData(data);

		if (data) {
			if (data.email) {
				if (this.mEmail) {
					this.mEmail.value = data.email;
					new Ui.MessageBox(App.translations.Welcome + " " + data.display_name, App.translations.ProceedToSignIn, "/img/messagebox-256.webp", "var(--success-color)", function () { });
				}
			}
		}
	}

	protected onCreateLoggedOut(): void {
		this.addClass("auth-page");
		const content = new Ui.Div();
		content.style.position = "relative";
		content.style.width = "100%";
		content.style.textAlign = "center";
		this.attach(content);
		const logo = document.createElement("img");
		logo.src = "/img/birds-192.webp";
		logo.alt = "";
		logo.width = 192;
		logo.height = 192;
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
				signIn.click();
			}
		});
		this.attach(this.mPassword);

		const staySignedInRow = document.createElement("div");
		staySignedInRow.className = "check-row";
		this.container.appendChild(staySignedInRow);

		const staySignedIn = new Ui.Input("checkbox");
		staySignedIn.container.setAttribute("name", "stay_signed_in");
		staySignedIn.container.id = "stay_signed_in";
		this.mChildren.push(staySignedIn);
		staySignedInRow.appendChild(staySignedIn.container);

		const staySignedInLbl = document.createElement("label");
		staySignedInLbl.setAttribute("for", staySignedIn.container.id);
		staySignedInLbl.innerHTML = App.translations.StaySignedIn;
		staySignedInRow.appendChild(staySignedInLbl);

		const help = document.createElement("p");
		help.className = "details";
		help.innerHTML = App.translations.MarkSignInCheckbox;
		this.container.appendChild(help);

		signIn.innerHTML = App.translations.SignIn;
		signIn.addEventListener("click", function () {
			const stayLoggedIn = (staySignedIn.container as HTMLInputElement).checked;

			const d = {
				grant_type: "password",
				email: self.mEmail.value,
				password: self.mPassword.value,
				scopes: ["user", "admin"],
				stay_signed_in: stayLoggedIn
			}
			Ui.setBusy(signIn, true);
			const r = new HttpRequest("POST", "/v1/auth", d);
			r.setNoRetry();
			r.process(function (status: any, data: Responses.AuthenticationResponse) {
				Session.store(data);
				const url = new URL(window.location.href);
				const redirectParam = Utils.safeRedirect(url.searchParams.get("redirect"));
				if (redirectParam) {
					window.open(redirectParam, "_self");
				} else {
					location.reload();
				}
			}, function (status: any, data: Responses.GeneralErrorResponse, err: any) {
				Ui.setBusy(signIn, false);
				if (data && data.code === Responses.ErrorCode.ErrorAccountDeactivated) {
					new Ui.MessageBox(App.translations.Error, App.translations.AccountDeactivated, "/img/messagebox-256.webp", "var(--error-color)", function () { });
				} else {
					new Ui.MessageBox(App.translations.Error, App.translations.InvalidEmailOrPassword, "/img/messagebox-256.webp", "var(--error-color)", function () { });
				}
			}, function () {
				// Network errors skip the handler above; on success the page reloads anyway.
				Ui.setBusy(signIn, false);
			});
		});

		this.container.appendChild(signIn);

		const hr = document.createElement("hr");
		this.container.appendChild(hr);

		const self = this;
		const noAccountYet = document.createElement("p");
		noAccountYet.innerHTML = App.translations.NoAccountYet;
		this.container.appendChild(noAccountYet);

		const forgot = document.createElement("a");
		forgot.innerHTML = App.translations.ForgotYourPassword;
		forgot.href = "/forgot";
		const forgotP = document.createElement("p");
		forgotP.appendChild(forgot);
		this.container.appendChild(forgotP);
	}
}
