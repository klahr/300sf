/// <reference path="ui/Div.ts"/>
/// <reference path="App.ts"/>
/// <reference path="ui/ConfirmBox.ts"/>

class MainMenu extends Ui.Widget {
	private mContent = new Ui.Div();
	private mMainMenu: HTMLDivElement;
	private mAppListener: IAppListener;

	constructor(appListener: IAppListener) {
		super();

		this.mAppListener = appListener;

		const self = this;

		this.mContainer = document.createElement("div");
		this.mContainer.className = "uix-div main-menu-background";
		this.mContainer.addEventListener("click", function() {
			self.close();
		});

		this.mMainMenu = document.createElement("div");
		this.mMainMenu.className = "uix-div main-menu";
		this.mContainer.appendChild(this.mMainMenu);

		this.mContent.addClass("content");
		this.mContent.container.addEventListener("click", function(event) {
			event.stopPropagation();
		});
		this.mMainMenu.appendChild(this.mContent.container);

		const logo = document.createElement("img");
		logo.src = "/img/300sf.svg";
		logo.style.width = "100%";
		logo.style.marginTop = "29px";
		logo.style.cursor = "pointer";
		logo.alt = App.translations.AppName;
		logo.setAttribute("height", "36");
		logo.addEventListener("click", function() {
			self.mAppListener.showPage("home", null);
		});
		this.mContent.container.appendChild(logo);
	}

	protected onAttachedToParent(_parent: Ui.Widget): void {
		if (App.isLoggedIn()) {
			this.createLoggedIn();
		} else {
			this.createLoggedOut();
		}
	}

	public open(): void {
		this.mMainMenu.classList.add("open");
		this.mContainer.classList.add("open");
	}

	public close(): void {
		this.mContainer.classList.remove("open");
		this.mMainMenu.classList.remove("open");
	}

	private createLoggedIn(): void  {
		const self = this;

		this.addSection(App.translations.AppName);
		this.addEntry("<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>home</span> " + App.translations.Dashboard, function() {
			self.mAppListener.showPage("home", null);
		});

		this.addEntry("<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>raven</span> " + App.translations.Sightings, function() {
			self.mAppListener.showPage("reports", null);
		});

		this.addEntry("<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>edit_note</span>  " + App.translations.Checklists, function() {
			self.mAppListener.showPage("checklists", null);
		});

		this.addEntry("<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>label</span>  " + App.translations.Tags, function() {
			self.mAppListener.showPage("tags", null);
		});

		this.addSection(App.translations.Checklists);
		const checklists = document.createElement("div");
		checklists.style.padding = "0";
		this.mContent.container.appendChild(checklists);

		const r = new HttpRequest("GET", "/v1/lists/list");
		r.process(function(status: any, data: any) {
			data.result.forEach(e => {
				const c = document.createElement("p");
				c.className = "entry";
				let color = "#0d81b4";
				if (e.color !== "") {
					color = e.color;
				}
				c.innerHTML = "<span style='font-size: 1rem; color: " + color + ";' class='material-symbols-outlined'>trip_origin</span> " + App.getTranslation(e.name);
				checklists.appendChild(c);
				c.addEventListener("click", function() {
					self.mAppListener.showPage("list", e);
				});
			});
		});

		this.addSection(App.translations.Support);
		this.addEntry("<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>mail</span> " + App.translations.Contact, function() {
			self.mAppListener.showPage("contact", null);
		});

		this.addEntry("<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>support</span> " + App.translations.Contribute, function() {
			self.mAppListener.showPage("contribute", null);
		});

		this.addEntry("<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>volunteer_activism</span> " + App.translations.Donation, function() {
			self.mAppListener.showPage("donation", null);
		});

		this.addEntry("<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>policy</span> " + App.translations.PrivacyPolicy, function() {
			window.open("/privacy-policy", "_blank");
		});

		this.addEntry("<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>handshake</span> " + App.translations.TermsOfService, function() {
			window.open("/terms", "_blank");
		});

		this.addSection(App.translations.Account);
		this.addEntry("<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>manage_accounts</span> " + App.translations.Account, function() {
			self.mAppListener.showPage("settings", null);
		});

		this.addEntry("<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>logout</span> " + App.translations.LogOut, function() {
			new Ui.ConfirmBox(App.translations.LogOut, App.translations.ConfirmLogOut, null, "var(--success-color)", function() {
				self.mAppListener.logOut();
			}, function() {
			}, App.translations.LogOut, App.translations.Cancel);
		});
	}

	private createLoggedOut(): void {
		const self = this;
		this.addSection(App.translations.Support);

		this.addEntry("<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>support</span> " + App.translations.Contribute, function() {
			self.mAppListener.showPage("contribute", null);
		});

		this.addEntry("<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>volunteer_activism</span> " + App.translations.Donation, function() {
			self.mAppListener.showPage("donation", null);
		});

		this.addEntry("<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>policy</span> " + App.translations.PrivacyPolicy, function() {
			window.open("/privacy-policy", "_blank");
		});

		this.addEntry("<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>handshake</span> " + App.translations.TermsOfService, function() {
			window.open("/terms", "_blank");
		});

		this.addSection(App.translations.Account);
		this.addEntry("<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>join</span> " + App.translations.SignUp, function() {
			window.open("/sign-up", "_self");
		});
	}

	private addSection(name: string): HTMLParagraphElement {
		const header = document.createElement("p");
		header.className = "section";
		header.innerHTML = name.toUpperCase();
		this.mContent.container.appendChild(header);
		return header;
	}

	private addEntry(name: string, callback: Function): HTMLParagraphElement {
		const entry = document.createElement("p");
		entry.className = "entry";
		entry.innerHTML = name;
		entry.addEventListener("click", function() {
			callback();
		});
		this.mContent.container.appendChild(entry);
		return entry;
	}
}
