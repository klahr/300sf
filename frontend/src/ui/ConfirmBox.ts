/// <reference path="MessageBox.ts"/>

namespace Ui {
	export class ConfirmBox extends MessageBox {
		constructor(title: string, message: string, image: string | null, color: string, yesCallback: () => void, closeCallback: () => void, yesText: string, cancelText: string) {
			super(title, message, image, color, closeCallback);

			const actions = document.createElement("div");
			actions.className = "dialog-actions";
			this.popUp().container.appendChild(actions);

			const cancelButton = document.createElement("button");
			cancelButton.className = "secondary";
			cancelButton.innerHTML = cancelText;
			actions.appendChild(cancelButton);

			const self = this;
			const yesButton = document.createElement("button");
			yesButton.innerHTML = yesText;
			yesButton.addEventListener("click", function (event) {
				event.stopPropagation();
				self.close();
				yesCallback();
			});
			actions.appendChild(yesButton);
		}
	}
}
