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
/// <reference path="SetPasswordPage.ts"/>
/// <reference path="DeleteAccountPage.ts"/>
/// <reference path="Translations.ts"/>
/// <reference path="models/Translation.ts"/>
/// <reference path="models/User.ts"/>

interface IAppListener {
	openMainMenu(): void;
	showPage(id: string, data: any): void;
	logOut(): void;
}

class App implements IAppListener {
	private mLayerManager = new Ui.LayerManager();
	private mMainMenu = new MainMenu(this);
	private static sLanguage: string = "en";
	public static translations = Translations.en;
	private mPages = new Array<Page>();
	private static sUser: Models.User;

	constructor(language: string, redirect: string, queryParameters: object) {
		const lang = getCookie("lang");
		App.setLanguage(lang !== "" ? lang : language);

		this.mLayerManager.addLayer("base");
		this.mLayerManager.showLayer("base");

		this.addPage(new HomePage(this));
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

	private addPage(page: Page): void {
		this.mPages.push(page);
		this.mLayerManager.addToLayer("base", page);
	}

	public logOut(): void {
		setCookie("x300sf_at", "", 0);
		setCookie("lang", "", 0);
		window.open("/", "_self");
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
