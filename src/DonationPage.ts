/// <reference path="utils/Date.ts"/>
/// <reference path="Page.ts"/>
/// <reference path="ui/Input.ts"/>

class DonationPage extends Page {
	constructor(appListener: IAppListener) {
		super(appListener);

		this.addProfile();

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

	public onBecomeVisible(): void {
		super.onBecomeVisible();
		this.refresh();
	}

	private refresh(): void {
		this.mProfile.setImage("/img/portraits/" + App.user.portrait);
	}
}