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

class ListPage extends Page {
	private mContent: Ui.Div;
	private mTitle: HTMLHeadingElement;
	private mTable: Ui.Table;
	private mProgress: Ui.ProgressBar;
	private mProgressLbl: HTMLParagraphElement;
	private mFilterInput: Ui.Input;

	constructor(appListener: IAppListener) {
		super(appListener);

		this.mContent = new Ui.Div();
		this.mContent.style.position = "relative";
		this.mContent.style.width = "100%";
		this.attach(this.mContent);

		const profile = new Ui.Image("/img/99249498-452f-4b0b-a1f5-45902f7ffe48.png", "48px", "48px");
		profile.addClass("portrait");
		profile.container.addEventListener("click", function() {
			openProfilePopUp(function() {
				console.log("close profile");
			});
		});

		this.mTitle = document.createElement("h1");
		this.container.appendChild(this.mTitle);

		this.mProgressLbl = document.createElement("p");
		this.mProgressLbl.style.margin = "0";
		this.mProgressLbl.style.fontWeight = "bold";
		this.mProgressLbl.style.cssFloat = "right";
		this.mProgressLbl.style.color = "var(--success-color)";
		this.container.appendChild(this.mProgressLbl);

		this.mProgress = new Ui.ProgressBar();
		this.mProgress.container.style.position = "relative";
		this.attach(this.mProgress);

		this.mTable = new Ui.Table();
		this.attach(this.mTable);

		this.attach(profile);

		const filter = new Ui.Div
		filter.addClass("filter");
		this.attach(filter);

		const self = this;
		this.mFilterInput = new Ui.Input("text");
		this.mFilterInput.setPlaceHolder(App.translations.Search + "...");
		filter.attach(this.mFilterInput);
		this.mFilterInput.container.addEventListener("focus", function() {
			self.mFilterInput.select();
		});
		(this.mFilterInput.container as HTMLInputElement).addEventListener("keydown", function(e) {
			if (e.key === "Enter") {
				this.blur();
			}
		});
		(this.mFilterInput.container as HTMLInputElement).addEventListener("keyup", function() {
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

	public setData(data: any): void {
		super.setData(data);
		const self = this;
		if (data.name === undefined) {
			// TODO: Optimize this call that returns more then necessary.
			const r = new HttpRequest("GET", "/v1/lists/list");
			r.process(function(status: any, data: any) {
				data.result.forEach( e => {
					if (e.uuid === self.data.uuid) {
						self.setData(e);
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
		history.pushState({}, "", "/" + this.id + "/" + this.data.uuid);

		this.mFilterInput.value = "";
		this.mTitle.innerHTML = "";
		this.mTable.clear();
		this.mProgress.setProgress(0);
		this.mProgressLbl.innerHTML = "";

		const self = this;

		const r = new HttpRequest("GET", "/v1/lists/" + (this.data as Models.List).uuid + "/reports");
		r.process(function(status: any, data: any) {
			console.log(self.data);
			self.mTitle.innerHTML = App.getTranslation((self.data as Models.List).name);
			const reports = data.result as Models.Report[];
			let lastOrder: Models.Order;
			self.data.species.forEach(e => {
				if (!lastOrder || lastOrder.uuid !== e.order.uuid) {
					lastOrder = e.order;

					const entry = document.createElement("div");
					entry.style.textAlign = "center";
					entry.style.backgroundColor = "var(--background-color)";
					entry.className = "list-entry";
	
					const name = document.createElement("p");
					name.innerHTML = App.getTranslation(e.order.name).toUpperCase();
					name.style.cursor = "pointer";
					name.style.fontWeight = "bold";
					name.style.fontStyle = "initial";
					name.style.color = "var(--primary-variant-color)";
					entry.appendChild(name);
					name.addEventListener("click", function() {
						const r = new HttpRequest("GET", "/v1/orders/" + e.order.uuid);
						r.process(function(status: any, data: any) {
						});
					});

					const subName = document.createElement("p");
					subName.innerHTML = e.order.name.la;
					entry.appendChild(subName);
					self.mTable.addRow([entry]);
				}
				const entry = document.createElement("div");
				entry.className = "list-entry";

				const name = document.createElement("h3");
				name.innerHTML = App.getTranslation(e.name);
				name.style.cursor = "pointer";
				entry.appendChild(name);
				name.addEventListener("click", function() {
					openSpeciesPopUp(e);
				});

				const uniqueSpeciesInReports = Utils.numberOfUniqueSpeciesInReports(reports);
				self.mProgress.setProgress(uniqueSpeciesInReports / self.data.species.length * 100);
				self.mProgressLbl.innerHTML = uniqueSpeciesInReports + "/" + self.data.species.length;

				const subName = document.createElement("p");
				subName.innerHTML = e.name.la;
				entry.appendChild(subName);

				const check = document.createElement("span");
				check.className = "material-symbols-outlined";
				check.innerHTML = "check_circle";
				check.style.color = "var(--secondary-color)";
				check.style.cursor = "pointer";
				check.addEventListener("click", function() {
					openReportPopUp(e, self.data, function() {
						self.setData(self.data);
					});
				});
				entry.appendChild(check);

				let found = false;
				reports.forEach(r => {
					if (r.species.uuid === e.uuid) {
						found = true;
						return;
					}
				});

				if (found) {
					name.style.color = "var(--success-color)";
					check.style.color = "var(--success-color)";
				}

				self.mTable.addRow([entry]);
			});
		}, function(status: any, data: any) {
			console.log(data);
		});
	}
}
