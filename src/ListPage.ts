class ListPage extends Page {
	private mContent: Ui.Div;

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
		this.attach(profile);
	}

	public get id(): string {
		return "list-page";
	}

	public setData(data: any): void {
		super.setData(data);

		const self = this;

		self.mContent.container.innerHTML = "";

		let d = data as Models.List;
		d.species.forEach(e => {
			const name = document.createElement("p");
			name.innerHTML = App.getTranslation(e.name);
			self.mContent.container.appendChild(name);
		});
	}
}
