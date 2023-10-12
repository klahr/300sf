class ManageTagsPage extends Page {
	private mContent: Ui.Div;
	private mTitle: HTMLHeadingElement;

	constructor(appListener: IAppListener) {
		super(appListener);

		this.mContent = new Ui.Div();
		this.mContent.style.position = "relative";
		this.mContent.style.width = "100%";
		this.attach(this.mContent);

		this.addProfile();

		this.mTitle = document.createElement("h1");
		this.mTitle.innerHTML = App.translations.ManageTags;
		this.container.appendChild(this.mTitle);
	}

	public get id(): string {
		return "tags";
	}

	public onBecomeVisible(): void {
		super.onBecomeVisible();
		this.refresh();
	}

	private refresh(): void {
		this.mProfile.setImage("/img/portraits/" + App.user.portrait);
	}
}