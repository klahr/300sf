/// <reference path="Div.ts"/>
/// <reference path="Button.ts"/>

namespace Ui {
	export class MessageBox extends Div {
		private mRoot: HTMLDivElement;
		private mPopup: Ui.Div;
		private mCloseCallback: () => void;

		constructor(title: string, message: string, image: string | null, color: string, closeCallback: () => void) {
			super();

			var id = this.hashStr(title + message);
			if (document.getElementById(id) !== null) {
				return;
			}

			const self = this;

			this.mCloseCallback = closeCallback;

			this.addClass("uix-popup-base");
			this.container.id = id;

			this.mRoot = document.getElementById("root") as HTMLDivElement;
			this.mRoot.appendChild(this.mContainer);

			setTimeout(function () {
				self.addClass("uix-popup-background");
			}, 50);

			this.mPopup = new Ui.Div();
			this.mPopup.addClass("uix-messagebox");
			this.mPopup.addClass("animate__animated");
			this.mPopup.addClass("animate__backInDown");
			this.mPopup.style.setProperty("--animate-duration", "750ms");
			this.mPopup.style.padding = "8px";
			this.attach(this.mPopup);

			const c = new Ui.Div();
			c.container.style.padding = "16px";
			c.container.style.boxSizing = "border-box";
			c.container.style.width = "100%";
			c.container.style.height = "100%";
			c.container.style.overflow = "auto";
			c.container.style.color = "white";
			c.container.style.position = "relative";
			c.container.style.textAlign = "left";
			c.container.style.border = "2px solid " + color;
			c.container.style.borderRadius = "16px";
			c.container.style.textAlign = "center";
			this.mPopup.attach(c);

			if (image) {
				const img = document.createElement("img");
				img.src = image;
				img.style.width = "75%";
				img.style.maxWidth = "128px";
				img.style.borderRadius = "50%";
				c.container.appendChild(img);
			}

			const t = document.createElement("h2");
			t.innerHTML = title;
			c.container.appendChild(t);

			const m = document.createElement("p");
			m.innerHTML = message;
			c.container.appendChild(m);

			setTimeout(function () {
				self.mContainer.addEventListener("click", function () {
					self.close();
				});
			}, 500);
		}

		public close(): void {
			const self = this;
			this.mPopup.removeClass("animate__zoomIn");
			this.mPopup.addClass("animate__backOutUp");
			this.mPopup.style.setProperty('--animate-duration', '1s');

			setTimeout(function () {
				self.removeClass("uix-popup-background");
			}, 250);

			setTimeout(function () {
				self.mRoot.removeChild(self.mContainer);
			}, 500);

			if (this.mCloseCallback) {
				this.mCloseCallback();
			}
		}

		protected popUp(): Div {
			return this.mPopup;
		}

		protected hashStr(str: string): string {
			let hash = 0;
			for (let i = 0; i < str.length; i++) {
				const char = str.charCodeAt(i);
				hash = (hash << 5) - hash + char;
			}
			return Math.abs(hash).toString();
		}
	}
}
