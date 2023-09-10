/// <reference path="ui/Div.ts"/>

class MainMenu extends Ui.Widget {
    constructor() {
        super();

        this.mContainer = document.createElement("div");
        this.mContainer.className = "uix-div main-menu";
        this.mContainer.style.width = "100%";
        this.mContainer.style.height = "100%";
        this.mContainer.addEventListener("click", function() {
            this.classList.remove("open");
        });

        const content = new Ui.Div();
        content.addClass("content");
        this.attach(content);
    }
}
