/// <reference path="Page.ts"/>
/// <reference path="App.ts"/>

class AdminPage extends Page {
	constructor(appListener: IAppListener) {
		super(appListener);

		const self = this;

		this.addProfile();

		const title = document.createElement("h1");
		title.innerHTML = "Admin";
		this.container.appendChild(title);

		const speciesLbl = document.createElement("h2");
		speciesLbl.innerHTML = "Species";
		this.container.appendChild(speciesLbl);

		const announcementsLbl = document.createElement("h2");
		announcementsLbl.innerHTML = "Announcements";
		this.container.appendChild(announcementsLbl);
	}

	public get id(): string {
		return "admin";
	}

	public requireSignIn(): boolean {
		return true;
	}
}
