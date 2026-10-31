/// <reference path="HttpRequest.ts"/>
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

		// The progress is set under Admin → Donation.
		const amount = document.createElement("p");
		amount.className = "funding-amount";
		funding.appendChild(amount);

		const bar = new Ui.ProgressBar();
		this.mChildren.push(bar);
		funding.appendChild(bar.container);
		new HttpRequest("GET", "/v1/donation", null).process(function (status: any, data: { raised: number, goal: number }) {
			amount.innerHTML = "<strong>" + data.raised.toLocaleString(App.language) + "</strong> / " + data.goal.toLocaleString(App.language) + " SEK " + App.translations.Raised;
			requestAnimationFrame(function () {
				requestAnimationFrame(function () {
					bar.setProgress(Math.min(100, data.raised / data.goal * 100));
				});
			});
		}, function () {
			funding.style.display = "none";
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
		// Anyone may donate; signed out there is no profile to show.
		this.mProfile.container.style.display = App.user ? "" : "none";
		if (App.user) {
			this.mProfile.setImage("/img/portraits/" + App.user.portrait);
		}
	}
}
