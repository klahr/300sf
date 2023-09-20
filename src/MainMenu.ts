/// <reference path="ui/Div.ts"/>

class MainMenu extends Ui.Widget {
	private mContent = new Ui.Div();

	constructor() {
		super();

		this.mContainer = document.createElement("div");
		this.mContainer.className = "uix-div main-menu";
		this.mContainer.addEventListener("click", function() {
			this.classList.remove("open");
		});

		this.mContent.addClass("content");
		this.mContent.container.addEventListener("click", function(event) {
			event.stopPropagation();
		});
		this.attach(this.mContent);

		const logo = document.createElement("img");
		logo.src = "img/300sf.svg";
		logo.style.width = "100%";
		logo.style.marginTop = "29px";
		logo.setAttribute("height", "36");
		this.mContent.container.appendChild(logo);

		if (App.isLoggedIn()) {
			this.createLoggedIn();
		} else {
			this.createLoggedOut();
		}
	}

	private createLoggedIn(): void  {
		this.addSection("Checklists");
		this.addEntry("300 Svenska Fåglar", function() {
		});
		this.addEntry("Tävling 2023", function() {
		});
		this.addEntry("+ New List", function() {
		}).style.color = "#2782ac";

		this.addSection("Account");
		this.addEntry("Profile", function() {
			console.log("Profile");
		});

		this.addEntry("Log Out", function() {
			setCookie("x300sf_at", "", 0);
			location.reload();
		});
	}

	private createLoggedOut(): void {
		this.addSection("A Poem");
		const poem = document.createElement("p");
		poem.style.padding = "0 16px";
		poem.style.fontStyle = "italic";
		poem.innerHTML = App.translations.Poem;
		this.mContent.container.appendChild(poem);
	}

	private addSection(name: string): HTMLParagraphElement {
		const header = document.createElement("p");
		header.className = "section";
		header.innerHTML = name.toUpperCase();
		this.mContent.container.appendChild(header);
		return header;
	}

	private addEntry(name: string, callback: Function): HTMLParagraphElement {
		const entry = document.createElement("p");
		entry.className = "entry";
		entry.innerHTML = name;
		entry.addEventListener("click", function() {
			callback();
		});
		this.mContent.container.appendChild(entry);
		return entry;
	}
}
