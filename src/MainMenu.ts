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

		const header = document.createElement("p");
		header.innerHTML = "300sf";
		this.mContent.container.appendChild(header);

		this.addEntry("Log out", function() {
			console.log("Log out");
		});
	}

	private addEntry(name: string, callback: Function): void {
		const entry = document.createElement("p");
		entry.innerHTML = name;
		entry.addEventListener("click", function() {
			callback();
		});
		this.mContent.container.appendChild(entry);
	}
}
