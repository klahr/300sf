/// <reference path="ui/LayerManager.ts"/>
/// <reference path="HomePage.ts"/>
/// <reference path="SignUpPage.ts"/>
/// <reference path="ListPage.ts"/>
/// <reference path="SettingsPage.ts"/>
/// <reference path="DonationPage.ts"/>
/// <reference path="ForgotPage.ts"/>
/// <reference path="ManageTagsPage.ts"/>
/// <reference path="ReportsPage.ts"/>
/// <reference path="MainMenu.ts"/>
/// <reference path="ManageChecklistsPage.ts"/>
/// <reference path="ContributePage.ts"/>
/// <reference path="Translations.ts"/>
/// <reference path="models/Translation.ts"/>
/// <reference path="models/User.ts"/>

interface IAppListener {
	openMainMenu(): void;
	showPage(id: string, data: any): void;
}

class App implements IAppListener {
	private mLayerManager = new Ui.LayerManager();
	private mMainMenu = new MainMenu(this);
	private static sLanguage: string = "en";
	public static translations = Translations.en;
	public static apiKey: string;
	private mPages = new Array<Page>();
	private static sUser: Models.User;

	constructor(language: string, apiKey: string, redirect: string, queryParameters: object) {
		const lang = getCookie("lang");
		App.setLanguage(lang !== "" ? lang : language);

		App.apiKey = apiKey;

		this.mLayerManager.addLayer("base");
		this.mLayerManager.showLayer("base");

		const homePage = new HomePage(this);
		this.mPages.push(homePage);
		this.mLayerManager.addToLayer("base", homePage);

		const signUpPage = new SignUpPage(this);
		this.mPages.push(signUpPage);
		this.mLayerManager.addToLayer("base", signUpPage);

		const forgotPage = new ForgotPage(this);
		this.mPages.push(forgotPage);
		this.mLayerManager.addToLayer("base", forgotPage);

		const listPage = new ListPage(this);
		this.mPages.push(listPage);
		this.mLayerManager.addToLayer("base", listPage);

		const contributePage = new ContributePage(this);
		this.mPages.push(contributePage);
		this.mLayerManager.addToLayer("base", contributePage);

		const donationPage = new DonationPage(this);
		this.mPages.push(donationPage);
		this.mLayerManager.addToLayer("base", donationPage);

		const settingsPage = new SettingsPage(this);
		this.mPages.push(settingsPage);
		this.mLayerManager.addToLayer("base", settingsPage);

		const manageChecklistsPage = new ManageCheckListsPage(this);
		this.mPages.push(manageChecklistsPage);
		this.mLayerManager.addToLayer("base", manageChecklistsPage);

		const manageTagsPage = new ManageTagsPage(this);
		this.mPages.push(manageTagsPage);
		this.mLayerManager.addToLayer("base", manageTagsPage);

		const reportsPage = new ReportsPage(this);
		this.mPages.push(reportsPage);
		this.mLayerManager.addToLayer("base", reportsPage);

		this.mLayerManager.addToLayer("base", this.mMainMenu);

		if (App.isLoggedIn()) {
			this.fetchUser();
		}

		const self = this;
		document.addEventListener("visibilitychange", () => {
			if (document.visibilityState === "visible") {
				if (App.isLoggedIn()) {
					self.fetchUser();
				} else {
					window.open("/?redirect=" + window.location.href, "_self");
				}
			}
		});

		this.showPage(redirect === "" ? "home" : redirect, queryParameters);
	}

	public static logOut(): void {
		setCookie("x300sf_at", "", 0);
		setCookie("lang", "", 0);
		location.reload();
	}

	private fetchUser(): void {
		const me = new HttpRequest("GET", "/v1/users/me");
		me.setSync();
		me.process(function(status: any, data: any) {
			const d = data as Responses.UsersMe;
			App.setUser(d.user);
		});
	}

	public static isLoggedIn(): boolean {
		return getCookie("x300sf_at") !== "";
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

	public static get language(): string {
		return App.sLanguage;
	}

	public static setUser(user: Models.User): void {
		App.sUser = user;
	}

	public static get user(): Models.User {
		return App.sUser;
	}

	openMainMenu(): void {
		this.mMainMenu.open();;
	}

	showPage(id: string, data: any): void {
		this.mPages.forEach(e => {
			if (e.id === id) {
				e.setVisible(true);
				e.setData(data);
			} else {
				e.setVisible(false);
			}
		});
		this.mMainMenu.close();
	}
}
