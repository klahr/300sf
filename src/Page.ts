/// <reference path="ui/PopUp.ts"/>

class Page extends Ui.Widget {
    protected mAppListener: IAppListener;

    constructor(appListener: IAppListener) {
		super();

		this.mAppListener = appListener;

		this.mContainer = document.createElement("div");
        this.mContainer.className = "uix-div surface";
        this.mContainer.style.height = "32px";

		if (this.isLoggedIn()) {
			this.onCreateLoggedIn();
		} else {
			this.onCreateLoggedOut();
		}
	}

    public onCreateLoggedIn(): void {
        const self = this;

        const menu =  document.createElement("span");
        menu.className = "material-symbols-outlined menu";
        menu.innerHTML = "menu";
        this.mContainer.appendChild(menu);
        menu.addEventListener("click", function() {
            self.mAppListener.openMainMenu();
        });

        const profile = new Ui.Image("img/99249498-452f-4b0b-a1f5-45902f7ffe48.png", "48px", "48px");
        profile.style.left = "unset";
        profile.style.right = "var(--medium-distance)";
        profile.style.top = "var(--small-distance)";
        profile.style.borderRadius = "100%";
        profile.container.addEventListener("click", function() {
            new Ui.PopUp("img/300sf.svg", "Sidan finns ännu inte.", function() {
			}, null);
        });
        this.attach(profile);
    }

    public onCreateLoggedOut(): void {}

    public isLoggedIn(): boolean {
        return true;
    }
}
