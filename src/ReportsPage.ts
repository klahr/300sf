class ReportsPage extends Page {
	private mTitle: HTMLHeadingElement;

	constructor(appListener: IAppListener) {
		super(appListener);

		this.mTitle = document.createElement("h1");
		this.mTitle.innerHTML = "Reports";
		this.container.appendChild(this.mTitle);

		this.addProfile();
	}

	public get id(): string {
		return "reports";
	}
}