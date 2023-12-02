/// <reference path="Page.ts"/>
/// <reference path="ui/Image.ts"/>
/// <reference path="ui/DivSelect.ts"/>
/// <reference path="ProfilePopUp.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="EditListPopUp.ts"/>
/// <reference path="requests/Lists.ts"/>
/// <reference path="models/Visibility.ts"/>
/// <reference path="models/ListTemplate.ts"/>

class ManageCheckListsPage extends Page {
	private mTitle: HTMLHeadingElement;
	private mTemplatesTitle: HTMLHeadingElement;
	private mChecklists: Ui.Div;
	private mTemplates: Ui.Div;
	private mTemplateSelect: Ui.Select;
	private mTemplateDescription: HTMLParagraphElement;
	private mCreateButton: HTMLButtonElement;
	private mCreateTemplateButton: HTMLButtonElement;
	private mName: Ui.Input;
	private mTemplateName: Ui.Input;

	constructor(appListener: IAppListener) {
		super(appListener);

		this.addProfile();

		this.mTitle = document.createElement("h1");
		this.mTitle.innerHTML = App.translations.Checklists;
		this.container.appendChild(this.mTitle);

		this.mChecklists = new Ui.Div();
		this.mChecklists.style.position = "relative";
		this.attach(this.mChecklists);

		const newChecklistsLbl = document.createElement("h2");
		newChecklistsLbl.innerHTML = App.translations.NewChecklist;
		this.container.appendChild(newChecklistsLbl);

		this.mName = new Ui.Input("text");
		this.mName.setPlaceHolder(App.translations.Name);
		this.mName.container.addEventListener("input", function () {
			self.updateCreateButtonStatus();
		});
		this.attach(this.mName);

		const self = this;
		this.mTemplateSelect = new Ui.Select(App.translations.SelectTemplate + "...", App.translations.Search, function (value: string | null) {
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
		colorSelect.style.width = "100%";
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
		this.mCreateButton.addEventListener("click", function () {
			const r = new Requests.Lists();
			r.name = new Models.Translation();
			r.name.en = self.mName.value;
			r.name.sv = self.mName.value;
			r.name.de = self.mName.value;
			r.template = self.mTemplateSelect.value!;
			r.visibility = Models.Visibility.Hidden;
			r.color = colorSelect.value!;
			const req = new HttpRequest("POST", "/v1/lists", r);
			req.process(function (status: any, data: any) {
				location.reload();
			});
		});

		this.mTemplatesTitle = document.createElement("h1");
		this.mTemplatesTitle.innerHTML = App.translations.Templates;
		this.container.appendChild(this.mTemplatesTitle);

		this.mTemplates = new Ui.Div();
		this.mTemplates.style.position = "relative";
		this.attach(this.mTemplates);

		const newTemplateLbl = document.createElement("h2");
		newTemplateLbl.innerHTML = App.translations.NewTemplate;
		this.container.appendChild(newTemplateLbl);

		this.mTemplateName = new Ui.Input("text");
		this.mTemplateName.setPlaceHolder(App.translations.Name);
		this.mTemplateName.container.addEventListener("input", function () {
			self.updateCreateTemplateButtonStatus();
		});
		this.attach(this.mTemplateName);

		this.mCreateTemplateButton = document.createElement("button");
		this.mCreateTemplateButton.innerHTML = App.translations.CreateTemplate;
		this.mCreateTemplateButton.style.width = "auto";
		this.mCreateTemplateButton.style.cssFloat = "right";
		this.mCreateTemplateButton.style.borderRadius = "8px";
		this.mCreateTemplateButton.style.backgroundColor = "var(--success-color)";
		this.mCreateTemplateButton.disabled = true;
		this.container.appendChild(this.mCreateTemplateButton);
		this.mCreateTemplateButton.addEventListener("click", function () {
			/*const r = new Requests.Lists();
			r.name = new Models.Translation();
			r.name.en = self.mName.value;
			r.name.sv = self.mName.value;
			r.name.de = self.mName.value;
			r.template = self.mTemplateSelect.value!;
			r.visibility = Models.Visibility.Hidden;
			r.color = colorSelect.value!;
			const req = new HttpRequest("POST", "/v1/lists", r);
			req.process(function (status: any, data: any) {
				location.reload();
			});*/
			new Ui.MessageBox(App.translations.Error, "Not yet implemented.", "/img/messagebox-256.webp", "var(--error-color)", function () { });
		});
	}

	public get id(): string {
		return "checklists";
	}

	public requireSignIn(): boolean {
		return true;
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
		if (this.mName.value === "") {
			this.mCreateButton.disabled = true;
			return;
		}
		this.mCreateButton.disabled = false;
	}

	private updateCreateTemplateButtonStatus(): void {
		if (this.mTemplateName.value === "") {
			this.mCreateTemplateButton.disabled = true;
			return;
		}
		this.mCreateTemplateButton.disabled = false;
	}

	private refresh(): void {
		this.mProfile.setImage("/img/portraits/" + App.user.portrait);
		const self = this;
		const r = new HttpRequest("GET", "/v1/lists/list");
		r.process(function (status: any, data: any) {
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
				c.addEventListener("click", function () {
					openEditListPopUp(e);
				});
			});
		});

		this.mTemplateDescription.innerHTML = "";
		this.mTemplateSelect.clear();
		const s = new HttpRequest("GET", "/v1/list_templates/list");
		s.process(function (status: any, data: any) {
			data.result.forEach(element => {
				const o = self.mTemplateSelect.addOption(App.getTranslation(element.name), element.uuid);
				o.setAttribute("description", App.getTranslation(element.description));
			});
		});

		this.mTemplates.clear();
		const t = new HttpRequest("GET", "/v1/list_templates/list/me");
		t.process(function (status: any, data: any) {
			data.result.forEach(element => {
				const e = element as Models.ListTemplate;
				const c = document.createElement("p");
				c.className = "entry";
				var icon: string;
				switch (e.visibility) {
					case Models.Visibility.Private:
						icon = "lock";
						break;
					case Models.Visibility.Hidden:
						icon = "do_not_disturb_on_total_silence";
						break;
					default:
						icon = "public";
				}
				c.innerHTML = "<span style='font-size: 1rem; color: #0d81b4;' class='material-symbols-outlined'>public</span> " + App.getTranslation(e.name);
				self.mTemplates.container.appendChild(c);
				c.addEventListener("click", function () {
					new Ui.MessageBox(App.translations.Error, "Not yet implemented.", "/img/messagebox-256.webp", "var(--error-color)", function () { });
				});
			});
		});
	}
}