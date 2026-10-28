/// <reference path="Page.ts"/>
/// <reference path="AnalyzePopUp.ts"/>
/// <reference path="App.ts"/>

class ExperimentalPage extends Page {
	constructor(appListener: IAppListener) {
		super(appListener);

		const title = document.createElement("h1");
		title.innerHTML = App.translations.Experimental;
		this.container.appendChild(title);

		this.addProfile();
	}

	public get id(): string {
		return "experimental";
	}

	public requireSignIn(): boolean {
		return true;
	}

	public onBecomeVisible(): void {
		super.onBecomeVisible();
		this.refresh();
	}

	private refresh(): void {
		this.mProfile.setImage("/img/portraits/" + App.user.portrait);
	}
}