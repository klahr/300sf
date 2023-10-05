/// <reference path="Page.ts"/>
/// <reference path="ui/Image.ts"/>
/// <reference path="ProfilePopUp.ts"/>
/// <reference path="HttpRequest.ts"/>

class ManageCheckListsPage extends Page {
	private mContent: Ui.Div;
	private mTitle: HTMLHeadingElement;
	private mChecklists: Ui.Div;

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
		this.attach(profile);

		this.mTitle = document.createElement("h1");
		this.mTitle.innerHTML = App.translations.ManageChecklists;
		this.container.appendChild(this.mTitle);

		const checklistsLbl = document.createElement("h2");
		checklistsLbl.innerHTML = App.translations.Checklists;
		this.container.appendChild(checklistsLbl);

		this.mChecklists = new Ui.Div();
		this.mChecklists.style.position = "relative";
		this.attach(this.mChecklists);

		const newChecklistsLbl = document.createElement("h2");
		newChecklistsLbl.innerHTML = "Create new checklist";
		this.container.appendChild(newChecklistsLbl);

		const nameLbl = document.createElement("h3");
		nameLbl.className = "title";
		nameLbl.innerHTML = "List name";
		this.container.appendChild(nameLbl);
		const name = new Ui.Input("text");
		this.attach(name);

		const colorLbl = document.createElement("h3");
		colorLbl.className = "title";
		colorLbl.innerHTML = "Color";
		this.container.appendChild(colorLbl);

		const colorSelect = new Ui.DivSelect();
		colorSelect.addOption("", "#ff00ff").style.backgroundColor = "#ff00ff";
		colorSelect.addOption("", "#00ff00").style.backgroundColor = "#00ff00";
		colorSelect.addOption("", "#00ffff").style.backgroundColor = "#00ffff";
		this.attach(colorSelect);
		colorSelect.select("#00ff00");

		const createButton = document.createElement("button");
		createButton.innerHTML = "Create Checklist";
		createButton.style.width = "auto";
		createButton.style.cssFloat = "right";
		createButton.style.borderRadius = "8px";
		createButton.style.backgroundColor = "var(--success-color)";
		createButton.disabled = true;
		this.container.appendChild(createButton);
	}

	public get id(): string {
		return "checklists";
	}

	public setVisible(visible: boolean): void {
		super.setVisible(visible);
		this.refresh();
	}

	private refresh(): void {
		const self = this;
		const r = new HttpRequest("GET", "/v1/lists/list");
		r.process(function(status: any, data: any) {
			self.mChecklists.clear();
			data.result.forEach(e => {
				const c = document.createElement("p");
				c.className = "entry";
				c.innerHTML = "<span style='font-size: 1rem; color: #2782ac;' class='material-symbols-outlined'>edit_note</span> " + App.getTranslation(e.name);
				self.mChecklists.container.appendChild(c);
				c.addEventListener("click", function() {
					new Ui.MessageBox("Edit List", "Here you will edit list.", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--success-color)", function() {});
				});
			});
		});
	}
}