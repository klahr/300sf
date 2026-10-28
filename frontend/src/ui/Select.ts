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
		private mOutsideListener = (event: PointerEvent) => {
			if (!this.mContainer.contains(event.target as Node)) {
				this.hide();
			}
		};
		private mFocusOutListener = (event: FocusEvent) => {
			if (event.relatedTarget && !this.mContainer.contains(event.relatedTarget as Node)) {
				this.hide();
			}
		};

		constructor(placeholder: string, searchText: string, selectCallback: (value: string | null) => void = function () { }) {
			super();

			this.mSelectCallback = selectCallback;

			const self = this;

			this.mContainer = document.createElement("div");
			this.mContainer.className = "uix-select";

			this.mResult = document.createElement("p");
			this.mResult.className = "uix-select-value";
			this.mResult.setAttribute("role", "button");
			this.mResult.tabIndex = 0;
			this.mResult.addEventListener("keydown", function (e) {
				if (e.key === "Enter" || e.key === " ") {
					e.preventDefault();
					self.toggleVisible();
				}
			});
			this.mResult.innerHTML = placeholder;
			this.mResult.addEventListener("click", function () {
				self.toggleVisible();
			});
			this.mContainer.appendChild(this.mResult);

			this.mOpenArrow = document.createElement("span");
			this.mOpenArrow.className = "material-symbols-outlined uix-select-arrow";
			this.mOpenArrow.innerHTML = "expand_more";
			this.mContainer.appendChild(this.mOpenArrow);


			this.mExpandable = new Ui.Div();
			this.mExpandable.addClass("uix-select-panel");
			this.mExpandable.setVisible(false);
			this.attach(this.mExpandable);

			this.mSearch = new Ui.Input("text");
			this.mSearch.setPlaceHolder(searchText + "...");

			this.mSearch.container.addEventListener("keydown", function (e) {
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
									allOptions[i + 1].classList.add("selected");
									selected = allOptions[i + 1];
									++i;
								} else {
									selected = allOptions[i];
								}
							} else {
								if (i > 0) {
									if (allOptions[i - 1].getAttribute("value") !== "") {
										allOptions[i].classList.remove("selected");
										allOptions[i - 1].classList.add("selected");
										selected = allOptions[i - 1];
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
					}
				} else if (e.key === "Escape") {
					e.stopPropagation();
					self.hide();
					self.mResult.focus();
				} else if (e.key == "Enter") {
					for (let i = 0; i < allOptions.length; ++i) {
						if (allOptions[i].classList.contains("selected")) {
							self.mResult.textContent = allOptions[i].textContent;
							self.mResult.setAttribute("value", allOptions[i].getAttribute("value")!);
							self.toggleVisible();
							self.mSelectCallback(allOptions[i].getAttribute("value"));
							break;
						}
					}
				}
			});

			(this.mSearch.container as HTMLInputElement).addEventListener("keyup", function (event) {
				if (event.key === "ArrowUp" || event.key === "ArrowDown") {
					return;
				}
				const filter = this.value.toLowerCase();
				const options = self.mDropDown.getElementsByTagName("p");
				for (let i = 0; i < options.length; i++) {
					if (options[i].textContent!.toLowerCase().indexOf(filter) > -1) {
						options[i].style.display = "";
						if (options[i].textContent!.toLowerCase() === filter) {
							self.updateScroll(options[i]);
							options[i].classList.add("selected");
						} else {
							options[i].classList.remove("selected");
						}
					} else {
						options[i].style.display = "none";
					}
				}
			});
			this.mExpandable.attach(this.mSearch);

			this.mDropDown = document.createElement("div");
			this.mDropDown.className = "uix-select-options";
			this.mExpandable.container.appendChild(this.mDropDown);
		}

		public setSelected(value: string): void {
			const allOptions = this.mDropDown.getElementsByTagName("p");
			for (var i = 0; i < allOptions.length; ++i) {
				const v = allOptions[i].getAttribute("value")!;
				if (v === value) {
					allOptions[i].click();
					this.hide();
					break;
				}
			}
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
			this.addClass("open");
			this.mSearch.select();
			const self = this;
			// Close when the user taps or tabs anywhere else.
			setTimeout(function () {
				document.addEventListener("pointerdown", self.mOutsideListener);
				self.mContainer.addEventListener("focusout", self.mFocusOutListener);
			}, 0);
		}

		public hide(): void {
			this.mExpandable.setVisible(false);
			this.removeClass("open");
			document.removeEventListener("pointerdown", this.mOutsideListener);
			this.mContainer.removeEventListener("focusout", this.mFocusOutListener);
		}

		public get value(): string | null {
			return this.mResult.getAttribute("value");
		}

		public addOption(name: string, value: string): HTMLParagraphElement {
			const self = this;
			const option = document.createElement("p");
			option.className = "uix-option";
			option.textContent = name;
			option.setAttribute("value", value);
			option.addEventListener("click", function () {
				self.mResult.textContent = this.textContent;
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

		public getOption(value: string): HTMLParagraphElement | null {
			const options = this.mDropDown.getElementsByClassName("uix-option");
			for (let i = 0; i < options.length; ++i) {
				if (options[i].getAttribute("value") === value) {
					return options[i] as HTMLParagraphElement;
				}
			}
			return null;
		}

		public setDisabled(disabled: boolean): void {
			if (disabled) {
				this.addClass("disabled");
				this.mResult.tabIndex = -1;
			} else {
				this.removeClass("disabled");
				this.mResult.tabIndex = 0;
			}
		}

		public clear(): void {
			this.mDropDown.innerHTML = "";
		}
	}
}
