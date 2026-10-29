/// <reference path="Page.ts"/>
/// <reference path="ui/Input.ts"/>
/// <reference path="App.ts"/>
/// <reference path="EditLabelPopUp.ts"/>
/// <reference path="requests/Tags.ts"/>
/// <reference path="Colors.ts"/>

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
		this.mTags.addClass("tag-cloud");
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
		colorSelect.addClass("swatches");
		Colors.palette.forEach(c => colorSelect.addOption("", c).style.backgroundColor = c);
		this.attach(colorSelect);
		colorSelect.select(Colors.fallback);

		this.mCreateButton = document.createElement("button");
		this.mCreateButton.innerHTML = App.translations.CreateTag;
		this.mCreateButton.className = "inline success end";
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

		this.groupSections();
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
			(data.result || []).forEach(e => {
				const t = e as Models.Tag;
				const tag = document.createElement("div");
				tag.className = "tag";
				if (t.color !== "") {
					tag.style.setProperty("--tag", t.color);
				}
				const c = document.createElement("p");
				c.textContent = t.name;
				tag.appendChild(c);
				tag.setAttribute("role", "button");
				tag.tabIndex = 0;
				self.mTags.container.appendChild(tag);
				tag.addEventListener("keydown", function (event) {
					if (event.key === "Enter") {
						openEditTagPopUp(e);
					}
				});
				tag.addEventListener("click", function () {
					openEditTagPopUp(e);
				});
			});
		});
	}
}