/// <reference path="Widget.ts"/>

namespace Ui {
    export class Select extends Widget {
        private mDropDown: HTMLDivElement;
        private mSearch: Ui.Input;

        constructor() {
            super();

            this.mContainer = document.createElement("div");
            this.mContainer.className = "uix-div";
            this.mContainer.style.position = "relative";
            this.mContainer.style.backgroundColor = "var(--widget-color)";
            this.mContainer.style.borderRadius = "var(--small-distance)";

            const self = this;
            this.mSearch = new Ui.Input("text");
            this.mSearch.style.margin = "0";
            this.mSearch.setPlaceHolder("Search...");
            this.mSearch.style.borderBottomLeftRadius = "0";
            this.mSearch.style.borderBottomRightRadius = "0";

            (this.mSearch.container as HTMLInputElement).addEventListener("keyup", function() {
                const filter = this.value.toLowerCase();
                const options = self.mDropDown.getElementsByTagName("p");
                for (let i = 0; i < options.length; i++) {
                    if (options[i].innerHTML.toLowerCase().indexOf(filter) > -1) {
                        options[i].style.display = "";
                    } else {
                        options[i].style.display = "none";
                    }
                }
            });
            this.attach(this.mSearch);

            this.mDropDown = document.createElement("div");
            this.mDropDown.style.margin = "0 24px var(--small-distance) 24px";
            this.mDropDown.style.overflowY = "auto";
            this.mDropDown.style.height = "100px";
            this.mContainer.appendChild(this.mDropDown);
        }

        public get value(): string | null {
            const options = this.mDropDown.getElementsByTagName("p");
            for (let i = 0; i < options.length; ++i) {
                if (options[i].innerHTML === this.mSearch.value) {
                    return options[i].getAttribute("value");
                }
            }
            return null;
        }

        public addOption(name: string, value: string): void {
            const self = this;
            const option = document.createElement("p");
            option.className = "uix-option";
            option.innerHTML = name;
            option.setAttribute("value", value);
            option.addEventListener("click", function() {
                self.mSearch.value = this.innerHTML;
            });
            this.mDropDown.appendChild(option);
        }
    }
}
