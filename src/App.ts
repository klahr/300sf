interface IAppListener {
	
}

class App implements IAppListener {
	private mLayerManager = new Ui.LayerManager();	

	constructor() {
		this.mLayerManager.addLayer("base");
		this.mLayerManager.showLayer("base");

		const homePage = new HomePage(this);
		this.mLayerManager.addToLayer("base", homePage);
	}
}
