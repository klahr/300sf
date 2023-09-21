/// <reference path="Div.ts"/>
/// <reference path="Button.ts"/>

namespace Ui {
	export class PopUp extends Div {
		private mRoot: HTMLDivElement;
		private mPopup: Ui.Div;

		constructor(content: HTMLElement, closeCallback: () => void) {
			super();

			this.addClass("uix-popup-base");

			this.mRoot = document.getElementById("root") as HTMLDivElement;
			this.mRoot.appendChild(this.mContainer);

			const self = this;

			setTimeout(function() {
				self.addClass("uix-popup-background");
			}, 50);

			this.mPopup = new Ui.Div();
			this.mPopup.addClass("uix-popup");
			this.mPopup.addClass("animate__animated");
			this.mPopup.addClass("animate__bounceIn");
			this.mPopup.style.setProperty("--animate-duration", "1000ms");
			this.attach(this.mPopup);

			const c = new Ui.Div();
			c.container.style.padding = "16px";
			c.container.style.boxSizing = "border-box";
			c.container.style.width = "100%";
			c.container.style.height = "100%";
			c.container.style.overflow = "auto";
			c.container.style.color = "white";
			c.container.addEventListener("click", function() {
				self.close();
			});
			this.mPopup.attach(c);

			const close =  document.createElement("span");
			close.className = "material-symbols-outlined close";
			close.style.cssFloat = "right";
			close.style.cursor = "pointer";
			close.style.fontSize = "48px";
			close.innerHTML = "close";
			c.container.appendChild(close);

			c.container.appendChild(content);
		}

		public close(): void {
			const self = this;
			this.mPopup.removeClass("animate__bounceIn");
			this.mPopup.addClass("animate__bounceOut");
			this.mPopup.style.setProperty('--animate-duration', '1000ms');

			setTimeout(function() {
				self.removeClass("uix-popup-background");
			}, 250);

			setTimeout(function() {
				self.mRoot.removeChild(self.mContainer);
			}, 500);
		}
	}
}
