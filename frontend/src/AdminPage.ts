/// <reference path="Page.ts"/>
/// <reference path="App.ts"/>
/// <reference path="AdminSpeciesPopUp.ts"/>
/// <reference path="AdminAnnouncementsPopUp.ts"/>

class AdminPage extends Page {
	private mTools: HTMLDivElement;
	constructor(appListener: IAppListener) {
		super(appListener);

		this.addProfile();

		const title = document.createElement("h1");
		title.innerHTML = App.translations.Admin;
		this.container.appendChild(title);

		this.mTools = document.createElement("div");
		this.mTools.className = "manage-list";
		this.container.appendChild(this.mTools);
		this.addEntry(App.translations.SpeciesPlural, "raven", openAdminSpeciesPopUp);
		this.addEntry(App.translations.Announcements, "campaign", openAdminAnnouncementsPopUp);

		const inviteLbl = document.createElement("h2");
		inviteLbl.textContent = App.translations.Invite;
		this.container.appendChild(inviteLbl);

		const inviteEmail = new Ui.Input("email");
		inviteEmail.setPlaceHolder(App.translations.Email);
		this.attach(inviteEmail);

		const inviteButton = document.createElement("button");
		inviteButton.innerHTML = "<span class='material-symbols-outlined'>send</span>" + App.translations.SendInvite;
		this.container.appendChild(inviteButton);
		inviteButton.addEventListener("click", function() {
			const r = new HttpRequest("POST", "/v1/invite", {email: inviteEmail.value, language_code: "sv"}).process((status: any, data: any) => {
			});
		});

		this.groupSections();
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
		const c = document.createElement("div");
		c.className = "manage-row";
		c.setAttribute("role", "button");
		c.tabIndex = 0;
		c.innerHTML = "<span class='material-symbols-outlined tool-icon'>" + icon + "</span><p>" + name + "</p><span class='material-symbols-outlined'>chevron_right</span>";
		c.addEventListener("keydown", function (event) {
			if (event.key === "Enter") {
				action();
			}
		});
		this.mTools.appendChild(c);
		c.addEventListener("click", function () {
			action();
		});
	}
}
