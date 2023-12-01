/// <reference path="Page.ts"/>
/// <reference path="ui/Input.ts"/>
/// <reference path="App.ts"/>

class ManageCompetitionsPage extends Page {
	private mCreateButton: HTMLButtonElement;

	constructor(appListener: IAppListener) {
		super(appListener);
		const self = this;

		this.addProfile();

		const title = document.createElement("h1");
		title.innerHTML = "Competitions";
		this.container.appendChild(title);

		const newLbl = document.createElement("h2");
		newLbl.innerHTML = "New competition";
		this.container.appendChild(newLbl);

		this.mCreateButton = document.createElement("button");
		this.mCreateButton.innerHTML = "Create competition ...";
		this.mCreateButton.style.width = "auto";
		this.mCreateButton.style.cssFloat = "right";
		this.mCreateButton.style.borderRadius = "8px";
		this.mCreateButton.style.backgroundColor = "var(--success-color)";
		this.container.appendChild(this.mCreateButton);
		this.mCreateButton.addEventListener("click", function () {
			new CreateCompetitionPopUp((competition: Models.Competition | null) => {
				console.log("closed");
			});
		});
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