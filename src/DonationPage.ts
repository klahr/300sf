/// <reference path="utils/Date.ts"/>
/// <reference path="Page.ts"/>
/// <reference path="ui/Input.ts"/>
/// <reference path="App.ts"/>

class DonationPage extends Page {
	constructor(appListener: IAppListener) {
		super(appListener);

		this.addProfile();

		const title = document.createElement("h1");
		title.style.textAlign = "center";
		title.innerHTML = App.translations.ThankYou;
		this.container.appendChild(title);

		const description = document.createElement("p");
		description.style.textAlign = "center";
		description.innerHTML = App.translations.DonationDescription;
		this.container.appendChild(description);

		const d = new Ui.Div();
		d.style.position = "relative";
		d.style.width = "100%";
		d.style.textAlign = "center";
		d.container.innerHTML = "<stripe-buy-button buy-button-id='buy_btn_1OAYnZJ2tiNkkhn0NLHP63kn' publishable-key='pk_live_51NyKuFJ2tiNkkhn0iHvdFGi21x3wg6lC7FJqZfyReNXgfTFTWVMeRCpHYsP4qtoQRwyNF7R23QMuJnSgRQMgsTHb001DGQqqV8'></stripe-buy-button>";
		this.attach(d);

		const fundingDiv = new Ui.Div();
		fundingDiv.style.position = "relative";
		this.attach(fundingDiv);

		const needed = 2352;
		const current = 700;
		const funding = new Ui.ProgressBar();
		funding.style.position = "relative";
		funding.setProgress(current / needed * 100);
		funding.style.width = "100px";
		funding.style.height = "24px";
		funding.style.marginTop = "32px";
		funding.style.transform = "rotate(-90deg)";
		fundingDiv.attach(funding);

		const amount = document.createElement("p");
		amount.style.marginLeft = "72px";
		amount.innerHTML = "700 / 2,352 SEK " + App.translations.Raised;
		fundingDiv.container.appendChild(amount);

		const amountDesc = document.createElement("p");
		amountDesc.className = "details";
		amountDesc.style.marginLeft = "72px";
		amountDesc.innerHTML = App.translations.RaisedDescription;
		fundingDiv.container.appendChild(amountDesc);
	}

	public get id(): string {
		return "donation";
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
}
