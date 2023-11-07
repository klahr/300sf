/// <reference path="Page.ts"/>
/// <reference path="ui/Input.ts"/>
/// <reference path="App.ts"/>
/// <reference path="EditLabelPopUp.ts"/>
/// <reference path="requests/Tags.ts"/>

class ManageTagsPage extends Page {
	private mTitle: HTMLHeadingElement;
	private mTags: Ui.Div;
	private mName: Ui.Input;
	private mCreateButton: HTMLButtonElement;

	constructor(appListener: IAppListener) {
		super(appListener);
		const self = this;

		this.addProfile();

		this.mTitle = document.createElement("h1");
		this.mTitle.innerHTML = App.translations.Tags;
		this.container.appendChild(this.mTitle);

		this.mTags = new Ui.Div();
		this.mTags.style.position = "relative";
		this.mTags.style.margin = "var(--small-distance)";
		this.attach(this.mTags);

		const newTagLbl = document.createElement("h2");
		newTagLbl.innerHTML = App.translations.NewLabel;
		this.container.appendChild(newTagLbl);

		this.mName = new Ui.Input("text");
		this.mName.setPlaceHolder(App.translations.Name);
		this.mName.container.addEventListener("input", function () {
			self.updateCreateButtonStatus();
		});
		this.attach(this.mName);

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
		this.mCreateButton.innerHTML = App.translations.CreateTag;
		this.mCreateButton.style.width = "auto";
		this.mCreateButton.style.cssFloat = "right";
		this.mCreateButton.style.borderRadius = "8px";
		this.mCreateButton.style.backgroundColor = "var(--success-color)";
		this.mCreateButton.disabled = true;
		this.container.appendChild(this.mCreateButton);
		this.mCreateButton.addEventListener("click", function () {
			const r = new Requests.Tags();
			r.name = self.mName.value;
			r.color = colorSelect.value!;
			const req = new HttpRequest("POST", "/v1/tags", r);
			req.process(function (status: any, data: any) {
				location.reload();
			});
		});
	}

	public get id(): string {
		return "tags";
	}

	public requireSignIn(): boolean {
		return true;
	}

	public onBecomeVisible(): void {
		super.onBecomeVisible();
		this.refresh();
	}

	private updateCreateButtonStatus(): void {
		if (this.mName.value === "") {
			this.mCreateButton.disabled = true;
			return;
		}
		this.mCreateButton.disabled = false;
	}

	private refresh(): void {
		this.mProfile.setImage("/img/portraits/" + App.user.portrait);
		const self = this;
		const r = new HttpRequest("GET", "/v1/tags/list");
		r.process(function (status: any, data: any) {
			self.mTags.clear();
			data.result.forEach(e => {
				const t = e as Models.Tag;
				const tag = document.createElement("div");
				tag.className = "tag";
				if (t.color !== "") {
					tag.style.backgroundColor = t.color;
				}
				const c = document.createElement("p");
				c.innerHTML = t.name;
				tag.appendChild(c);
				self.mTags.container.appendChild(tag);
				c.addEventListener("click", function () {
					openEditTagPopUp(e);
				});
			});
		});
	}
}