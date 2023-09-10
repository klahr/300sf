/// <reference path="ui/LayerManager.ts"/>
/// <reference path="HomePage.ts"/>

interface IAppListener {
	openMainMenu(): void;
}

class App implements IAppListener {
	private mLayerManager = new Ui.LayerManager();
	private mMainMenu = new MainMenu();

	constructor() {
		this.mLayerManager.addLayer("base");
		this.mLayerManager.showLayer("base");

		const homePage = new HomePage(this);
		this.mLayerManager.addToLayer("base", homePage);

		this.mLayerManager.addToLayer("base", this.mMainMenu);
	}

	openMainMenu(): void {
		this.mMainMenu.addClass("open");
	}
}
