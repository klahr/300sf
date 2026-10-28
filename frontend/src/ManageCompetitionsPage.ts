/// <reference path="Page.ts"/>
/// <reference path="App.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="CompetitionSheet.ts"/>
/// <reference path="CreateCompetitionPopUp.ts"/>

// Competitions: invitations, the ones you are in, and public ones to join. ?c=<uuid> opens one,
// which is the link people share.
class ManageCompetitionsPage extends Page {
	private mLists: HTMLDivElement;
	private mOpenedFromLink = "";

	constructor(appListener: IAppListener) {
		super(appListener);
		const self = this;

		this.addProfile();

		const header = document.createElement("div");
		header.className = "page-header";
		this.container.appendChild(header);

		const title = document.createElement("h1");
		title.textContent = App.translations.Competitions;
		header.appendChild(title);

		const create = Ui.actionButton(App.translations.NewCompetition, "add", "success compact");
		create.addEventListener("click", function () {
			openCreateCompetition(function (competition: Models.Competition) {
				self.refresh();
				Competitions.open(competition.uuid, () => self.refresh());
			});
		});
		header.appendChild(create);

		const intro = document.createElement("p");
		intro.className = "page-intro";
		intro.textContent = App.translations.CompetitionsIntro;
		this.container.appendChild(intro);

		this.mLists = document.createElement("div");
		this.mLists.className = "competition-lists";
		this.container.appendChild(this.mLists);
	}

	public get id(): string {
		return "competitions";
	}

	public requireSignIn(): boolean {
		return true;
	}

	public onBecomeVisible(): void {
		super.onBecomeVisible();
		this.mProfile.setImage("/img/portraits/" + App.user.portrait);
		this.refresh();
	}

	public setData(data: any): void {
		super.setData(data);
		const self = this;
		const uuid = data && data.c ? String(data.c) : "";
		if (uuid !== "" && uuid !== this.mOpenedFromLink) {
			this.mOpenedFromLink = uuid;
			Competitions.open(uuid, () => self.refresh());
		}
	}

	private refresh(): void {
		const self = this;
		if (this.mLists.childElementCount === 0) {
			this.mLists.appendChild(this.section(App.translations.YourCompetitions, null));
		}
		let mine: Array<Models.Competition> | null = null;
		let open: Array<Models.Competition> | null = null;
		const render = function () {
			if (mine === null || open === null) {
				return;
			}
			self.mLists.innerHTML = "";
			const invitations = mine.filter(c => c.invited);
			const yours = mine.filter(c => !c.invited);
			const joinable = open.filter(c => !c.joined && !c.invited);
			if (invitations.length > 0) {
				self.mLists.appendChild(self.section(App.translations.Invitations, invitations));
			}
			self.mLists.appendChild(self.section(App.translations.YourCompetitions, yours));
			if (joinable.length > 0) {
				self.mLists.appendChild(self.section(App.translations.OpenCompetitions, joinable));
			}
		};
		new HttpRequest("GET", "/v1/competitions/list/me").process(function (status: any, data: any) {
			// Running ones first, then those about to start, then finished ones, newest first.
			const order = { ongoing: 0, upcoming: 1, finished: 2 };
			mine = ((data.result || []) as Array<Models.Competition>).sort((a, b) => order[a.status] - order[b.status]);
			render();
		});
		new HttpRequest("GET", "/v1/competitions/list").process(function (status: any, data: any) {
			open = (data.result || []) as Array<Models.Competition>;
			render();
		});
	}

	// A heading and the cards; null shows placeholders while loading.
	private section(title: string, competitions: Array<Models.Competition> | null): HTMLElement {
		const self = this;
		const section = document.createElement("section");
		section.className = "competition-section";
		const h = document.createElement("h2");
		h.textContent = title;
		section.appendChild(h);

		const grid = document.createElement("div");
		grid.className = "competition-grid";
		section.appendChild(grid);

		if (competitions === null) {
			for (let i = 0; i < 2; ++i) {
				const p = document.createElement("div");
				p.className = "competition-card skeleton";
				grid.appendChild(p);
			}
			return section;
		}
		if (competitions.length === 0) {
			const empty = document.createElement("div");
			empty.className = "empty-state";
			empty.innerHTML = "<span class='material-symbols-outlined'>emoji_events</span>";
			const p = document.createElement("p");
			p.textContent = App.translations.NoCompetitionsYet;
			empty.appendChild(p);
			section.replaceChild(empty, grid);
			return section;
		}
		competitions.forEach(c => grid.appendChild(this.card(c, () => self.refresh())));
		return section;
	}

	private card(c: Models.Competition, changed: () => void): HTMLElement {
		const card = document.createElement("button");
		card.className = "competition-card status-" + c.status;
		card.style.setProperty("--competition-color", c.color || "var(--primary)");

		const top = document.createElement("div");
		top.className = "competition-card-top";
		const status = document.createElement("span");
		status.className = "competition-status";
		status.textContent = c.invited ? App.translations.Invitation : Competitions.statusLabel(c);
		const when = document.createElement("span");
		when.textContent = Competitions.period(c);
		top.appendChild(status);
		top.appendChild(when);
		card.appendChild(top);

		const name = document.createElement("h3");
		name.textContent = App.getTranslation(c.name);
		card.appendChild(name);

		const meta = document.createElement("p");
		meta.className = "competition-card-meta";
		const parts = [c.list_name ? App.getTranslation(c.list_name) : App.translations.AllSpecies];
		if (c.zones.length > 0) {
			parts.push(c.zones.length === 1 ? App.getTranslation(c.zones[0]) : App.translations.NumberOfCounties.replace("{0}", String(c.zones.length)));
		}
		if (c.end_on_reports) {
			parts.push(App.translations.FirstToWins.replace("{0}", String(c.end_on_reports)));
		}
		meta.textContent = parts.join(" · ");
		card.appendChild(meta);

		const bottom = document.createElement("div");
		bottom.className = "competition-card-bottom";
		const people = document.createElement("span");
		people.innerHTML = "<span class='material-symbols-outlined'>group</span>";
		people.appendChild(document.createTextNode(String(c.participants)));
		bottom.appendChild(people);
		if (c.joined) {
			const place = document.createElement("span");
			place.className = "competition-place";
			place.textContent = (c.my_count > 0 ? "#" + c.my_rank + " · " : "") + App.translations.NumberOfSpecies.replace("{0}", String(c.my_count));
			bottom.appendChild(place);
		}
		card.appendChild(bottom);

		card.addEventListener("click", function () {
			Competitions.open(c.uuid, changed);
		});
		return card;
	}
}
