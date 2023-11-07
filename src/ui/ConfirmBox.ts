/// <reference path="MessageBox.ts"/>

namespace Ui {
	export class ConfirmBox extends MessageBox {
		constructor(title: string, message: string, image: string | null, color: string, yesCallback: () => void, closeCallback: () => void, yesText: string, cancelText: string) {
			super(title, message, image, color, closeCallback);

			const cancelButton = document.createElement("button");
			cancelButton.innerHTML = cancelText;
			cancelButton.style.width = "calc(50% - var(--small-distance)*2)";
			cancelButton.style.backgroundColor = "transparent";
			cancelButton.style.boxShadow = "0 0 0 1px inset white";
			this.popUp().container.appendChild(cancelButton);

			const self = this;
			const yesButton = document.createElement("button");
			yesButton.innerHTML = yesText;
			yesButton.style.width = "calc(50% - var(--small-distance)*2)";
			yesButton.addEventListener("click", function (event) {
				event.stopPropagation();
				self.container.click();
				yesCallback();
			});
			this.popUp().container.appendChild(yesButton);
		}
	}
}