/// <reference path="ui/Div.ts"/>
/// <reference path="App.ts"/>
/// <reference path="AccountMenu.ts"/>
/// <reference path="ReportPopUp.ts"/>
/// <reference path="Overview.ts"/>

// The app's navigation. On large screens a sidebar: the report button, the three main pages with the
// user's checklists and their progress under Checklists, and the account card at the bottom. On
// smaller screens a tab bar along the bottom with the same pages, the report button in the middle and
// the account last. Entries are real links so they can be opened in a new tab; a plain click switches
// the page in place. Signed out, there is no navigation.
class MainMenu extends Ui.Widget {
	private mAppListener: IAppListener;
	private mActive = "";
	private mLists: HTMLDivElement;

	constructor(appListener: IAppListener) {
		super();
		this.mAppListener = appListener;
		this.mContainer = document.createElement("div");
		this.mContainer.className = "uix-div app-nav";
	}

	protected onAttachedToParent(_parent: Ui.Widget): void {
		if (!App.isLoggedIn()) {
			return;
		}
		this.mContainer.appendChild(this.sidebar());
		this.mContainer.appendChild(this.tabBar());
		this.loadLists();

		const self = this;
		window.addEventListener("listschanged", function () {
			self.loadLists();
		});
		window.addEventListener("sightingadded", function () {
			self.loadLists();
		});
	}

	// Kept for the app shell; the navigation is always visible now.
	public open(): void { }

	public close(): void { }

	private sidebar(): HTMLElement {
		const self = this;
		const nav = document.createElement("nav");
		nav.className = "sidebar";
		nav.setAttribute("aria-label", App.translations.Menu);

		const logo = this.link("home", "/home", null);
		logo.className = "sidebar-logo";
		logo.innerHTML = "<img src='/img/300sf.svg' alt=''>";
		logo.setAttribute("aria-label", App.translations.AppName);
		delete logo.dataset.page;
		nav.appendChild(logo);

		const report = document.createElement("button");
		report.className = "sidebar-report";
		report.innerHTML = "<span class='material-symbols-outlined'>add</span>" + App.translations.SubmitSighting;
		report.addEventListener("click", function () {
			openQuickReport();
		});
		nav.appendChild(report);

		const main = document.createElement("div");
		main.className = "sidebar-main";
		nav.appendChild(main);
		main.appendChild(this.navItem("home", "/home", "home", App.translations.Dashboard));
		main.appendChild(this.navItem("reports", "/reports", "raven", App.translations.Sightings));
		main.appendChild(this.navItem("checklists", "/checklists", "checklist", App.translations.Checklists));

		this.mLists = document.createElement("div");
		this.mLists.className = "sidebar-lists";
		main.appendChild(this.mLists);
		main.appendChild(this.navItem("competitions", "/competitions", "emoji_events", App.translations.Competitions));

		const account = document.createElement("button");
		account.className = "sidebar-account";
		account.setAttribute("aria-haspopup", "menu");
		account.setAttribute("aria-expanded", "false");
		const portrait = document.createElement("img");
		portrait.alt = "";
		const text = document.createElement("span");
		text.className = "sidebar-account-text";
		const name = document.createElement("span");
		name.className = "sidebar-account-name";
		const handle = document.createElement("span");
		handle.className = "sidebar-account-handle";
		if (App.user) {
			portrait.src = "/img/portraits/" + App.user.portrait;
			name.textContent = App.user.display_name;
			handle.textContent = "@" + App.user.username;
		}
		text.appendChild(name);
		text.appendChild(handle);
		account.appendChild(portrait);
		account.appendChild(text);
		account.insertAdjacentHTML("beforeend", "<span class='material-symbols-outlined'>more_horiz</span>");
		account.addEventListener("click", function () {
			AccountMenu.show(self.mAppListener, account);
		});
		nav.appendChild(account);

		return nav;
	}

	private tabBar(): HTMLElement {
		const self = this;
		const nav = document.createElement("nav");
		nav.className = "tab-bar";
		nav.setAttribute("aria-label", App.translations.Menu);

		nav.appendChild(this.tab("home", "/home", "home", App.translations.Dashboard));
		nav.appendChild(this.tab("reports", "/reports", "raven", App.translations.Sightings));

		const report = document.createElement("button");
		report.className = "tab tab-report";
		report.setAttribute("aria-label", App.translations.SubmitSighting);
		report.innerHTML = "<span class='tab-report-circle'><span class='material-symbols-outlined'>add</span></span><span class='tab-label'>" + App.translations.ReportShort + "</span>";
		report.addEventListener("click", function () {
			openQuickReport();
		});
		nav.appendChild(report);

		nav.appendChild(this.tab("checklists", "/checklists", "checklist", App.translations.TabLists));

		const account = document.createElement("button");
		account.className = "tab tab-account";
		account.dataset.page = "account";
		account.setAttribute("aria-haspopup", "menu");
		account.innerHTML = "<img alt='' src='/img/portraits/" + (App.user ? Utils.escapeHtml(App.user.portrait) : "") + "'><span class='tab-label'>" + App.translations.Profile + "</span>";
		account.addEventListener("click", function () {
			AccountMenu.show(self.mAppListener, account);
		});
		nav.appendChild(account);

		return nav;
	}

	// A link that switches the page in place on a plain click, and behaves as a link otherwise.
	private link(page: string, href: string, data: any): HTMLAnchorElement {
		const self = this;
		const a = document.createElement("a");
		a.href = href;
		a.dataset.page = page;
		a.addEventListener("click", function (e) {
			if (e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) {
				return;
			}
			e.preventDefault();
			self.mAppListener.showPage(page.startsWith("list:") ? "list" : page, data);
		});
		return a;
	}

	private navItem(page: string, href: string, icon: string, label: string): HTMLAnchorElement {
		const a = this.link(page, href, null);
		a.className = "nav-item";
		a.innerHTML = "<span class='material-symbols-outlined'>" + icon + "</span>";
		a.appendChild(document.createTextNode(label));
		return a;
	}

	private tab(page: string, href: string, icon: string, label: string): HTMLAnchorElement {
		const a = this.link(page, href, null);
		a.className = "tab";
		a.innerHTML = "<span class='material-symbols-outlined'>" + icon + "</span><span class='tab-label'></span>";
		(a.querySelector(".tab-label") as HTMLElement).textContent = label;
		return a;
	}

	// The checklists under Checklists, each with how far the user has come.
	private loadLists(): void {
		const self = this;
		new HttpRequest("GET", "/v1/users/me").process(function (status: any, data: any) {
			const d = data as Responses.UsersMe;
			self.mLists.innerHTML = "";
			(d.list_reports || []).forEach(e => {
				const seen = Utils.numberOfUniqueSpeciesInReports(e.reports || []);
				const total = e.list.species.length;
				const a = self.link("list:" + e.list.uuid, "/list/" + e.list.uuid, e.list);
				a.className = "nav-list";
				a.style.setProperty("--dot", e.list.color !== "" ? e.list.color : "var(--blue)");

				const name = document.createElement("span");
				name.className = "nav-list-name";
				name.textContent = App.getTranslation(e.list.name);
				a.appendChild(name);

				const count = document.createElement("span");
				count.className = "nav-list-count";
				count.textContent = seen + "/" + total;
				a.appendChild(count);

				const bar = document.createElement("span");
				bar.className = "nav-list-bar";
				bar.innerHTML = "<span style='width: " + (total > 0 ? seen / total * 100 : 0) + "%'></span>";
				a.appendChild(bar);

				self.mLists.appendChild(a);
			});

			const add = self.link("checklists", "/checklists", null);
			add.className = "nav-list-add";
			add.innerHTML = "<span class='material-symbols-outlined'>add</span>";
			add.appendChild(document.createTextNode(App.translations.NewChecklist));
			self.mLists.appendChild(add);

			self.setActive(self.mActive);
		});
	}

	// Highlights the page on screen. Checklists are "list:<uuid>"; they also light up Checklists, and
	// the account pages light up the profile tab.
	public setActive(page: string): void {
		this.mActive = page;
		const accountPages = ["settings", "tags", "contact", "donation", "admin"];
		this.mContainer.querySelectorAll("[data-page]").forEach(el => {
			const e = el as HTMLElement;
			const p = e.dataset.page || "";
			let active = p === page;
			if (e.classList.contains("nav-list-add")) {
				active = false;
			} else if (p === "checklists" && page.startsWith("list:")) {
				active = true;
			} else if (p === "account" && accountPages.indexOf(page) >= 0) {
				active = true;
			}
			e.classList.toggle("active", active);
			if (active && p === page) {
				e.setAttribute("aria-current", "page");
			} else {
				e.removeAttribute("aria-current");
			}
		});
	}
}
