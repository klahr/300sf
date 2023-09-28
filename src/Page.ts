/// <reference path="ui/PopUp.ts"/>

class Page extends Ui.Widget {
	private mAppListener: IAppListener;

	constructor(appListener: IAppListener) {
		super();

		this.mAppListener = appListener;

		this.mContainer = document.createElement("div");
		this.mContainer.className = "uix-div surface";
		this.mContainer.style.minHeight = "32px";

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
	}

	protected get appListener(): IAppListener {
		return this.mAppListener;
	}

	public setData(data: any): void {
		super.setData(data);
		history.pushState({}, "", "/" + this.id + window.location.search);
	}

	protected onCreateLoggedIn(): void {}

	protected onCreateLoggedOut(): void {}
}
