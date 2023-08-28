namespace Ui {
	export class PopUp extends Div {
		private mRoot: HTMLDivElement;
		private mPopup: Ui.Div;

		constructor(icon: string, text: string, okCallback: any, cancelCallback: any) {
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
			this.attach(this.mPopup);

			if (icon) {
				const i = document.createElement("img");
				i.src = icon;
				i.style.width = "30%";
				i.style.margin = "5vw";
				this.mPopup.container.appendChild(i);
			}

			if (text) {
				const t = document.createElement("p");
				t.innerHTML = text;
				t.style.fontSize = "calc(12px + 2vw)";
				t.style.padding = "0 4vw 4vw 4vw";
				this.mPopup.container.appendChild(t);
			}

			if (okCallback) {
				const okButton = new Ui.Button("OK", null);
				okButton.style.position = "relative";
				okButton.style.width = "20vw";
				okButton.style.marginBottom = "4vw";
				okButton.style.display = "inline-block";
				this.mPopup.attach(okButton);
				okButton.container.addEventListener("click", function() {
					self.close();
					okCallback();
				});
			}

			if (cancelCallback) {
				const cancelButton = new Ui.Button("Avbryt", null);
				cancelButton.style.position = "relative";
				cancelButton.style.width = "20vw";
				cancelButton.style.marginLeft = "4vw";
				cancelButton.style.marginBottom = "4vw";
				cancelButton.style.backgroundColor = "var(--second-color)";
				cancelButton.style.display = "inline-block";
				this.mPopup.attach(cancelButton);
				cancelButton.container.addEventListener("click", function() {
					self.close();
					cancelCallback();
				});
			}
		}

		public close(): void {
			const self = this;
			this.mPopup.removeClass("animate__bounceIn");
			this.mPopup.addClass("animate__bounceOut");
			this.mPopup.style.setProperty('--animate-duration', '0.5s');

			setTimeout(function() {
				self.removeClass("uix-popup-background");
			}, 250);

			setTimeout(function() {
				self.mRoot.removeChild(self.mContainer);
			}, 500);
		}
	}
}
