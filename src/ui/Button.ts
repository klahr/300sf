/// <reference path="Widget.ts"/>

namespace Ui {
	export class Button extends Widget {
		constructor(text: string, icon: string) {
			super();

			this.mContainer = document.createElement("div");
			this.mContainer.className = "uix-button";

			if (icon) {
				const img = new Ui.Image(icon, "6vw", "6vw");
				img.style.left = "4vw";
				img.style.top = "3vw";
				this.attach(img);
			}

			const txt = document.createElement("p");
			txt.innerHTML = text;
			this.mContainer.appendChild(txt);

			if (icon) {
				txt.style.marginLeft = "8vw";
			}
		}
	}
}
