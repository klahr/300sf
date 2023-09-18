/// <reference path="ui/PopUp.ts"/>

class Page extends Ui.Widget {
    protected mAppListener: IAppListener;

    constructor(appListener: IAppListener) {
		super();

		this.mAppListener = appListener;

		this.mContainer = document.createElement("div");
        this.mContainer.className = "uix-div surface";
        this.mContainer.style.height = "32px";

        const menu =  document.createElement("span");
        menu.className = "material-symbols-outlined menu";
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

    protected onCreateLoggedIn(): void {}

    protected onCreateLoggedOut(): void {}
}
