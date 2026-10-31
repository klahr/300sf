/// <reference path="ui/LayerManager.ts"/>
/// <reference path="HomePage.ts"/>
/// <reference path="SignUpPage.ts"/>
/// <reference path="ListPage.ts"/>
/// <reference path="SettingsPage.ts"/>
/// <reference path="DonationPage.ts"/>
/// <reference path="ForgotPage.ts"/>
/// <reference path="AdminPage.ts"/>
/// <reference path="ManageTagsPage.ts"/>
/// <reference path="ReportsPage.ts"/>
/// <reference path="MainMenu.ts"/>
/// <reference path="ManageChecklistsPage.ts"/>
/// <reference path="PendingReport.ts"/>
/// <reference path="ManageCompetitionsPage.ts"/>
/// <reference path="ContributePage.ts"/>
/// <reference path="SetPasswordPage.ts"/>
/// <reference path="DeleteAccountPage.ts"/>
/// <reference path="ContactPage.ts"/>
/// <reference path="ExperimentalPage.ts"/>
/// <reference path="Translations.ts"/>
/// <reference path="models/Translation.ts"/>
/// <reference path="models/User.ts"/>
/// <reference path="LocalStorage.ts"/>
/// <reference path="Session.ts"/>
/// <reference path="Outbox.ts"/>
/// <reference path="Updates.ts"/>
/// <reference path="NewsPage.ts"/>
/// <reference path="News.ts"/>

interface IAppListener {
	showPage(id: string, data: any): void;
	logOut(): void;
}

class App implements IAppListener {
	private mLayerManager = new Ui.LayerManager();
	private mMainMenu = new MainMenu(this);
	private mCurrentPage = "home";
	private mCurrentData: any = null;
	private static sLanguage: string = "en";
	public static translations = Translations.en;
	private mPages = new Array<Page>();
	private static sUser: Models.User;
	private static sVersion = "dev";
	// Where to get the source of what runs here (site.source-url), which the AGPL asks us to offer.
	public static sourceUrl = "https://github.com/klahr/300sf";

	constructor(language: string, redirect: string, queryParameters: object, version: string) {
		App.sVersion = version || "dev";
		App.hideIconNames();

		Session.forgetLegacyTokens();
		// Pages read the signed-in state when they are built, so renew an expired
		// access token first.
		Session.refreshSync();

		const ks = new LocalStorage();
		const lang = ks.get("lang") as string;
		App.setLanguage(lang !== null && lang !== "" ? lang : language);

		// Before any page or the navigation is built: they need the user, and are written in the
		// user's language, which this sets.
		if (App.isLoggedIn()) {
			this.fetchUser();
		}

		this.mLayerManager.addLayer("base");
		this.mLayerManager.showLayer("base");

		this.addPage(new HomePage(this));
		this.addPage(new AdminPage(this));
		this.addPage(new SignUpPage(this));
		this.addPage(new ForgotPage(this));
		this.addPage(new ListPage(this));
		this.addPage(new ContributePage(this));
		this.addPage(new DonationPage(this));
		this.addPage(new SettingsPage(this));
		this.addPage(new ManageCheckListsPage(this));
		this.addPage(new ManageTagsPage(this));
		this.addPage(new ReportsPage(this));
		this.addPage(new SetPasswordPage(this));
		this.addPage(new DeleteAccountPage(this));
		this.addPage(new ContactPage(this));
		this.addPage(new ManageCompetitionsPage(this));
		this.addPage(new ExperimentalPage(this));
		this.addPage(new NewsPage(this));

		document.body.classList.toggle("signed-out", !App.isLoggedIn());
		this.mLayerManager.addToLayer("base", this.mMainMenu);

		this.showPage(redirect === "" ? "home" : redirect, queryParameters);
		if (App.isLoggedIn()) {
			PendingReport.resume();
		}

		// Back and forward: close the newest dialog, or show the page for the URL without a reload.
		const self = this;

		// A sighting was reported, from any form: show the current page again so it includes it.
		window.addEventListener("sightingadded", function () {
			self.showPage(self.mCurrentPage, self.mCurrentData);
		});
		// Sightings saved while offline, from this visit or an earlier one.
		Outbox.send();
		Updates.watch();
		News.refresh(true);

		window.addEventListener("popstate", function () {
			if (Ui.PopUp.handleBack()) {
				return;
			}
			self.showPageForLocation();
		});
	}

	// Icons are ligatures, so a screen reader would read "volunteer_activism" before "Donation". Hide
	// them from assistive technology as they are added; buttons with only an icon carry aria-label.
	private static hideIconNames(): void {
		const hide = function (root: ParentNode) {
			root.querySelectorAll(".material-symbols-outlined:not([aria-hidden])").forEach(e => e.setAttribute("aria-hidden", "true"));
		};
		new MutationObserver(function (mutations) {
			mutations.forEach(m => m.addedNodes.forEach(n => {
				if (n instanceof Element) {
					if (n.classList.contains("material-symbols-outlined")) {
						n.setAttribute("aria-hidden", "true");
					}
					hide(n);
				}
			}));
		}).observe(document.body, { childList: true, subtree: true });
	}

	// The page for the current URL: /list/<uuid> or /<page id>, with the query string as its data.
	private showPageForLocation(): void {
		const parts = window.location.pathname.split("/").filter(p => p !== "");
		const query: { [key: string]: string } = {};
		new URLSearchParams(window.location.search).forEach((value, key) => {
			query[key] = value;
		});
		if (parts[0] === "list" && parts[1]) {
			this.showPage("list", { uuid: parts[1] });
			return;
		}
		const id = parts[0] || "home";
		this.showPage(this.mPages.some(p => p.id === id) ? id : "home", query);
	}

	private addPage(page: Page): void {
		this.mPages.push(page);
		this.mLayerManager.addToLayer("base", page);
	}

	public logOut(redirect = ""): void {
		Session.signOut(function () {
			if (redirect !== "") {
				window.open("/home?redirect=" + redirect, "_self");
				return;
			}
			window.open("/", "_self");
		});
	}

	// Kept for when the app opens without a connection.
	private fetchUser(): void {
		const ks = new LocalStorage();
		const me = new HttpRequest("GET", "/v1/users/whoami");
		me.setSync();
		me.setQuiet();
		me.process(function (status: any, data: any) {
			const user = data as Responses.UsersWhoAmI;
			const before = ks.get("user") as any as Models.User | null;
			if (before && before.uuid !== user.data.uuid) {
				Session.forgetOfflineCopies();
			}
			ks.set("user", user.data, 0);
			App.setUser(user.data);
			App.setLanguage(user.data.language.code);
		}, function (status: number) {
			const user = ks.get("user") as any as Models.User | null;
			if (user && Outbox.isTemporary(status)) {
				App.setUser(user);
				App.setLanguage(user.language.code);
			}
		});
	}

	public static isLoggedIn(): boolean {
		return Session.exists;
	}

	public static getTranslation(translation: Models.Translation): string {
		let t = "";
		switch (App.sLanguage) {
			case "sv":
				t = translation.sv;
				break;
			case "de":
				t = translation.de;
				break;
			default:
				break;
		}
		if (t === "") {
			t = translation.en;
		}
		return t;
	}

	public static setLanguage(language: string): void {
		switch (language) {
			case "sv":
				App.translations = Translations.sv;
				App.sLanguage = "sv";
				break;
			case "de":
				App.translations = Translations.de;
				App.sLanguage = "de";
				break;
			default:
				App.translations = Translations.en;
				App.sLanguage = "en";
				break;
		}
	}

	// The release the server sent this page with.
	public static get version(): string {
		return App.sVersion;
	}

	public static get language(): string {
		return App.sLanguage;
	}

	public static setUser(user: Models.User): void {
		App.sUser = user;
	}

	public static get user(): Models.User {
		return App.sUser;
	}


	showPage(id: string, data: any): void {
		if (id === "admin") {
			const ls = new LocalStorage();
			if (!App.user || !ls.get("scope")?.includes("admin")) {
				id = "home";
			}
		}
		let page: Page | null = null;
		this.mPages.forEach(e => {
			if (e.id === id) {
				page = e;
			} else {
				e.setVisible(false);
			}
		});
		if (page) {
			const p = page as Page;
			if (p.requireSignIn() && (!App.isLoggedIn() || !App.user)) {
				Session.signInAgain();
				return;
			}
			this.mCurrentPage = id;
			this.mCurrentData = data;
			p.setVisible(true);
			p.setData(data);
			this.mMainMenu.setActive(id === "list" && data && data.uuid ? "list:" + data.uuid : id);
		}
		this.mMainMenu.close();
	}
}
