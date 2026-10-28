/// <reference path="Widget.ts"/>

namespace Ui {
	export class MultiTextArea extends Widget {
		private mSelect: Select;
	
		constructor() {
			super();

			this.mContainer = document.createElement("div");
			this.mContainer.className = "uix-div";
			this.mContainer.style.width = "100%";
			this.mContainer.style.position = "relative";

			const self = this;
			this.mSelect = new Select("", "", (value: string | null) => {
				self.setSelected(value!);
			});

			this.attach(this.mSelect);
		}

		public setSelected(value: string): void {
			this.mSelect.setSelected(value);
			const children = this.mContainer.children;
			for (let i = 0; i < children.length; ++i) {
				const e = children[i];
				if (e instanceof HTMLTextAreaElement) {
					const i = e as HTMLTextAreaElement;
					i.style.display = i.getAttribute("data") === value ? "inline-block" : "none";
				}
			};
		}

		public addOption(name: string, optionName: string, value: string): void {
			const i = document.createElement("textarea");
			i.setAttribute("data", value);
			i.placeholder = optionName;
			i.style.display = "none";
			i.style.height = "128px";
			this.mContainer.appendChild(i);
			this.mSelect.addOption(name, value);
		}

		public getValue(value: string): string {
			const children = this.mContainer.children;
			for (let i = 0; i < children.length; ++i) {
				const e = children[i];
				if (e instanceof HTMLTextAreaElement) {
					const i = e as HTMLTextAreaElement;
					if (i.getAttribute("data") === value) {
						return i.value;
					}
				}
			};
			return "";
		}

		public addEventListener<K extends keyof HTMLElementEventMap>(type: K, listener: (this: HTMLElement, ev: HTMLElementEventMap[K]) => any, options?: boolean | AddEventListenerOptions): void {
			const children = this.mContainer.children;
			for (let i = 0; i < children.length; ++i) {
				const e = children[i];
				if (e instanceof HTMLTextAreaElement) {
					const i = e as HTMLTextAreaElement;
					i.addEventListener(type, listener);
				}
			};
		}
	}
}