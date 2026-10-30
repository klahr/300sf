/// <reference path="HttpRequest.ts"/>
/// <reference path="models/Announcement.ts"/>

// The unread news posts, shared by everything that shows them: the bell in the
// top bar, News in the sidebar and the posts on the home page.
namespace News {
	let unread: Array<Models.NewsPost> = [];
	let checkedAt = 0;
	const listeners: Array<(posts: Array<Models.NewsPost>) => void> = [];

	// Posts are few and rarely new, so a check every few minutes is enough.
	const recheckAfterMs = 5 * 60 * 1000;

	export function onChange(listener: (posts: Array<Models.NewsPost>) => void): void {
		listeners.push(listener);
		listener(unread);
	}

	function set(posts: Array<Models.NewsPost>): void {
		unread = posts;
		listeners.forEach(l => l(unread));
	}

	export function refresh(force = false): void {
		if (!App.isLoggedIn() || (!force && Date.now() - checkedAt < recheckAfterMs)) {
			return;
		}
		checkedAt = Date.now();
		new HttpRequest("GET", "/v1/announcements/list", null).process((status: any, data: any) => {
			set((data.data || []) as Array<Models.NewsPost>);
		});
	}

	export function markRead(uuid: string): void {
		set(unread.filter(p => p.uuid !== uuid));
		new HttpRequest("PUT", "/v1/announcements/" + uuid + "/read").process();
	}

	export function markAllRead(): void {
		if (unread.length === 0) {
			return;
		}
		set([]);
		new HttpRequest("PUT", "/v1/announcements/read_all").process();
	}

	// A count badge that follows the number of unread posts.
	export function badge(): HTMLSpanElement {
		const b = document.createElement("span");
		b.className = "news-badge";
		onChange(function (posts) {
			b.textContent = posts.length === 0 ? "" : posts.length > 9 ? "9+" : String(posts.length);
		});
		return b;
	}

	// The bell that opens the news feed.
	export function bell(appListener: IAppListener): HTMLButtonElement {
		const b = document.createElement("button");
		b.className = "icon-button news-bell";
		b.innerHTML = "<span class='material-symbols-outlined'>notifications</span>";
		b.appendChild(badge());
		onChange(function (posts) {
			b.setAttribute("aria-label", posts.length > 0 ? App.translations.UnreadNews + ": " + posts.length : App.translations.News);
			b.classList.toggle("has-unread", posts.length > 0);
		});
		b.addEventListener("click", function () {
			appListener.showPage("news", null);
		});
		return b;
	}

	// Coming back to the app is a good time to look for new posts.
	document.addEventListener("visibilitychange", function () {
		if (document.visibilityState === "visible") {
			refresh();
		}
	});
}
