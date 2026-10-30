/// <reference path="Page.ts"/>
/// <reference path="App.ts"/>
/// <reference path="News.ts"/>
/// <reference path="utils/NewsText.ts"/>

// The news feed: every published post, newest first. Posts the reader hadn't
// read are marked new for this visit; opening the feed marks them read.
class NewsPage extends Page {
	private mPosts: HTMLDivElement;
	private mMore: HTMLButtonElement;
	private mLast: string | null = null;
	// The post to show, when the page was opened from one (home page).
	private mFocus: string | null = null;

	constructor(appListener: IAppListener) {
		super(appListener);

		const self = this;

		this.addProfile();

		const title = document.createElement("h1");
		title.textContent = App.translations.News;
		this.container.appendChild(title);

		const description = document.createElement("p");
		description.className = "lead";
		description.textContent = App.translations.NewsDescription;
		this.container.appendChild(description);

		this.mPosts = document.createElement("div");
		this.mPosts.className = "news-feed";
		this.container.appendChild(this.mPosts);

		this.mMore = document.createElement("button");
		this.mMore.className = "secondary news-more";
		this.mMore.textContent = App.translations.ShowMore;
		this.mMore.hidden = true;
		this.mMore.addEventListener("click", function () {
			self.load(false);
		});
		this.container.appendChild(this.mMore);
	}

	public get id(): string {
		return "news";
	}

	public requireSignIn(): boolean {
		return true;
	}

	public setData(data: any): void {
		super.setData(data);
		this.mFocus = data && data.post ? data.post : null;
	}

	public onBecomeVisible(): void {
		super.onBecomeVisible();
		this.mProfile.setImage("/img/portraits/" + App.user.portrait);
		this.load(true);
	}

	private load(first: boolean): void {
		const self = this;
		const url = "/v1/announcements/feed" + (first || !this.mLast ? "" : "?before=" + encodeURIComponent(this.mLast));
		this.mMore.disabled = true;
		new HttpRequest("GET", url, null).process((status: any, data: any) => {
			const posts = (data.data || []) as Array<Models.NewsPost>;
			if (first) {
				self.mPosts.replaceChildren();
				self.mLast = null;
			}
			posts.forEach(p => self.mPosts.appendChild(NewsPage.post(p)));
			if (posts.length > 0) {
				self.mLast = posts[posts.length - 1].uuid;
			}
			if (first && posts.length === 0) {
				const empty = document.createElement("div");
				empty.className = "empty-state";
				empty.innerHTML = "<span class='material-symbols-outlined'>notifications</span>";
				const p = document.createElement("p");
				p.textContent = App.translations.NoNews;
				empty.appendChild(p);
				self.mPosts.appendChild(empty);
			}
			self.mMore.hidden = !data.more;
			if (first) {
				News.markAllRead();
				self.showFocus();
			}
		}, undefined, () => {
			self.mMore.disabled = false;
		});
	}

	private showFocus(): void {
		if (!this.mFocus) {
			return;
		}
		const e = this.mPosts.querySelector("[data-uuid='" + CSS.escape(this.mFocus) + "']");
		this.mFocus = null;
		if (e) {
			e.scrollIntoView({ block: "start", behavior: "smooth" });
			e.classList.add("focus");
		}
	}

	// A post as the feed, the home page and the admin's preview show it.
	public static post(p: { uuid: string, title: Models.LocalizedText, body: Models.LocalizedText, icon: string, color: string, published_at: Date, read?: boolean }): HTMLElement {
		const article = document.createElement("article");
		article.className = "news-post";
		article.dataset.uuid = p.uuid;
		if (p.color) {
			article.style.setProperty("--tone", p.color);
		}
		if (p.read === false) {
			article.classList.add("unread");
		}

		const icon = document.createElement("span");
		icon.className = "news-icon";
		icon.innerHTML = "<span class='material-symbols-outlined'></span>";
		(icon.firstChild as HTMLElement).textContent = p.icon || "campaign";
		article.appendChild(icon);

		const main = document.createElement("div");
		main.className = "news-main";
		article.appendChild(main);

		const meta = document.createElement("p");
		meta.className = "news-meta";
		const date = document.createElement("time");
		const at = new Date(p.published_at);
		date.dateTime = at.toISOString();
		date.textContent = at.toLocaleDateString(App.language, { year: "numeric", month: "long", day: "numeric" });
		meta.appendChild(date);
		if (p.read === false) {
			const chip = document.createElement("span");
			chip.className = "news-new";
			chip.textContent = App.translations.NewBadge;
			meta.appendChild(chip);
		}
		main.appendChild(meta);

		const title = document.createElement("h2");
		title.textContent = NewsText.localized(p.title);
		main.appendChild(title);

		const body = document.createElement("div");
		body.className = "news-body";
		body.appendChild(NewsText.render(NewsText.localized(p.body)));
		main.appendChild(body);

		return article;
	}
}
