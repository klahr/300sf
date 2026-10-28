/// <reference path="App.ts"/>
/// <reference path="ui/Sheet.ts"/>
/// <reference path="ProfilePopUp.ts"/>
/// <reference path="LocalStorage.ts"/>

// Everything about the account and the service, in one place: opened from the account card in the
// sidebar, the avatar in the top bar and the profile tab. A dropdown on large screens, a sheet on phones.
namespace AccountMenu {
	interface Item {
		icon: string;
		label: string;
		action: () => void;
		danger?: boolean;
	}

	function items(appListener: IAppListener): Array<Array<Item>> {
		const account: Array<Item> = [
			{ icon: "account_circle", label: App.translations.Profile, action: () => openProfilePopUp() },
			{ icon: "settings", label: App.translations.Settings, action: () => appListener.showPage("settings", null) },
		];
		// The tab bar has no room for competitions; the sidebar on wider screens has them.
		if (window.matchMedia("(max-width: 1023px)").matches) {
			account.unshift({ icon: "emoji_events", label: App.translations.Competitions, action: () => appListener.showPage("competitions", null) });
		}
		const scope = new LocalStorage().get("scope");
		if (scope && scope.includes("admin")) {
			account.push({ icon: "shield_person", label: App.translations.Admin, action: () => appListener.showPage("admin", null) });
		}
		return [
			account,
			[
				{ icon: "help", label: App.translations.HelpAndContact, action: () => appListener.showPage("contact", null) },
				{ icon: "volunteer_activism", label: App.translations.Donation, action: () => appListener.showPage("donation", null) },
				{ icon: "description", label: App.translations.TermsOfService, action: () => window.open("/terms", "_blank") },
				{ icon: "privacy_tip", label: App.translations.PrivacyPolicy, action: () => window.open("/privacy-policy", "_blank") },
			],
			[
				{ icon: "logout", label: App.translations.LogOut, action: () => appListener.logOut(), danger: true },
			],
		];
	}

	// Calls chosen with the item's action, so the caller can close itself before running it.
	function list(groups: Array<Array<Item>>, chosen: (action: () => void) => void): HTMLElement {
		const menu = document.createElement("div");
		menu.className = "account-items";
		menu.setAttribute("role", "menu");
		groups.forEach(group => {
			const g = document.createElement("div");
			g.className = "account-group";
			group.forEach(item => {
				const b = document.createElement("button");
				b.className = item.danger ? "account-item is-danger" : "account-item";
				b.setAttribute("role", "menuitem");
				b.innerHTML = "<span class='material-symbols-outlined'>" + item.icon + "</span>";
				b.appendChild(document.createTextNode(item.label));
				b.addEventListener("click", function () {
					chosen(item.action);
				});
				g.appendChild(b);
			});
			menu.appendChild(g);
		});
		return menu;
	}

	let open: HTMLElement | null = null;

	export function show(appListener: IAppListener, anchor: HTMLElement): void {
		if (window.matchMedia("(min-width: 1024px)").matches) {
			showDropdown(appListener, anchor);
		} else {
			showSheet(appListener);
		}
	}

	function showSheet(appListener: IAppListener): void {
		const sheet = new Ui.Sheet({ title: App.user ? App.user.display_name : App.translations.Account, subtitle: App.user ? "@" + App.user.username : "", size: "compact" });
		sheet.header.classList.add("with-portrait");
		const portrait = document.createElement("img");
		portrait.alt = "";
		if (App.user) {
			portrait.src = "/img/portraits/" + App.user.portrait;
		}
		sheet.header.insertBefore(portrait, sheet.names);
		sheet.section().appendChild(list(items(appListener), function (action: () => void) {
			sheet.close();
			Ui.PopUp.whenHistorySettled(action);
		}));
		sheet.open();
	}

	function showDropdown(appListener: IAppListener, anchor: HTMLElement): void {
		// A second click on the anchor closes it.
		if (open) {
			closeOpen();
			return;
		}
		const menu = document.createElement("div");
		menu.className = "account-dropdown";
		menu.tabIndex = -1;
		menu.appendChild(list(items(appListener), function (action: () => void) {
			close();
			action();
		}));
		document.body.appendChild(menu);
		open = menu;
		anchor.setAttribute("aria-expanded", "true");

		// Above the anchor when it is in the lower half of the screen (the sidebar card), else below.
		const r = anchor.getBoundingClientRect();
		if (r.top > window.innerHeight / 2) {
			menu.style.bottom = (window.innerHeight - r.top + 8) + "px";
			menu.style.left = Math.max(8, r.left) + "px";
		} else {
			menu.style.top = (r.bottom + 8) + "px";
			menu.style.right = Math.max(8, window.innerWidth - r.right) + "px";
		}

		const outside = function (e: Event) {
			if (!menu.contains(e.target as Node) && !anchor.contains(e.target as Node)) {
				close();
			}
		};
		const key = function (e: KeyboardEvent) {
			if (e.key === "Escape") {
				close();
				anchor.focus();
			}
		};
		setTimeout(function () {
			document.addEventListener("pointerdown", outside);
		}, 0);
		document.addEventListener("keydown", key);
		// The menu takes focus so Tab moves into it, without a focus ring on an item after a click.
		menu.focus();

		function close() {
			document.removeEventListener("pointerdown", outside);
			document.removeEventListener("keydown", key);
			anchor.setAttribute("aria-expanded", "false");
			menu.classList.add("closing");
			setTimeout(() => menu.remove(), 150);
			open = null;
		}
		closeOpen = close;
	}

	let closeOpen: () => void = function () { };
}
