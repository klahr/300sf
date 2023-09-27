/// <reference path="ui/LayerManager.ts"/>
/// <reference path="HomePage.ts"/>
/// <reference path="SignUpPage.ts"/>
/// <reference path="ListPage.ts"/>
/// <reference path="ForgotPage.ts"/>
/// <reference path="MainMenu.ts"/>
/// <reference path="Translations.ts"/>
/// <reference path="models/Translation.ts"/>

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

	constructor(language: string, apiKey: string, redirect: string, queryParameters: object) {
		App.setLanguage(language);
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

		this.showPage(redirect === "" ? "home" : redirect, queryParameters);

		this.mLayerManager.addToLayer("base", this.mMainMenu);
	}

	public static isLoggedIn(): boolean {
		return getCookie("x300sf_at") !== "";
	}

	public static getTranslation(translation: Models.Translation): string {
		switch (App.sLanguage) {
			case "sv":
				return translation.sv;
			default:
				return translation.en;
		}
	}

	public static setLanguage(language: string): void {
		switch (language) {
			case "sv":
				App.translations = Translations.sv;
				App.sLanguage = "sv";
				break;
			default:
				App.translations = Translations.en;
				App.sLanguage = "en";
				break;
		}
	}

	public static getLanguage(): string {
		return App.sLanguage;
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
