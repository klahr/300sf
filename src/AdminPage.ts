/// <reference path="Page.ts"/>

class AdminPage extends Page {
	constructor(appListener: IAppListener) {
		super(appListener);

		const self = this;

		this.addProfile();

		const title = document.createElement("h1");
		title.innerHTML = "Admin";
		this.container.appendChild(title);
	}

	public get id(): string {
		return "admin";
	}

	public requireSignIn(): boolean {
		return true;
	}
}
