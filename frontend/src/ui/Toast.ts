/// <reference path="Div.ts"/>
/// <reference path="Button.ts"/>

namespace Ui {
	export class Toast extends Div {
		private mPopup: Ui.Div;
		private mCloseCallback: () => void;
		private mClosed = false;

		// With a duration of 0 it stays until tapped.
		constructor(title: string, message: string, image: string | null, color: string, closeCallback: () => void, durationMs: number = 3500) {
			super();

			const id = this.hashStr(title + message);
			if (document.getElementById(id) !== null) {
				return;
			}

			const self = this;

			this.mCloseCallback = closeCallback;

			// One host for all toasts so a newer one replaces the older.
			document.querySelectorAll("div.uix-toast-host").forEach(e => e.remove());
			this.addClass("uix-toast-host");
			this.container.id = id;
			this.container.setAttribute("role", "status");
			document.body.appendChild(this.mContainer);

			this.mPopup = new Ui.Div();
			this.mPopup.addClass("uix-toast");
			this.mPopup.style.setProperty("--tone", color);
			this.mPopup.container.addEventListener("click", function () {
				self.close();
			});
			this.attach(this.mPopup);

			if (image) {
				const img = document.createElement("img");
				img.src = image;
				img.alt = "";
				img.width = 40;
				img.height = 40;
				img.style.borderRadius = "50%";
				this.mPopup.container.appendChild(img);
			} else {
				const icon = document.createElement("span");
				icon.className = "material-symbols-outlined";
				icon.textContent = "check_circle";
				this.mPopup.container.appendChild(icon);
			}

			const text = document.createElement("div");
			this.mPopup.container.appendChild(text);

			const t = document.createElement("h2");
			t.textContent = title;
			text.appendChild(t);

			const m = document.createElement("p");
			m.innerHTML = message;
			text.appendChild(m);

			requestAnimationFrame(function () {
				requestAnimationFrame(function () {
					self.mPopup.addClass("open");
				});
			});

			if (durationMs > 0) {
				setTimeout(function () {
					self.close();
				}, durationMs);
			}
		}

		public close(): void {
			if (this.mClosed) {
				return;
			}
			this.mClosed = true;

			const self = this;
			this.mPopup.removeClass("open");

			setTimeout(function () {
				self.mContainer.remove();
			}, 400);

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
