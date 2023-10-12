/// <reference path="ui/Table.ts"/>
/// <reference path="Page.ts"/>
/// <reference path="App.ts"/>

class ReportsPage extends Page {
	private mTitle: HTMLHeadingElement;
	private mTable: Ui.Table;
	private mProgressLbl: HTMLParagraphElement;

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
	}

	public get id(): string {
		return "reports";
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
		const r = new HttpRequest("GET", "/v1/reports/list");
		r.process(function(status: any, data: any) {
			self.mTable.clear();
			const reports = data.result as Models.Report[];
			self.mProgressLbl.innerHTML = reports.length.toString();
			console.log(reports);
			reports.forEach(e => {
				const entry = document.createElement("div");
				entry.addEventListener("click", function() {
					console.log("show edit report");
				});
				entry.className = "list-entry";

				const name = document.createElement("h3");
				name.innerHTML = App.getTranslation(e.species.name);
				name.style.cursor = "pointer";
				name.style.color = "var(--success-color)";
				name.style.width = "fit-content";
				entry.appendChild(name);
				name.addEventListener("click", function(event) {
					event.stopPropagation();
					openSpeciesPopUp(e.species);
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
					comment.style.backgroundColor = "#1d3944";
					comment.style.borderRadius = "0 12px 0 12px";
					comment.style.lineHeight = "unset";
					comment.style.padding = "8px";
					comment.style.width = "fit-content";
					comment.innerHTML = e.comment.text;
					entry.appendChild(comment);
				}

				self.mTable.addRow([entry]);
			});
		});
	}
}