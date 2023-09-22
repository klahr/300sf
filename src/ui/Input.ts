/// <reference path="Widget.ts"/>

namespace Ui {
	export class Input extends Widget {
		constructor(type: string) {
			super();

			this.mContainer = document.createElement("input");
			(this.mContainer as HTMLInputElement).type = type;
			this.mContainer.className = "uix-input";
		}

		public setPlaceHolder(str: string): void {
			this.mContainer.setAttribute("placeholder", str);
		}

		public set value(v: string) {
			(this.mContainer as HTMLInputElement).value = v;
		}

		public get value(): string {
			return (this.mContainer as HTMLInputElement).value;
		}

		public focus(): void {
			(this.mContainer as HTMLInputElement).focus();
		}

		public select(): void {
			(this.mContainer as HTMLInputElement).select();
		}

		public setEmpty(): void {
			(this.mContainer as HTMLInputElement).value = "";
		}
	}
}
