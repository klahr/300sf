class ForgotPage extends Page {
    constructor(appListener: IAppListener) {
        super(appListener);

        const content = new Ui.Div();
		content.style.position = "relative";
		content.style.width = "100%";
		content.style.textAlign = "center";
		this.attach(content);
		const logo = document.createElement("img");
		logo.src = "img/99249498-452f-4b0b-a1f5-45902f7ffe48.png";
		logo.alt = "TODO";
		logo.style.width = "192px";
		logo.style.borderRadius = "50%";
		content.container.appendChild(logo);

		const title = document.createElement("h1");
		title.innerHTML = App.translations.ForgotPassword;
		this.container.appendChild(title);

		const email = new Ui.Input("email");
		email.style.position = "relative";
		email.setPlaceHolder(App.translations.Email);
		this.attach(email);

		const reset = document.createElement("button");
		reset.innerHTML = App.translations.SendResetLink;
		reset.addEventListener("click", function() {
//			const r = new HttpRequest("POST", "/v1/users", JSON.stringify(content));
//			r.process(function(status: any, data: any) {
//				console.log(data);
//			});
		});

		this.container.appendChild(reset);
    }

	public get id(): string {
		return "forgot-page";
	}
}
