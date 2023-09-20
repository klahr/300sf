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
		return false;
	}

	public static setLanguage(language: string): void {
		switch (language) {
			case "sv":
				App.translations = Translations.sv;
				break;
			default:
				App.translations = Translations.en;
				break;
		}
	}

	openMainMenu(): void {
		this.mMainMenu.addClass("open");
	}
}
