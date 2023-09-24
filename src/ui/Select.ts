/// <reference path="Widget.ts"/>
/// <reference path="Input.ts"/>

namespace Ui {
	export class Select extends Widget {
		private mDropDown: HTMLDivElement;
		private mResult: HTMLParagraphElement;
		private mSearch: Ui.Input;
		private mSelection: HTMLParagraphElement | null = null;

		constructor() {
			super();

			const self = this;

			this.mContainer = document.createElement("div");
			this.mContainer.className = "uix-select";

			this.mResult = document.createElement("p");
			this.mResult.style.padding = "var(--medium-distance)";
			this.mResult.style.cursor = "pointer";
			this.mResult.innerHTML = "Tufted duck";
			this.mResult.addEventListener("click", function() {
				self.toggleVisible();
			})
			this.mContainer.appendChild(this.mResult);

			this.mSearch = new Ui.Input("text");
			this.mSearch.style.margin = "0";
			this.mSearch.setPlaceHolder("Search...");
			this.mSearch.style.borderBottomLeftRadius = "0";
			this.mSearch.style.borderBottomRightRadius = "0";
			this.mSearch.style.display = "none";

			this.mSearch.container.addEventListener("keydown", function(e) {
				const allOptions = self.mDropDown.getElementsByTagName("p");
				if (e.key == "ArrowDown") {
					let i;
					for (i = 0; i < allOptions.length; ++i) {
						const value = allOptions[i].getAttribute("value")!;
						if (value === "") {
							continue;
						}
						if (allOptions[i].style.visibility !== "none") {
							if (self.mSelection === null) {
								self.mSelection = allOptions[i];
								allOptions[i].classList.add("selected");
							} else {
								allOptions[i].classList.remove("selected");
								if (self.mSelection === allOptions[i]) {
									if (i == allOptions.length - 1) {
										console.log("last");
									} else {
										for (; i < allOptions.length; ++i) {
											if (allOptions[i].style.visibility !== "none") {
												self.mSelection = allOptions[i];
												allOptions[i].classList.add("selected");
												break;
											}
										}
									}
								}
							}
						}
					}
					console.log("down");
				} else if (e.key == "ArrowUp") {
					console.log("up");
				} else if (e.key == "Enter") {
					for (let i = 0; i < allOptions.length; ++i) {
						if (self.mSelection === allOptions[i]) {
							self.mResult.innerHTML = allOptions[i].innerHTML;
							self.mResult.setAttribute("value", allOptions[i].getAttribute("value")!);
							break;
						}
					}
				} else {
					self.mSelection = null;
				}
			});

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
			this.mDropDown.style.display = "none";
			this.mDropDown.style.width = "100%";
			this.mContainer.appendChild(this.mDropDown);
	}

		public toggleVisible(): void {
			if (this.mSearch.style.display === "inline-block") {
				this.mSearch.style.display = "none";
				this.mDropDown.style.display = "none";
			} else {
				this.mSearch.style.display = "inline-block";
				this.mDropDown.style.display = "inline-block";
				this.mSearch.select();
			}
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
				self.mResult.innerHTML = this.innerHTML;
				self.mResult.setAttribute("value", this.getAttribute("value")!);
			});
			this.mDropDown.appendChild(option);
		}
	}
}
