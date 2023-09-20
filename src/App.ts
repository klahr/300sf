/// <reference path="ui/LayerManager.ts"/>
/// <reference path="HomePage.ts"/>
/// <reference path="MainMenu.ts"/>
/// <reference path="Translations.ts"/>

interface IAppListener {
	openMainMenu(): void;
}

class App implements IAppListener {
	private mLayerManager = new Ui.LayerManager();
	private mMainMenu = new MainMenu();
	private static sLanguage: string = "en";
	public static translations = Translations.en;

	constructor(language: string) {
		App.setLanguage(language);

		this.mLayerManager.addLayer("base");
		this.mLayerManager.showLayer("base");

		const homePage = new HomePage(this);
		this.mLayerManager.addToLayer("base", homePage);

		this.mLayerManager.addToLayer("base", this.mMainMenu);
	}

	public static isLoggedIn(): boolean {
		return getCookie("x300sf_at") !== "";
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
		this.mMainMenu.addClass("open");
	}
}
