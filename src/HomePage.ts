/// <reference path="ui/Widget.ts"/>
/// <reference path="ui/Image.ts"/>
/// <reference path="Page.ts"/>

class HomePage extends Page {
	constructor(appListener: IAppListener) {
		super(appListener);
	}

	public onCreateLoggedOut(): void {
		const logo = new Ui.Image("img/300sf.svg", "128px", "128px");
		this.attach(logo);
	}
}
