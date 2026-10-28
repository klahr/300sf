/// <reference path="ui/Table.ts"/>
/// <reference path="Page.ts"/>
/// <reference path="App.ts"/>

class ReportsPage extends Page {
	private mTitle: HTMLHeadingElement;
	private mTable: Ui.Table;
	private mProgressLbl: HTMLParagraphElement;
	private mFilterInput: Ui.Input;
	private mFilterTags: Ui.Div;

	constructor(appListener: IAppListener) {
		super(appListener);

		this.mTitle = document.createElement("h1");
		this.mTitle.innerHTML = App.translations.Sightings;
		this.container.appendChild(this.mTitle);

		this.mProgressLbl = document.createElement("p");
		this.mProgressLbl.style.margin = "0";
		this.mProgressLbl.style.fontWeight = "bold";
		this.mProgressLbl.style.cssFloat = "right";
		this.mProgressLbl.style.color = "var(--success-color)";
		this.container.appendChild(this.mProgressLbl);

		this.mTable = new Ui.Table();
		this.attach(this.mTable);

		this.addProfile();

		const filter = new Ui.Div
		filter.addClass("filter");
		this.attach(filter);

		const self = this;
		this.mFilterInput = new Ui.Input("text");
		this.mFilterInput.setPlaceHolder(App.translations.Search + "...");
		filter.attach(this.mFilterInput);
		this.mFilterInput.container.addEventListener("focus", function () {
			self.mFilterInput.select();
		});
		(this.mFilterInput.container as HTMLInputElement).addEventListener("keydown", function (e) {
			if (e.key === "Enter") {
				this.blur();
			}
		});
		(this.mFilterInput.container as HTMLInputElement).addEventListener("keyup", function () {
			self.filter();
		});

		this.mFilterTags = new Ui.Div();
		this.mFilterTags.style.position = "relative";
		this.mFilterTags.style.margin = "var(--small-distance)";
		filter.attach(this.mFilterTags);
	}


	private filter(): void {
		let count = 0;
		const tags = Array();
		const allTags = this.mFilterTags.container.getElementsByClassName("tag");
		for (let i = 0; i < allTags.length; ++i) {
			if (allTags[i].classList.contains("selected")) {
				tags.push(allTags[i].getAttribute("uuid"));
			}
		}
		const listEntries = this.mTable.container.getElementsByTagName("tr");
		const value = this.mFilterInput.value.toLowerCase();
		for (let i = 0; i < listEntries.length; ++i) {
			if (tags.length > 0) {
				let ok = false;
				const entryTags = listEntries[i].getElementsByClassName("tag");
				for (let j = 0; j < tags.length; ++j) {
					for (let k = 0; k < entryTags.length; ++k) {
						if (entryTags[k].getAttribute("uuid") === tags[j]) {
							ok = true;
							break;
						}
					}
					if (ok) {
						break;
					}
				}
				if (!ok) {
					listEntries[i].style.display = "none";
					continue;
				}
			}
			const h3s = listEntries[i].getElementsByTagName("h3");
			if (h3s.length > 0) {
				for (let j = 0; j < h3s.length; ++j) {
					if (h3s[j].innerHTML.toLocaleLowerCase().indexOf(value) > -1) {
						(listEntries[i] as HTMLElement).style.display = "";
						count++;
					} else {
						(listEntries[i] as HTMLElement).style.display = "none";
					}
				}
			}
		}
		this.mProgressLbl.innerHTML = count.toString();
	}

	public get id(): string {
		return "reports";
	}

	public requireSignIn(): boolean {
		return true;
	}

	public onBecomeVisible(): void {
		super.onBecomeVisible();
		this.refresh();
	}

	private refresh(): void {
		this.mProfile.setImage("/img/portraits/" + App.user.portrait);
	}

	public setData(data: any): void {
		super.setData(data);
		this.updateLists();
	}

	private updateLists(): void {
		const self = this;
		const t = new HttpRequest("GET", "/v1/tags/list");
		t.process(function (status: any, data: any) {
			self.mFilterTags.clear();
			data.result.forEach(e => {
				const tag = document.createElement("div");
				tag.className = "tag";
				tag.setAttribute("uuid", e.uuid);
				tag.addEventListener("click", () => {
					if (tag.classList.contains("selected")) {
						tag.classList.remove("selected");
					} else {
						tag.classList.add("selected");
					}
					self.filter();
				});
				if (e.color !== "") {
					tag.style.backgroundColor = e.color;
				}
				const c = document.createElement("p");
				c.textContent = e.name;
				tag.appendChild(c);
				self.mFilterTags.container.appendChild(tag);
			});
		});

		const r = new HttpRequest("GET", "/v1/reports/list");
		r.process(function (status: any, data: any) {
			self.mTable.clear();
			const reports = data.result as Models.Report[];
			self.mProgressLbl.innerHTML = reports.length.toString();
			reports.forEach(e => {
				const entry = document.createElement("div");
				entry.addEventListener("click", function () {
					openEditReportPopUp(e);
				});
				entry.className = "list-entry";

				const name = document.createElement("h3");
				name.innerHTML = App.getTranslation(e.species.name);
				name.style.cursor = "pointer";
				name.classList.add("species-name");
				name.style.width = "fit-content";
				entry.appendChild(name);
				name.addEventListener("click", function (event) {
					event.stopPropagation();
					openSpeciesPopUp(e.species, null, (speciesAdded: boolean) => {
						if (speciesAdded) {
							self.refresh();
						}
					});
				});

				const subName = document.createElement("p");
				subName.innerHTML = e.species.name.la;
				entry.appendChild(subName);

				const dateInterval = document.createElement("p");
				if (Utils.isValidDate(e.from)) {
					dateInterval.innerHTML = Utils.toYYYY_MM_DD_hh_mm(e.from);
				}
				if (Utils.isValidDate(e.until)) {
					dateInterval.innerHTML += " - " + Utils.toYYYY_MM_DD_hh_mm(e.until);
				}
				dateInterval.style.fontStyle = "unset";
				entry.appendChild(dateInterval);

				const location = document.createElement("p");
				if (e.zone) {
					location.innerHTML = App.getTranslation(e.zone.name);
				}
				location.style.fontStyle = "unset";
				entry.appendChild(location);

				if (e.comment.text !== "") {
					const comment = document.createElement("p");
					comment.classList.add("comment");
					comment.style.borderRadius = "0 12px 0 12px";
					comment.style.lineHeight = "unset";
					comment.style.padding = "8px";
					comment.style.width = "fit-content";
					comment.textContent = e.comment.text;
					entry.appendChild(comment);
				}

				if (e.tags.length > 0) {
					const tagsContainer = document.createElement("div");
					tagsContainer.style.position = "relative";
					entry.appendChild(tagsContainer);

					e.tags.forEach(e => {
						const tag = document.createElement("div");
						tag.className = "tag";
						tag.setAttribute("uuid", e.uuid);
						if (e.color !== "") {
							tag.style.backgroundColor = e.color;
						}
						const c = document.createElement("p");
						c.textContent = e.name;
						tag.appendChild(c);
						tagsContainer.appendChild(tag);
					});
				}

				self.mTable.addRow([entry]);
			});
		});
	}
}