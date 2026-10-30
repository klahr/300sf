/// <reference path="ui/Sheet.ts"/>
/// <reference path="ui/ConfirmBox.ts"/>
/// <reference path="App.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="models/Announcement.ts"/>
/// <reference path="Colors.ts"/>
/// <reference path="NewsPage.ts"/>
/// <reference path="News.ts"/>
/// <reference path="utils/NewsText.ts"/>

// The admin's tool for the news feed: every post with where it stands, and an
// editor with a preview of the post as readers will see it.
namespace AdminNews {
	type Status = "draft" | "scheduled" | "published" | "ended";

	function status(p: Models.AdminNewsPost): Status {
		const now = Date.now();
		if (Utils.isValidDate(p.on_hold)) {
			return "draft";
		}
		if (Utils.isValidDate(p.until) && new Date(p.until as Date).getTime() <= now) {
			return "ended";
		}
		if (Utils.isValidDate(p.from) && new Date(p.from as Date).getTime() > now) {
			return "scheduled";
		}
		return "published";
	}

	function statusLabel(s: Status): string {
		switch (s) {
			case "draft":
				return App.translations.Draft;
			case "scheduled":
				return App.translations.StatusScheduled;
			case "ended":
				return App.translations.StatusEnded;
		}
		return App.translations.StatusPublished;
	}

	function iconButton(icon: string, label: string, className = ""): HTMLButtonElement {
		const b = document.createElement("button");
		b.className = ("icon-button " + className).trim();
		b.setAttribute("aria-label", label);
		b.title = label;
		b.innerHTML = "<span class='material-symbols-outlined'>" + icon + "</span>";
		return b;
	}

	export function open(closeCallback: () => void = function () { }): void {
		const sheet = new Ui.Sheet({ title: App.translations.News, eyebrow: App.translations.Admin, size: "tall" });
		const top = sheet.section();
		const add = Ui.actionButton(App.translations.NewPost, "edit_square", "success");
		add.addEventListener("click", function () {
			edit(null, refresh);
		});
		top.appendChild(add);

		const list = document.createElement("div");
		list.className = "admin-news-list";
		sheet.section().appendChild(list);

		function refresh(): void {
			new HttpRequest("GET", "/v1/announcements/all", null).process((status: any, data: any) => {
				show((data.data || []) as Array<Models.AdminNewsPost>);
			});
			// What readers see may have changed.
			News.refresh(true);
		}

		function show(posts: Array<Models.AdminNewsPost>): void {
			list.replaceChildren();
			if (posts.length === 0) {
				const empty = document.createElement("div");
				empty.className = "empty-state";
				empty.innerHTML = "<span class='material-symbols-outlined'>campaign</span>";
				const p = document.createElement("p");
				p.textContent = App.translations.NoPosts;
				empty.appendChild(p);
				list.appendChild(empty);
				return;
			}
			posts.forEach(p => list.appendChild(row(p)));
		}

		function row(p: Models.AdminNewsPost): HTMLElement {
			const s = status(p);
			const r = document.createElement("div");
			r.className = "admin-news-row";
			r.style.setProperty("--tone", p.color || Colors.fallback);

			const icon = document.createElement("span");
			icon.className = "news-icon";
			icon.innerHTML = "<span class='material-symbols-outlined'></span>";
			(icon.firstChild as HTMLElement).textContent = p.icon || "campaign";
			r.appendChild(icon);

			const text = document.createElement("div");
			text.className = "admin-news-text";
			const title = document.createElement("p");
			title.className = "admin-news-title";
			title.textContent = NewsText.localized(p.title) || "–";
			text.appendChild(title);
			const meta = document.createElement("p");
			meta.className = "admin-news-meta";
			const chip = document.createElement("span");
			chip.className = "news-status is-" + s;
			chip.textContent = statusLabel(s);
			meta.appendChild(chip);
			let when = Utils.toYYYY_MM_DD(p.published_at);
			if (Utils.isValidDate(p.until)) {
				when += " – " + Utils.toYYYY_MM_DD(p.until as Date);
			}
			meta.appendChild(document.createTextNode(when));
			const reads = document.createElement("span");
			reads.className = "admin-news-reads";
			reads.title = App.translations.Readers;
			reads.innerHTML = "<span class='material-symbols-outlined'>visibility</span>";
			reads.appendChild(document.createTextNode(String(p.reads)));
			meta.appendChild(reads);
			text.appendChild(meta);
			r.appendChild(text);

			const actions = document.createElement("div");
			actions.className = "admin-news-actions";
			const editButton = iconButton("edit", App.translations.Edit);
			editButton.addEventListener("click", function () {
				edit(p, refresh);
			});
			actions.appendChild(editButton);

			const live = s !== "draft";
			const toggle = iconButton(live ? "unpublished" : "publish", live ? App.translations.Unpublish : App.translations.Publish);
			toggle.addEventListener("click", function () {
				toggle.disabled = true;
				new HttpRequest("PUT", "/v1/announcements/" + p.uuid + (live ? "/pause" : "/resume")).process(function () {
					refresh();
				}, undefined, function () {
					toggle.disabled = false;
				});
			});
			actions.appendChild(toggle);

			const remove = iconButton("delete", App.translations.Delete, "delete");
			remove.addEventListener("click", function () {
				new Ui.ConfirmBox(App.translations.AreYouSure, App.translations.ConfirmDeleteAnnouncement, null, "var(--error-color)", function () {
					new HttpRequest("DELETE", "/v1/announcements/" + p.uuid).process(function () {
						refresh();
					});
				}, function () { }, App.translations.Delete, App.translations.Cancel);
			});
			actions.appendChild(remove);
			r.appendChild(actions);
			return r;
		}

		sheet.open(closeCallback);
		refresh();
	}

	// "2026-10-30T14:05" for a datetime-local input, in local time.
	function toLocalInput(d: Date | null): string {
		if (!Utils.isValidDate(d)) {
			return "";
		}
		const t = new Date(d as Date);
		return Utils.toYYYY_MM_DD(t) + "T" + String(t.getHours()).padStart(2, "0") + ":" + String(t.getMinutes()).padStart(2, "0");
	}

	function fromLocalInput(value: string): string | null {
		return value === "" ? null : new Date(value).toISOString();
	}

	// The editor, for a new post (post null) or an existing one. saved is
	// called once it is stored.
	function edit(post: Models.AdminNewsPost | null, saved: () => void): void {
		const sheet = new Ui.Sheet({ title: post ? App.translations.EditPost : App.translations.NewPost, size: "tall" });

		// The text, a language at a time; English first, as it is required.
		const languages: Array<["en" | "sv" | "de", string]> = [["en", App.translations.English], ["sv", App.translations.Swedish], ["de", App.translations.German]];
		const text = sheet.section();
		const tabs = document.createElement("div");
		tabs.className = "segmented";
		tabs.setAttribute("role", "tablist");
		text.appendChild(tabs);

		const titles: { [lang: string]: HTMLInputElement } = {};
		const bodies: { [lang: string]: HTMLTextAreaElement } = {};
		const panes: { [lang: string]: HTMLElement } = {};
		let current = "en";

		languages.forEach(([lang, label]) => {
			const tab = document.createElement("button");
			tab.type = "button";
			tab.setAttribute("role", "tab");
			tab.dataset.lang = lang;
			tab.textContent = label;
			tab.addEventListener("click", function () {
				select(lang);
			});
			tabs.appendChild(tab);

			const pane = document.createElement("div");
			pane.className = "admin-news-pane";
			const title = document.createElement("input");
			title.type = "text";
			title.id = "news-title-" + lang;
			title.maxLength = 160;
			title.value = post ? (post.title[lang] || "") : "";
			sheet.field(pane, App.translations.Title, title);
			const body = document.createElement("textarea");
			body.id = "news-body-" + lang;
			body.rows = 8;
			body.maxLength = 10000;
			body.value = post ? (post.body[lang] || "") : "";
			sheet.field(pane, App.translations.Text, body);
			text.appendChild(pane);

			titles[lang] = title;
			bodies[lang] = body;
			panes[lang] = pane;
			title.addEventListener("input", update);
			body.addEventListener("input", update);
		});

		const hint = document.createElement("p");
		hint.className = "details";
		hint.textContent = App.translations.EnglishRequired + " " + App.translations.NewsFormatting;
		text.appendChild(hint);

		// The post as readers will see it, in the language being written.
		const previewSection = sheet.section(App.translations.Preview);
		const preview = document.createElement("div");
		preview.className = "news-feed admin-news-preview";
		previewSection.appendChild(preview);

		// Look.
		const look = sheet.section(App.translations.Icon + " · " + App.translations.Color);
		const icon = document.createElement("input");
		icon.type = "text";
		icon.id = "news-icon";
		icon.maxLength = 32;
		icon.placeholder = "campaign";
		icon.value = post ? post.icon : "";
		icon.addEventListener("input", update);
		sheet.field(look, App.translations.IconName, icon);

		const swatches = document.createElement("div");
		swatches.className = "color-swatches";
		swatches.setAttribute("role", "radiogroup");
		let color = post && post.color ? post.color : Colors.fallback;
		Colors.palette.forEach(c => {
			const b = document.createElement("button");
			b.type = "button";
			b.setAttribute("role", "radio");
			b.setAttribute("aria-label", c);
			b.style.backgroundColor = c;
			b.setAttribute("aria-checked", String(c === color));
			b.addEventListener("click", function () {
				color = c;
				Array.from(swatches.children).forEach(e => e.setAttribute("aria-checked", String(e === b)));
				update();
			});
			swatches.appendChild(b);
		});
		look.appendChild(swatches);

		// When.
		const when = sheet.section(App.translations.Period);
		const grid = document.createElement("div");
		grid.className = "sheet-grid";
		when.appendChild(grid);
		const from = document.createElement("input");
		from.type = "datetime-local";
		from.id = "news-from";
		from.value = post ? toLocalInput(post.from) : "";
		sheet.field(grid, App.translations.PublishFrom, from);
		const until = document.createElement("input");
		until.type = "datetime-local";
		until.id = "news-until";
		until.value = post ? toLocalInput(post.until) : "";
		sheet.field(grid, App.translations.TakeDownAfter, until);
		from.addEventListener("input", update);
		until.addEventListener("input", update);

		const error = document.createElement("p");
		error.className = "details error";
		sheet.footer.appendChild(error);

		const actions: Array<HTMLButtonElement> = [];
		if (post) {
			actions.push(sheet.addAction(Ui.actionButton(App.translations.Save, "save", "success")));
			actions[0].addEventListener("click", function () {
				save(false);
			});
		} else {
			const draft = sheet.addAction(Ui.actionButton(App.translations.SaveDraft, "draft", "secondary"));
			draft.addEventListener("click", function () {
				save(true);
			});
			const publish = sheet.addAction(Ui.actionButton(App.translations.Publish, "publish", "success"));
			publish.addEventListener("click", function () {
				save(false);
			});
			actions.push(draft, publish);
		}

		function values(of: { [lang: string]: HTMLInputElement | HTMLTextAreaElement }): Models.LocalizedText {
			return { en: of.en.value.trim(), sv: of.sv.value.trim(), de: of.de.value.trim() };
		}

		function select(lang: string): void {
			current = lang;
			Array.from(tabs.children).forEach(t => {
				const on = (t as HTMLElement).dataset.lang === lang;
				t.classList.toggle("selected", on);
				t.setAttribute("aria-selected", String(on));
			});
			Object.keys(panes).forEach(l => panes[l].hidden = l !== lang);
			update();
		}

		function update(): void {
			// Languages with a title are marked, so a missing one stands out.
			Array.from(tabs.children).forEach(t => {
				const lang = (t as HTMLElement).dataset.lang as string;
				t.classList.toggle("filled", titles[lang].value.trim() !== "" && bodies[lang].value.trim() !== "");
			});
			// Empty languages show the English their readers will see instead.
			["sv", "de"].forEach(lang => {
				titles[lang].placeholder = titles.en.value;
				bodies[lang].placeholder = bodies.en.value;
			});
			const title = titles[current].value.trim() || titles.en.value.trim();
			const body = bodies[current].value.trim() || bodies.en.value.trim();
			const one = function (s: string): Models.LocalizedText {
				return { en: s, sv: s, de: s };
			};
			preview.replaceChildren(NewsPage.post({
				uuid: "",
				title: one(title),
				body: one(body),
				icon: icon.value.trim() || "campaign",
				color: color,
				published_at: from.value !== "" ? new Date(from.value) : new Date(),
				read: false,
			}));

			let problem = "";
			if (titles.en.value.trim() === "" || bodies.en.value.trim() === "") {
				problem = App.translations.EnglishRequired;
			} else if (from.value !== "" && until.value !== "" && new Date(until.value) <= new Date(from.value)) {
				problem = App.translations.EndBeforeStart;
			}
			error.textContent = problem;
			actions.forEach(a => a.disabled = problem !== "");
		}

		function save(draft: boolean): void {
			const d = {
				title: values(titles),
				body: values(bodies),
				icon: icon.value.trim(),
				color: color,
				from: fromLocalInput(from.value),
				until: fromLocalInput(until.value),
				draft: draft,
			};
			actions.forEach(a => Ui.setBusy(a, true));
			const request = post ? new HttpRequest("PUT", "/v1/announcements/" + post.uuid, d) : new HttpRequest("POST", "/v1/announcements", d);
			request.process(function () {
				sheet.close();
				saved();
			}, function (status: any, data: Responses.GeneralErrorResponse) {
				error.textContent = data && data.message ? data.message : App.translations.SomethingWentWrong;
			}, function () {
				actions.forEach(a => Ui.setBusy(a, false));
			});
		}

		select("en");
		sheet.open();
		titles.en.focus();
	}
}

function openAdminAnnouncementsPopUp(closeCallback: () => void = function () { }): void {
	AdminNews.open(closeCallback);
}
