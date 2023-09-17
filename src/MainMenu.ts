/// <reference path="ui/Div.ts"/>

class MainMenu extends Ui.Widget {
	private mContent = new Ui.Div();

	constructor() {
		super();

		this.mContainer = document.createElement("div");
		this.mContainer.className = "uix-div main-menu";
		this.mContainer.style.width = "100%";
		this.mContainer.style.height = "100%";
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

		this.addSection("Checklists");
		this.addEntry("300 Svenska Fåglar", function() {
		});
		this.addEntry("Tävling 2023", function() {
		});

		this.addSection("Account");
		this.addEntry("Profile", function() {
			console.log("Profile");
		});

		this.addEntry("Log Out", function() {
			console.log("Log out");
		});
	}

	private addSection(name: string): void {
		const header = document.createElement("p");
		header.innerHTML = name.toUpperCase();
		header.style.fontSize = "0.7rem";
		header.style.fontWeight = "bold";
		header.style.padding = "16px";
		header.style.opacity = "0.5";
		this.mContent.container.appendChild(header);
	}

	private addEntry(name: string, callback: Function): void {
		const entry = document.createElement("p");
		entry.innerHTML = name;
		entry.style.padding = "0 0 8px 16px";
		entry.style.color = "#2782ac";
		entry.addEventListener("click", function() {
			callback();
		});
		this.mContent.container.appendChild(entry);
	}
}
