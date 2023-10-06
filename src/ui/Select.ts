/// <reference path="Widget.ts"/>
/// <reference path="Div.ts"/>
/// <reference path="Input.ts"/>

namespace Ui {
	export class Select extends Widget {
		private mDropDown: HTMLDivElement;
		private mResult: HTMLParagraphElement;
		private mSearch: Ui.Input;
		private mExpandable: Ui.Div;
		private mOpenArrow: HTMLSpanElement;
		private mSelectCallback: (value: string | null) => void | null;

		constructor(placeholder: string, searchText: string, selectCallback: (value: string | null) => void = function() {}) {
			super();

			this.mSelectCallback = selectCallback;

			const self = this;

			this.mContainer = document.createElement("div");
			this.mContainer.className = "uix-select";

			this.mResult = document.createElement("p");
			this.mResult.style.padding = "12px 24px";
			this.mResult.style.cursor = "pointer";
			this.mResult.style.color = "var(--black-or-white-color)";
			this.mResult.innerHTML = placeholder;
			this.mResult.addEventListener("click", function() {
				self.toggleVisible();
			})
			this.mContainer.appendChild(this.mResult);

			this.mOpenArrow = document.createElement("span");
			this.mOpenArrow.className = "material-symbols-outlined";
			this.mOpenArrow.style.position = "absolute";
			this.mOpenArrow.style.right = "16px";
			this.mOpenArrow.style.top = "8px";
			this.mOpenArrow.style.fontSize = "32px";
			this.mOpenArrow.style.color = "var(--black-or-white-color)";
			this.mOpenArrow.innerHTML = "expand_more";
			this.mOpenArrow.style.pointerEvents = "none";
			this.mContainer.appendChild(this.mOpenArrow);

			this.mExpandable = new Ui.Div();
			this.mExpandable.style.backgroundColor = "var(--widget-color)";
			this.mExpandable.style.position = "absolute";
			this.mExpandable.style.left = "0";
			this.mExpandable.style.top = "48px";
			this.mExpandable.style.width = "100%";
			this.mExpandable.style.height = "256px";
			this.mExpandable.style.borderBottomLeftRadius = "var(--small-distance)";
			this.mExpandable.style.borderBottomRightRadius = "var(--small-distance)";
			this.mExpandable.style.zIndex = "1";
			this.mExpandable.setVisible(false);
			this.attach(this.mExpandable);

			this.mSearch = new Ui.Input("text");
			this.mSearch.style.margin = "0";
			this.mSearch.setPlaceHolder(searchText + "...");
			this.mSearch.style.borderBottomLeftRadius = "0";
			this.mSearch.style.borderBottomRightRadius = "0";

			this.mSearch.container.addEventListener("keydown", function(e) {
				const allOptionsSource = self.mDropDown.getElementsByTagName("p");
				const allOptions = Array<HTMLParagraphElement>();
				for (let i = 0; i < allOptionsSource.length; ++i) {
					if (allOptionsSource[i].style.display !== "none") {
						allOptions.push(allOptionsSource[i]);
					}
				}
				if (e.key === "ArrowDown" || e.key === "ArrowUp") {
					e.preventDefault();
					let i;
					let selected: HTMLParagraphElement | null = null;
					for (i = 0; i < allOptions.length; ++i) {
						const value = allOptions[i].getAttribute("value")!;
						if (value === "") {
							continue;
						}
						if (allOptions[i].classList.contains("selected")) {
							if (e.key === "ArrowDown") {
								if (i < allOptions.length - 1) {
									allOptions[i].classList.remove("selected");
									allOptions[i+1].classList.add("selected");
									selected = allOptions[i+1];
									++i;
								} else {
									selected = allOptions[i];
								}
							} else {
								if (i > 0) {
									if (allOptions[i-1].getAttribute("value") !== "") {
										allOptions[i].classList.remove("selected");
										allOptions[i-1].classList.add("selected");
										selected = allOptions[i-1];
									}
								}
							}
						}
					}
					if (!selected) {
						for (let i = 0; i < allOptions.length; ++i) {
							if (allOptions[i].getAttribute("value") !== "") {
								allOptions[i].classList.add("selected");
								break;
							}
						}
					}
					if (selected) {
						self.updateScroll(selected);
						self.mSelectCallback(selected.getAttribute("value"));
					}
				} else if (e.key == "Enter") {
					for (let i = 0; i < allOptions.length; ++i) {
						if (allOptions[i].classList.contains("selected")) {
							self.mResult.innerHTML = allOptions[i].innerHTML;
							self.mResult.setAttribute("value", allOptions[i].getAttribute("value")!);
							self.toggleVisible();
							break;
						}
					}
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
			this.mExpandable.attach(this.mSearch);

			this.mDropDown = document.createElement("div");
			this.mDropDown.style.margin = "0 24px var(--small-distance) 24px";
			this.mDropDown.style.overflowY = "auto";
			this.mDropDown.style.height = "calc(100% - 56px)";
			this.mDropDown.style.width = "calc(100% - 48px)";
			this.mExpandable.container.appendChild(this.mDropDown);
		}

		public toggleVisible(): void {
			if (this.mExpandable.isVisible()) {
				this.hide();
			} else {
				this.show();
			}
 		}

		public show(): void {
			this.mExpandable.setVisible(true);
			this.style.borderBottomLeftRadius = "0";
			this.style.borderBottomRightRadius = "0";
			this.mOpenArrow.innerHTML = "expand_less";
			this.mSearch.select();
		}

		public hide(): void {
			this.mExpandable.setVisible(false);
			this.style.borderBottomLeftRadius = "";
			this.style.borderBottomRightRadius = "";
			this.mOpenArrow.innerHTML = "expand_more";
		}

		public get value(): string | null {
			return this.mResult.getAttribute("value");
		}

		public addOption(name: string, value: string): HTMLParagraphElement {
			const self = this;
			const option = document.createElement("p");
			option.className = "uix-option";
			option.innerHTML = name;
			option.setAttribute("value", value);
			option.addEventListener("click", function() {
				self.mResult.innerHTML = this.innerHTML;
				self.mResult.setAttribute("value", this.getAttribute("value")!);
				const allOptionsSource = self.mDropDown.getElementsByTagName("p");
				for (let i = 0; i < allOptionsSource.length; ++i) {
					allOptionsSource[i].classList.remove("selected");
				}
				option.classList.add("selected");
				self.updateScroll(this);
				self.toggleVisible();
				self.mSelectCallback(option.getAttribute("value"));
			});
			this.mDropDown.appendChild(option);
			return option;
		}

		private updateScroll(selectedP: HTMLParagraphElement): void {
			const parentDivRect: DOMRect = this.mDropDown.getBoundingClientRect();
			const selectedPRect = selectedP.getBoundingClientRect();
			if (selectedPRect.bottom > parentDivRect.bottom) {
				this.mDropDown.scrollTop += selectedPRect.bottom - parentDivRect.bottom;
			} else if (selectedPRect.top < parentDivRect.top) {
				this.mDropDown.scrollTop -= parentDivRect.top - selectedPRect.top;
			}
		}

		public getOption(value: string): HTMLParagraphElement  | null {
			const options = this.mDropDown.getElementsByClassName("uix-option");
			for (let i = 0; i < options.length; ++i) {
				if (options[i].getAttribute("value") === value) {
					return options[i] as HTMLParagraphElement;
				}
			}
			return null;
		}

		public clear(): void {
			console.log("clear");
			this.mDropDown.innerHTML = "";
		}
	}
}
