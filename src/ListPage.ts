/// <reference path="ui/Table.ts"/>
/// <reference path="ui/ProgressBar.ts"/>

class ListPage extends Page {
	private mContent: Ui.Div;
	private mTitle: HTMLHeadingElement;
	private mTable: Ui.Table;
	private mProgress: Ui.ProgressBar;
	private mProgressLbl: HTMLParagraphElement;

	constructor(appListener: IAppListener) {
		super(appListener);

		this.mContent = new Ui.Div();
		this.mContent.style.position = "relative";
		this.mContent.style.width = "100%";
		this.attach(this.mContent);

		const profile = new Ui.Image("img/99249498-452f-4b0b-a1f5-45902f7ffe48.png", "48px", "48px");
		profile.style.left = "unset";
		profile.style.right = "var(--medium-distance)";
		profile.style.top = "var(--small-distance)";
		profile.style.borderRadius = "100%";
		profile.container.addEventListener("click", function() {
			new Ui.PopUp("img/300sf.svg", "Sidan finns ännu inte.", function() {
			}, null);
		});

		this.mTitle = document.createElement("h1");
		this.container.appendChild(this.mTitle);

		this.mProgressLbl = document.createElement("p");
		this.mProgressLbl.innerHTML = "4/300";
		this.mProgressLbl.style.margin = "0";
		this.mProgressLbl.style.fontWeight = "bold";
		this.mProgressLbl.style.cssFloat = "right";
		this.container.appendChild(this.mProgressLbl);

		this.mProgress = new Ui.ProgressBar();
		this.mProgress.container.style.position = "relative";
		this.attach(this.mProgress);

		this.mTable = new Ui.Table();
		this.attach(this.mTable);

		this.attach(profile);
	}

	public get id(): string {
		return "list-page";
	}

	public setData(data: any): void {
		super.setData(data);

		this.mTitle.innerHTML = "";
		this.mTable.clear();
		this.mProgress.setProgress(0);
		this.mProgressLbl.innerHTML = "";

		const self = this;

		const r = new HttpRequest("GET", "/v1/lists/" + (data as Models.List).uuid + "/reports");
		r.process(function(status: any, data: any) {
			self.mTitle.innerHTML = App.getTranslation((self.data as Models.List).name);
			const reports = data.result as Models.Report[];
			self.data.species.forEach(e => {
				const entry = document.createElement("div");
				entry.className = "list-entry";

				const name = document.createElement("h3");
				name.innerHTML = App.getTranslation(e.name);
				name.style.cursor = "pointer";
				entry.appendChild(name);
				name.addEventListener("click", function() {
					new Ui.PopUp("img/300sf.svg", e.name.sv, function() {
					}, null);
				});

				self.mProgress.setProgress(reports.length / self.data.species.length * 100);
				self.mProgressLbl.innerHTML = reports.length + "/" + self.data.species.length;

				const subName = document.createElement("p");
				subName.innerHTML = e.name.la;
				entry.appendChild(subName);

				const check = document.createElement("span");
				check.className = "material-symbols-outlined";
				check.innerHTML = "check_circle";
				check.style.color = "var(--secondary-color)";
				entry.appendChild(check);

				let found = false;
				reports.forEach(r => {
					if (r.species.uuid === e.uuid) {
						found = true;
						return;
					}
				});

				if (found) {
					name.style.color = "var(-success-color)";
					check.style.color = "var(--success-color)";
				}

				self.mTable.addRow([entry]);
			});
		}, function(status: any, data: any) {
			console.log(data);
		});
	}
}
