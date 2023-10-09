/// <reference path="Widget.ts"/>

namespace Ui {
	export class Image extends Widget {
		constructor(src: string, width: string, height: string) {
			super();

			const self = this;

			this.mContainer = document.createElement("div");
			this.mContainer.className = "uix-img";
			this.mContainer.style.backgroundImage = "url(" + src + ")";
			this.mContainer.style.width = width;
			this.mContainer.style.height = height;

			this.mContainer.addEventListener("click", function(event: any) {
				self.onClick(event);
			});
		}

		public setImage(src: string): void {
			this.mContainer.style.backgroundImage = "url(" + src + ")";
		}
	}
}
