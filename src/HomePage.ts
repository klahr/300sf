/// <reference path="ui/Widget.ts"/>
/// <reference path="ui/Image.ts"/>
/// <reference path="Page.ts"/>

class HomePage extends Page {
	constructor(appListener: IAppListener) {
		super(appListener);
	}

	protected onCreateLoggedIn(): void {
        const self = this;

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
	
	protected onCreateLoggedOut(): void {
		const logo = new Ui.Image("img/300sf.svg", "128px", "128px");
		this.attach(logo);
	}
}
