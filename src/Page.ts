/// <reference path="ui/PopUp.ts"/>

class Page extends Ui.Widget {
	private mAppListener: IAppListener;
	protected mProfile: Ui.Image;
	private mContent: Ui.Div;

	constructor(appListener: IAppListener) {
		super();

		this.mAppListener = appListener;

		this.mContainer = document.createElement("div");
		this.mContainer.className = "uix-div surface";
		this.mContainer.style.minHeight = "32px";

		this.mContent = new Ui.Div();
		this.mContent.style.width = "100%";
		this.attach(this.mContent);

		const menu =  document.createElement("span");
		menu.className = "material-symbols-outlined menu";
		menu.style.cursor = "pointer";
		menu.innerHTML = "menu";
		this.mContainer.appendChild(menu);
		const self = this;
		menu.addEventListener("click", function() {
			self.mAppListener.openMainMenu();
		});

		if (App.isLoggedIn()) {
			this.onCreateLoggedIn();
		} else {
			this.onCreateLoggedOut();
		}
	}

	public get id(): string {
		throw new Error("Page must implement id(): string")
	}

	public setVisible(visible: boolean): void {
		this.mContainer.style.display = visible ? "block" : "none";
		if (visible) {
			this.onBecomeVisible();
		}
	}

	protected onBecomeVisible(): void {}

	protected addProfile(): void {
		this.mProfile = new Ui.Image("/img/99249498-452f-4b0b-a1f5-45902f7ffe48.png", "48px", "48px");
		this.mProfile.addClass("portrait");
		this.mProfile.container.addEventListener("click", function() {
			openProfilePopUp(function() {
				console.log("close profile");
			});
		});
		this.attach(this.mProfile);
	}

	protected get appListener(): IAppListener {
		return this.mAppListener;
	}

	protected get profile(): Ui.Image {
		return this.mProfile;
	}

	public setData(data: any): void {
		super.setData(data);
		history.pushState({}, "", "/" + this.id + window.location.search);
	}

	protected onCreateLoggedIn(): void {}

	protected onCreateLoggedOut(): void {}
}
