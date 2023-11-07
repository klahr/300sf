/// <reference path="Widget.ts"/>

namespace Ui {
	export class DivSelect extends Widget {
		private mSelectCallback: (value: string | null) => void | null;

		constructor(selectCallback: (value: string | null) => void = function () { }) {
			super();

			this.mSelectCallback = selectCallback;

			this.mContainer = document.createElement("div");
			this.mContainer.style.position = "relative";
			this.mContainer.className = "uix-div";
		}

		public select(value: string): void {
			const allOptionsSource = this.container.getElementsByClassName("uix-option");
			for (let i = 0; i < allOptionsSource.length; ++i) {
				if (allOptionsSource[i].getAttribute("value") === value) {
					allOptionsSource[i].classList.add("selected");
				} else {
					allOptionsSource[i].classList.remove("selected");
				}
			}
		}

		public get value(): string | null {
			const allOptionsSource = this.container.getElementsByClassName("uix-option");
			for (let i = 0; i < allOptionsSource.length; ++i) {
				if (allOptionsSource[i].classList.contains("selected")) {
					return allOptionsSource[i].getAttribute("value");
				}
			}
			return null;
		}

		public addOption(className: string, value: string): HTMLDivElement {
			const self = this;
			const option = document.createElement("div");
			option.className = "uix-option";
			option.setAttribute("value", value);
			option.addEventListener("click", function () {
				const allOptionsSource = self.container.getElementsByClassName("uix-option");
				for (let i = 0; i < allOptionsSource.length; ++i) {
					allOptionsSource[i].classList.remove("selected");
				}
				option.classList.add("selected");
				self.mSelectCallback(option.getAttribute("value"));
			});
			this.container.appendChild(option);
			return option;
		}
	}
}