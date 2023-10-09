/// <reference path="utils/Date.ts"/>
/// <reference path="Page.ts"/>
/// <reference path="ui/Input.ts"/>

class DonationPage extends Page {
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
		title.innerHTML = App.translations.DonationDescription;
		this.container.appendChild(title);

		const d = new Ui.Div();
		d.style.position = "relative";
		d.style.width = "100%";
		d.style.textAlign = "center";
		d.container.innerHTML = "<stripe-buy-button buy-button-id='buy_btn_1NyLAFJ2tiNkkhn0oBh7Zgpd' publishable-key='pk_live_51NyKuFJ2tiNkkhn0iHvdFGi21x3wg6lC7FJqZfyReNXgfTFTWVMeRCpHYsP4qtoQRwyNF7R23QMuJnSgRQMgsTHb001DGQqqV8'></stripe-buy-button>";
		this.attach(d);
	}

	public get id(): string {
		return "donation";
	}
}