/// <reference path="utils/Date.ts"/>
/// <reference path="Page.ts"/>
/// <reference path="ui/Input.ts"/>

class ContributePage extends Page {
	private mContent: Ui.Div;

	constructor(appListener: IAppListener) {
		super(appListener);

		this.mContent = new Ui.Div();
		this.mContent.style.position = "relative";
		this.mContent.style.width = "100%";
		this.attach(this.mContent);

		this.addProfile();

		const title = document.createElement("p");
		title.style.textAlign = "center";
		title.innerHTML = "sourcehut";
		this.container.appendChild(title);
	}

	public get id(): string {
		return "contribute";
	}

	public onBecomeVisible(): void {
		super.onBecomeVisible();
		this.refresh();
	}

	private refresh(): void {
		this.mProfile.setImage("/img/portraits/" + App.user.portrait);
	}
}