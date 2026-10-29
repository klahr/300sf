/// <reference path="utils/Date.ts"/>
/// <reference path="Page.ts"/>
/// <reference path="ui/Input.ts"/>
/// <reference path="App.ts"/>

class DonationPage extends Page {
	constructor(appListener: IAppListener) {
		super(appListener);

		this.addProfile();

		const card = document.createElement("div");
		card.className = "section-card donation";
		this.container.appendChild(card);

		const icon = document.createElement("span");
		icon.className = "material-symbols-outlined";
		icon.textContent = "volunteer_activism";
		card.appendChild(icon);

		const title = document.createElement("h1");
		title.innerHTML = App.translations.ThankYou;
		card.appendChild(title);

		const description = document.createElement("p");
		description.innerHTML = App.translations.DonationDescription;
		card.appendChild(description);

		const buy = document.createElement("div");
		buy.innerHTML = "<stripe-buy-button buy-button-id='buy_btn_1OAYnZJ2tiNkkhn0NLHP63kn' publishable-key='pk_live_51NyKuFJ2tiNkkhn0iHvdFGi21x3wg6lC7FJqZfyReNXgfTFTWVMeRCpHYsP4qtoQRwyNF7R23QMuJnSgRQMgsTHb001DGQqqV8'></stripe-buy-button>";
		card.appendChild(buy);

		const funding = document.createElement("div");
		funding.className = "section-card funding";
		this.container.appendChild(funding);

		const needed = 2352;
		const current = 1000;
		const amount = document.createElement("p");
		amount.className = "funding-amount";
		amount.innerHTML = "<strong>" + current.toLocaleString(App.language) + "</strong> / " + needed.toLocaleString(App.language) + " SEK " + App.translations.Raised;
		funding.appendChild(amount);

		const bar = new Ui.ProgressBar();
		this.mChildren.push(bar);
		funding.appendChild(bar.container);
		requestAnimationFrame(function () {
			requestAnimationFrame(function () {
				bar.setProgress(current / needed * 100);
			});
		});

		const amountDesc = document.createElement("p");
		amountDesc.className = "details";
		amountDesc.innerHTML = App.translations.RaisedDescription;
		funding.appendChild(amountDesc);
	}

	public get id(): string {
		return "donation";
	}

	public onBecomeVisible(): void {
		super.onBecomeVisible();
		this.refresh();

		// Stripe's script sets its own cookies, so it is loaded only here, when someone may donate.
		if (!document.getElementById("stripe-buy-button")) {
			const script = document.createElement("script");
			script.id = "stripe-buy-button";
			script.async = true;
			script.src = "https://js.stripe.com/v3/buy-button.js";
			document.head.appendChild(script);
		}
	}

	private refresh(): void {
		if (App.user) {
			this.mProfile.setImage("/img/portraits/" + App.user.portrait);
		}
	}
}
