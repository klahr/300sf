/// <reference path="Page.ts"/>
/// <reference path="App.ts"/>
/// <reference path="AdminSpeciesPopUp.ts"/>

class AdminPage extends Page {
	constructor(appListener: IAppListener) {
		super(appListener);

		this.addProfile();

		const title = document.createElement("h1");
		title.innerHTML = "Admin";
		this.container.appendChild(title);

		this.addEntry(App.translations.SpeciesPlural, "raven", openAdminSpeciesPopUp);
		this.addEntry("Announcements", "campaign", () => {});
	}

	public get id(): string {
		return "admin";
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

	public requireSignIn(): boolean {
		return true;
	}

	private addEntry(name: string, icon: string, action: () => void) {
		const c = document.createElement("p");
		c.className = "entry";
		c.innerHTML = "<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>" + icon + "</span> " + name;
		this.container.appendChild(c);
		c.addEventListener("click", function () {
			action();
		});
	}
}
