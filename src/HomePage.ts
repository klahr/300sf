/// <reference path="ui/Widget.ts"/>
/// <reference path="ui/Image.ts"/>
/// <reference path="ui/Input.ts"/>
/// <reference path="Page.ts"/>

class HomePage extends Page {
	constructor(appListener: IAppListener) {
		super(appListener);
	}

	protected onCreateLoggedIn(): void {
        const self = this;

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
	
	protected onCreateLoggedOut(): void {
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
		title.innerHTML = "Sign in";
		this.container.appendChild(title);

		const email = new Ui.Input("email");
		email.style.position = "relative";
		email.setPlaceHolder("Email");
		this.attach(email);

		const password = new Ui.Input("password");
		password.style.position = "relative";
		password.setPlaceHolder("Password");
		this.attach(password);

		const staySignedIn = new Ui.Input("checkbox");
		staySignedIn.style.position = "relative";
		staySignedIn.container.setAttribute("name", "stay_signed_in");
		this.attach(staySignedIn);

		const staySignedInLbl = document.createElement("label");
		staySignedInLbl.setAttribute("for", staySignedIn.container.getAttribute("name")!);
		staySignedInLbl.innerHTML = "Stay signed in";
		this.container.appendChild(staySignedInLbl);

		const help = document.createElement("p");
		help.className = "details";
		help.innerHTML = "Please mark the checkbox if you prefer not to sign in during each visit.";
		this.container.appendChild(help);

		const signIn = document.createElement("button");
		signIn.innerHTML = "Sign in";
		this.container.appendChild(signIn);

		const hr = document.createElement("hr");
		this.container.appendChild(hr);

		const noAccountYet = document.createElement("p");
		noAccountYet.innerHTML = "Don't have an account yet? <a href=\"signup\">Sign up here</a>.";
		this.container.appendChild(noAccountYet);

		const forgot = document.createElement("a");
		forgot.href = "/forgot";
		forgot.innerHTML = "Forgot you password?";
		this.container.appendChild(forgot);
	}
}
