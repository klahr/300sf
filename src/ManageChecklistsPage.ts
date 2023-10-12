/// <reference path="Page.ts"/>
/// <reference path="ui/Image.ts"/>
/// <reference path="ui/DivSelect.ts"/>
/// <reference path="ProfilePopUp.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="requests/Lists.ts"/>
/// <reference path="models/Visibility.ts"/>

class ManageCheckListsPage extends Page {
	private mTitle: HTMLHeadingElement;
	private mChecklists: Ui.Div;
	private mTemplateSelect: Ui.Select;
	private mTemplateDescription: HTMLParagraphElement;
	private mCreateButton: HTMLButtonElement;
	private mName: Ui.Input;

	constructor(appListener: IAppListener) {
		super(appListener);

		this.addProfile();

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
		newChecklistsLbl.innerHTML = App.translations.New;
		this.container.appendChild(newChecklistsLbl);

		this.mName = new Ui.Input("text");
		this.mName.setPlaceHolder(App.translations.Name);
		this.mName.container.addEventListener("input", function() {
			self.updateCreateButtonStatus();
		});
		this.attach(this.mName);

		const self = this;
		this.mTemplateSelect = new Ui.Select(App.translations.SelectTemplate + "...", App.translations.Search, function(value: string | null) {
			self.updateCreateButtonStatus();
			if (value) {
				const o = self.mTemplateSelect.getOption(value);
				if (o) {
					const description = o.getAttribute("description");
					self.mTemplateDescription.innerHTML = description ? description : App.translations.NoProvidedInfo;
				}
			}
		});
		this.attach(this.mTemplateSelect);

		this.mTemplateDescription = document.createElement("p");
		this.container.appendChild(this.mTemplateDescription);

		const colorLbl = document.createElement("h3");
		colorLbl.className = "title";
		colorLbl.innerHTML = App.translations.Color;
		this.container.appendChild(colorLbl);

		const colorSelect = new Ui.DivSelect();
		colorSelect.addOption("", "#8227ac").style.backgroundColor = "#8227ac";
		colorSelect.addOption("", "#0d81b4").style.backgroundColor = "#0d81b4";
		colorSelect.addOption("", "#75b98e").style.backgroundColor = "#75b98e";
		colorSelect.addOption("", "#68e060").style.backgroundColor = "#68e060";
		colorSelect.addOption("", "#e06868").style.backgroundColor = "#e06868";
		colorSelect.addOption("", "#ac7827").style.backgroundColor = "#ac7827";
		this.attach(colorSelect);
		colorSelect.select("#0d81b4");

		this.mCreateButton = document.createElement("button");
		this.mCreateButton.innerHTML = App.translations.CreateChecklist;
		this.mCreateButton.style.width = "auto";
		this.mCreateButton.style.cssFloat = "right";
		this.mCreateButton.style.borderRadius = "8px";
		this.mCreateButton.style.backgroundColor = "var(--success-color)";
		this.mCreateButton.disabled = true;
		this.container.appendChild(this.mCreateButton);
		this.mCreateButton.addEventListener("click", function() {
			const r = new Requests.Lists();
			r.name = new Models.Translation();
			r.name.en = self.mName.value;
			r.name.sv = self.mName.value;
			r.name.de = self.mName.value;
			r.template = self.mTemplateSelect.value!;
			r.visibility = Models.Visibility.Hidden;
			r.color = colorSelect.value!;
			const req = new HttpRequest("POST", "/v1/lists", JSON.stringify(r));
			req.process(function(status: any, data: any) {
				location.reload();
			});
		});
	}

	public get id(): string {
		return "checklists";
	}

	public onBecomeVisible(): void {
		super.onBecomeVisible();
		this.refresh();
	}

	private updateCreateButtonStatus(): void {
		if (this.mTemplateSelect.value === null) {
			this.mCreateButton.disabled = true;
			return;
		}
		if(this.mName.value === "") {
			this.mCreateButton.disabled = true;
			return;
		}
		this.mCreateButton.disabled = false;
	}

	private refresh(): void {
		this.mProfile.setImage("/img/portraits/" + App.user.portrait);
		const self = this;
		const r = new HttpRequest("GET", "/v1/lists/list");
		r.process(function(status: any, data: any) {
			self.mChecklists.clear();
			data.result.forEach(e => {
				const c = document.createElement("p");
				c.className = "entry";
				let color = "#0d81b4";
				if (e.color !== "") {
					color = e.color;
				}
				c.innerHTML = "<span style='font-size: 1rem; color: " + color + ";' class='material-symbols-outlined'>trip_origin</span> " + App.getTranslation(e.name);
				self.mChecklists.container.appendChild(c);
				c.addEventListener("click", function() {
					new Ui.MessageBox("Edit List", "Here you will edit list.", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--success-color)", function() {});
				});
			});
		});

		this.mTemplateDescription.innerHTML = "";
		this.mTemplateSelect.clear();
		const s = new HttpRequest("GET", "/v1/list_templates/list");
		s.process(function(status: any, data: any) {
			data.result.forEach(element => {
				const o = self.mTemplateSelect.addOption(App.getTranslation(element.name), element.uuid);
				o.setAttribute("description", App.getTranslation(element.description));
			});
		});
	}
}