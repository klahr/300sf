/// <reference path="utils/Date.ts"/>
/// <reference path="Page.ts"/>
/// <reference path="ui/Input.ts"/>

class ContributePage extends Page {
	constructor(appListener: IAppListener) {
		super(appListener);

		this.addProfile();

		const title = document.createElement("p");
		title.style.textAlign = "center";
		title.innerHTML = "sourcehut";
		this.container.appendChild(title);
	}

	public get id(): string {
		return "contribute";
	}

	public onBecomeVisible(): void {
		super.onBecomeVisible();
		this.refresh();
	}

	private refresh(): void {
		if (App.user) {
			this.mProfile.setImage("/img/portraits/" + App.user.portrait);
		}
	}
}