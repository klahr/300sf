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
		const ks = new LocalStorage();
		const lang = ks.get("lang") as string;
		App.setLanguage(lang !== null && lang !== "" ? lang : language);

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

		this.mLayerManager.addToLayer("base", this.mMainMenu);

		if (App.isLoggedIn()) {
			this.fetchUser();
		}

		this.showPage(redirect === "" ? "home" : redirect, queryParameters);
	}

	private addPage(page: Page): void {
		this.mPages.push(page);
		this.mLayerManager.addToLayer("base", page);
	}

	public logOut(redirect = ""): void {
		const ks = new LocalStorage();
		ks.clear();
		if (redirect !== "") {
			redirect = "?redirect=" + redirect;
		}
		window.open("/" + redirect, "_self");
	}

	private fetchUser(): void {
		const me = new HttpRequest("GET", "/v1/users/whoami");
		me.setSync();
		me.process(function (status: any, data: any) {
			const user = data as Responses.UsersWhoAmI;
			App.setUser(user.data);
			App.setLanguage(user.data.language.code);
		});
	}

	public static isLoggedIn(): boolean {
		const ks = new LocalStorage();
		return ks.get("x300sf_at") !== null;
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
			if (!App.isLoggedIn()) {
				const self = this;
				const ks = new LocalStorage();
				const d = {
					grant_type: "refresh_token",
					refresh_token: ks.get("x300sf_rt")
				}
				if (d.refresh_token !== null) {
					const r = new HttpRequest("POST", "/v1/auth", d);
					r.process(function (status: any, token: Responses.AuthenticationResponse) {
						ks.set("x300sf_at", token.access_token, token.expires_in);
						ks.set("x300sf_rt", token.refresh_token, 3600 * 7 * 60);
						ks.set("lang", token.lang, token.expires_in);
						ks.set("scope", token.scope, token.expires_in);
						window.location.reload();
					}, function (status: any, data: Responses.GeneralErrorResponse, err: any) {
						self.logOut(window.location.pathname);
					});
				} else {
					p.setVisible(true);
					p.setData(data);
				}
			} else {
				p.setVisible(true);
				p.setData(data);
			}
		}
		this.mMainMenu.close();
	}
}
