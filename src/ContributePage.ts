/// <reference path="utils/Date.ts"/>
/// <reference path="Page.ts"/>
/// <reference path="ui/Input.ts"/>

class ContributePage extends Page {
	private mContent: Ui.Div;

	constructor(appListener: IAppListener) {
		super(appListener);

		this.mContent = new Ui.Div();
		this.mContent.style.position = "relative";
		this.mContent.style.width = "100%";
		this.attach(this.mContent);

		const profile = new Ui.Image("/img/99249498-452f-4b0b-a1f5-45902f7ffe48.png", "48px", "48px");
		profile.addClass("portrait");
		profile.container.addEventListener("click", function() {
			openProfilePopUp(function() {
				console.log("close profile");
			});
		});
		this.attach(profile);

		const title = document.createElement("p");
		title.style.textAlign = "center";
		title.innerHTML = "sourcehut";
		this.container.appendChild(title);
	}

	public get id(): string {
		return "contribute";
	}
}