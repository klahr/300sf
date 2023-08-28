/// <reference path="ui/Widget.ts"/>

class HomePage extends Ui.Widget {
	private mAppListener: IAppListener;	

	constructor(appListener: IAppListener) {
		super();

		this.mAppListener = appListener;

		this.mContainer = document.createElement("div");
		this.mContainer.style.width = "100%";
		this.mContainer.style.height = "100%";
		this.mContainer.style.overflowY = "auto";
		this.mContainer.style.backgroundColor = "#ff00ff";

		if (this.isLoggedIn()) {
			this.setupLoggedIn();
		} else {
			this.setupLoggedOut();
		}
	}

	private isLoggedIn(): boolean {
		// TODO
		return false;
	}

	private setupLoggedOut(): void {
		const logo = new Ui.Image("img/300sf.svg", "128px", "128px");
		this.attach(logo);
	}

	private setupLoggedIn(): void {
		
	}
}
