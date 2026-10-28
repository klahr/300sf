/// <reference path="Div.ts"/>
/// <reference path="Button.ts"/>

namespace Ui {
	export class MessageBox extends Div {
		private mRoot: HTMLDivElement;
		private mPopup: Ui.Div;
		private mCloseCallback: () => void;
		private mClosed = false;

		constructor(title: string, message: string, image: string | null, color: string, closeCallback: () => void) {
			super();

			const self = this;

			this.mCloseCallback = closeCallback;
			this.mRoot = document.getElementById("root") as HTMLDivElement;

			this.mPopup = new Ui.Div();
			this.mPopup.addClass("uix-messagebox");
			this.mPopup.container.setAttribute("role", "alertdialog");
			this.mPopup.style.setProperty("--tone", color);
			this.attach(this.mPopup);

			// The same message is shown once; a duplicate stays detached and is never displayed.
			const id = this.hashStr(title + message);
			if (document.getElementById(id) !== null) {
				this.mClosed = true;
				return;
			}

			this.addClass("uix-popup-base");
			this.container.id = id;
			this.mRoot.appendChild(this.mContainer);

			if (image) {
				const img = document.createElement("img");
				img.src = image;
				img.alt = "";
				this.mPopup.container.appendChild(img);
			}

			const t = document.createElement("h2");
			t.textContent = title;
			this.mPopup.container.appendChild(t);

			const m = document.createElement("p");
			m.innerHTML = message;
			this.mPopup.container.appendChild(m);

			requestAnimationFrame(function () {
				requestAnimationFrame(function () {
					self.addClass("uix-popup-background");
					self.mPopup.addClass("open");
				});
			});

			setTimeout(function () {
				self.mContainer.addEventListener("click", function () {
					self.close();
				});
			}, 300);
		}

		public close(): void {
			if (this.mClosed) {
				return;
			}
			this.mClosed = true;

			const self = this;
			this.mPopup.removeClass("open");
			this.mPopup.addClass("closing");
			this.removeClass("uix-popup-background");

			setTimeout(function () {
				self.mContainer.remove();
			}, 260);

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
