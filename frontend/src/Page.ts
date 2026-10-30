/// <reference path="ui/PopUp.ts"/>
/// <reference path="News.ts"/>

class Page extends Ui.Widget {
	private mAppListener: IAppListener;
	protected mProfile: Ui.Image;
	private mAppBar: HTMLDivElement;

	constructor(appListener: IAppListener) {
		super();

		this.mAppListener = appListener;

		this.mContainer = document.createElement("div");
		this.mContainer.className = "uix-div surface";

		this.mAppBar = document.createElement("div");
		this.mAppBar.className = "app-bar";
		this.mContainer.appendChild(this.mAppBar);

		const self = this;

		const logo = document.createElement("img");
		logo.className = "app-bar-logo";
		logo.src = "/img/300sf.svg";
		logo.alt = App.translations.AppName;
		logo.addEventListener("click", function () {
			self.mAppListener.showPage("home", null);
		});
		this.mAppBar.appendChild(logo);

		const spacer = document.createElement("div");
		spacer.className = "spacer";
		this.mAppBar.appendChild(spacer);

		if (App.isLoggedIn()) {
			this.mAppBar.appendChild(News.bell(appListener));
		}

		this.mContainer.addEventListener("scroll", function () {
			self.mAppBar.classList.toggle("scrolled", self.mContainer.scrollTop > 4);
		}, { passive: true });

		if (App.isLoggedIn()) {
			this.onCreateLoggedIn();
		} else {
			this.onCreateLoggedOut();
		}
	}

	public get id(): string {
		throw new Error("Page must implement id(): string")
	}

	public setVisible(visible: boolean): void {
		const wasVisible = this.mContainer.style.display === "block";
		this.mContainer.style.display = visible ? "block" : "none";
		if (visible) {
			if (!wasVisible) {
				// Restart the entrance animation each time the page comes into view.
				this.mContainer.classList.remove("page-enter");
				void this.mContainer.offsetWidth;
				this.mContainer.classList.add("page-enter");
			}
			this.onBecomeVisible();
		}
	}

	public requireSignIn(): boolean {
		return false;
	}

	protected onBecomeVisible(): void { }

	protected addProfile(): void {
		this.mProfile = new Ui.Image("", "38px", "38px");
		this.mProfile.addClass("portrait");
		this.mProfile.container.setAttribute("role", "button");
		this.mProfile.container.setAttribute("aria-label", App.translations.Profile);
		this.mProfile.container.setAttribute("aria-haspopup", "menu");
		this.mProfile.container.tabIndex = 0;
		const self = this;
		this.mProfile.container.addEventListener("click", function () {
			AccountMenu.show(self.mAppListener, self.mProfile.container);
		});
		this.mProfile.container.addEventListener("keydown", function (event) {
			if (event.key === "Enter" || event.key === " ") {
				event.preventDefault();
				AccountMenu.show(self.mAppListener, self.mProfile.container);
			}
		});
		this.mChildren.push(this.mProfile);
		this.mAppBar.appendChild(this.mProfile.container);
	}

	protected get appListener(): IAppListener {
		return this.mAppListener;
	}

	protected get profile(): Ui.Image {
		return this.mProfile;
	}

	public setData(data: any): void {
		super.setData(data);
		Page.setUrl(this.path);
	}

	protected get path(): string {
		return "/" + this.id;
	}

	// Replaces the history entry when the path is unchanged (first load,
	// refreshes) and pushes one when the user moves to another page. Query
	// parameters such as ?token= only belong to the page they were sent to.
	protected static setUrl(path: string): void {
		if (window.location.pathname === path) {
			history.replaceState({}, "", path + window.location.search);
		} else {
			history.pushState({}, "", path);
		}
	}

	// Wraps each h2 and everything after it, up to the next h2, in a card. Call it once the page's
	// static content is built; the elements move but keep their identity, so references stay valid.
	protected groupSections(): HTMLDivElement[] {
		const cards: HTMLDivElement[] = [];
		let card: HTMLDivElement | null = null;
		Array.from(this.mContainer.children).forEach(e => {
			if (e.tagName === "H2") {
				card = document.createElement("div");
				card.className = "section-card";
				this.mContainer.insertBefore(card, e);
				cards.push(card);
			}
			if (card) {
				card.appendChild(e);
			}
		});
		return cards;
	}

	// Wraps the element and its following siblings in a card.
	protected wrapInCard(first: Element): HTMLDivElement {
		const card = document.createElement("div");
		card.className = "section-card";
		this.mContainer.insertBefore(card, first);
		let e: Element | null = first;
		while (e) {
			const next: Element | null = e.nextElementSibling;
			card.appendChild(e);
			e = next;
		}
		return card;
	}

	protected onCreateLoggedIn(): void { }

	protected onCreateLoggedOut(): void { }
}
