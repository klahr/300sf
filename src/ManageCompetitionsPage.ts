/// <reference path="Page.ts"/>
/// <reference path="ui/Input.ts"/>
/// <reference path="App.ts"/>

class ManageCompetitionsPage extends Page {
	constructor(appListener: IAppListener) {
		super(appListener);

		this.addProfile();
	}

	public get id(): string {
		return "competitions";
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