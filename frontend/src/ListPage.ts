/// <reference path="ui/Table.ts"/>
/// <reference path="ui/ProgressBar.ts"/>
/// <reference path="ui/Image.ts"/>
/// <reference path="utils/Report.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="Page.ts"/>
/// <reference path="models/Order.ts"/>
/// <reference path="models/Report.ts"/>
/// <reference path="models/List.ts"/>
/// <reference path="SpeciesPopUp.ts"/>
/// <reference path="OrderPopUp.ts"/>
/// <reference path="ShareChecklist.ts"/>

class ListPage extends Page {
	private mTitle: HTMLHeadingElement;
	private mTable: Ui.Table;
	private mProgress: Ui.ProgressBar;
	private mProgressLbl: HTMLParagraphElement;
	private mFilterInput: Ui.Input;
	private mFilterTitle: HTMLParagraphElement;
	private mShownUuid = "";

	constructor(appListener: IAppListener) {
		super(appListener);

		const self = this;

		this.addProfile();

		const header = document.createElement("div");
		header.className = "page-header";
		this.container.appendChild(header);

		this.mTitle = document.createElement("h1");
		header.appendChild(this.mTitle);

		const download = document.createElement("button");
		download.className = "icon-button";
		download.title = App.translations.DownloadCsv;
		download.setAttribute("aria-label", App.translations.DownloadCsv);
		download.innerHTML = "<span class='material-symbols-outlined'>download</span>";
		download.addEventListener("click", function (event) {
			event.stopPropagation();
			const d = {
				"list_uuid": self.mData.uuid,
				"format": "csv",
				"timezone_offset": Utils.getTimezoneOffset(),
				"time_zone": Utils.getTimeZone()
			};
			const r = new HttpRequest("PUT", "/v1/reports/export", d);
			r.setTextResponse();
			Ui.setBusy(download, true);
			r.process(function (status: any, data: any) {
				// The browser drops the byte order mark when reading the response; Excel needs it to read
				// the file as UTF-8, so put it back.
				const bom = String.fromCharCode(0xFEFF);
				const csv = String(data);
				DownloadFile(Utils.exportFileName(App.getTranslation((self.data as Models.List).name), "csv"), csv.startsWith(bom) ? csv : bom + csv, "text/csv");
			}, function () {
				new Ui.MessageBox(App.translations.Error, App.translations.SomethingWentWrong, "/img/messagebox-256.webp", "var(--error-color)", function () { });
			}, function () {
				Ui.setBusy(download, false);
			});
		});
		const share = document.createElement("button");
		share.className = "icon-button";
		share.title = App.translations.ShareChecklist;
		share.setAttribute("aria-label", App.translations.ShareChecklist);
		share.innerHTML = "<span class='material-symbols-outlined'>share</span>";
		share.addEventListener("click", function (event) {
			event.stopPropagation();
			ShareChecklist.open(self.data as Models.List, function () {
				self.mShownUuid = (self.data as Models.List).uuid;
				Page.setUrl(self.path);
			});
		});
		header.appendChild(share);
		header.appendChild(download);

		const summary = document.createElement("div");
		summary.className = "progress-summary";
		this.container.appendChild(summary);

		this.mProgressLbl = document.createElement("p");
		summary.appendChild(this.mProgressLbl);

		this.mProgress = new Ui.ProgressBar();
		this.mChildren.push(this.mProgress);
		summary.appendChild(this.mProgress.container);

		this.addClass("wide");
		this.mTable = new Ui.Table();
		this.mTable.addClass("species-table");
		this.attach(this.mTable);

		const filter = new Ui.Div
		filter.addClass("filter");
		this.attach(filter);

		this.mFilterTitle = document.createElement("p");
		filter.container.appendChild(this.mFilterTitle);

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
			const listEntries = self.mTable.container.getElementsByTagName("tr");
			const value = this.value.toLowerCase();
			let previousOrder: HTMLTableRowElement | null = null;
			for (let i = 0; i < listEntries.length; ++i) {
				const h3s = listEntries[i].getElementsByTagName("h3");
				if (h3s.length > 0) {
					for (let j = 0; j < h3s.length; ++j) {
						if (h3s[j].innerHTML.toLocaleLowerCase().indexOf(value) > -1) {
							(listEntries[i] as HTMLElement).style.display = "";
							if (previousOrder) {
								previousOrder.style.display = "";
								previousOrder = null;
							}
						} else {
							(listEntries[i] as HTMLElement).style.display = "none";
						}
					}
				} else {
					previousOrder = listEntries[i];
					previousOrder.style.display = "none";
				}
			}
		});
	}

	public get id(): string {
		return "list";
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

	protected get path(): string {
		return "/" + this.id + "/" + this.mData.uuid;
	}

	public setData(data: any): void {
		this.mData = data;
		const self = this;
		if (data.name === undefined) {
			// TODO: Optimize this call that returns more then necessary.
			const r = new HttpRequest("GET", "/v1/lists/list");
			const superSetData = super.setData;
			r.process(function (status: any, data: any) {
				(data.result || []).forEach(e => {
					if (e.uuid === self.data.uuid) {
						superSetData.call(self, e);
						self.updateLists();
						return;
					}
				});
			});
		} else {
			this.updateLists();
		}
	}

	private updateLists(): void {
		Page.setUrl(this.path);
		const scrollPostion = this.container.scrollTop;

		// Reloading the list on screen (after a sighting was reported) keeps what is shown and the
		// search until the new rows arrive; only another list starts from placeholders.
		const sameList = this.mShownUuid === (this.data as Models.List).uuid;
		this.mShownUuid = (this.data as Models.List).uuid;
		if (!sameList) {
			this.mFilterInput.value = "";
			this.mTitle.innerHTML = "";
			this.mTable.showPlaceholders(12);
			this.mProgress.setProgress(0);
			this.mProgressLbl.innerHTML = "";
		}
		const self = this;

		const r = new HttpRequest("GET", "/v1/lists/" + (this.data as Models.List).uuid + "/reports");
		r.process(function (status: any, data: any) {
			self.mTitle.textContent = App.getTranslation((self.data as Models.List).name);
			self.mFilterTitle.textContent = self.mTitle.textContent;
			self.mTable.clear();
			const reports = (data.result || []) as Models.Report[];
			const seenSpecies = new Set<string>();
			reports.forEach(r => {
				seenSpecies.add(r.species.uuid);
			});
			const uniqueSpeciesInReports = Utils.numberOfUniqueSpeciesInReports(reports);
			const total = self.data.species.length;
			self.mProgress.setProgress(total > 0 ? uniqueSpeciesInReports / total * 100 : 0);
			self.mProgressLbl.innerHTML = "<strong>" + uniqueSpeciesInReports + "</strong> / " + total + "<em>" + (total > 0 ? Math.floor(uniqueSpeciesInReports / total * 100) : 0) + "%</em>";
			let lastOrder: Models.Order;
			self.data.species.forEach(e => {
				if (!lastOrder || lastOrder.uuid !== e.order.uuid) {
					lastOrder = e.order;

					const entry = document.createElement("div");
					entry.className = "list-entry order";

					const name = document.createElement("p");
					name.innerHTML = App.getTranslation(e.order.name);
					name.classList.add("order-name");
					entry.appendChild(name);
					name.addEventListener("click", function () {
						openOrderPopUp(e.order);
					});

					const subName = document.createElement("p");
					subName.innerHTML = e.order.name.la;
					entry.appendChild(subName);
					self.mTable.addRow([entry]).classList.add("order-row");
				}
				const entry = document.createElement("div");
				entry.className = "list-entry";

				const name = document.createElement("h3");
				name.innerHTML = App.getTranslation(e.name);
				entry.appendChild(name);
				entry.addEventListener("click", function () {
					openSpeciesPopUp(e, self.data, (boolean) => { });
				});

				const subName = document.createElement("p");
				subName.innerHTML = e.name.la;
				entry.appendChild(subName);

				const check = document.createElement("span");
				check.className = "material-symbols-outlined";
				check.innerHTML = "check_circle";
				check.classList.add("seen-mark");
				check.setAttribute("role", "button");
				check.setAttribute("aria-label", App.translations.SubmitSighting);
				check.tabIndex = 0;
				check.addEventListener("keydown", function (event) {
					if (event.key === "Enter") {
						check.click();
					}
				});
				check.addEventListener("click", function (event) {
					event.stopPropagation();
					openReportPopUp(e, self.data);
				});
				entry.appendChild(check);

				if (seenSpecies.has(e.uuid)) {
					entry.classList.add("seen");
				}

				self.mTable.addRow([entry]);
			});
			if (self.mFilterInput.value !== "") {
				self.mFilterInput.container.dispatchEvent(new Event("keyup"));
			}
			self.container.scrollTop = scrollPostion;
		}, function (status: any, data: any) {
		});
	}
}
