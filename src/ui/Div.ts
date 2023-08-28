/// <reference path="Widget.ts"/>

namespace Ui {
	export class Div extends Widget {
		constructor() {
			super();

			const self = this;

			this.mContainer = document.createElement("div");
			this.mContainer.className = "uix-div";

			this.mContainer.addEventListener("click", function(event: any) {
				self.onClick(event);
			});
		}
	}
}
