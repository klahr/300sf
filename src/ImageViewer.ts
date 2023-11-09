/// <reference path="ui/Div.ts"/>

class ImageViewer extends Ui.Div {
	private mRoot: HTMLDivElement;
	private mCloseCallback: () => void;
	private mImg: HTMLImageElement;

	constructor(image: string, closeCallback: () => void) {
		super();

		const self = this;

		this.mCloseCallback = closeCallback;

		this.addClass("uix-popup-base");

		this.mRoot = document.getElementById("root") as HTMLDivElement;
		this.mRoot.appendChild(this.mContainer);

		setTimeout(function () {
			self.addClass("uix-popup-background");
		}, 50);

		this.mImg = document.createElement("img");
		this.mImg.src = image;
		this.mImg.style.maxWidth = "100%";
		this.mImg.style.maxHeight = "100%";
		this.mImg.style.marginTop = "50vh";
		this.mImg.style.transform = "translateY(-50%)";
		this.mImg.style.transition = "opacity 250ms";
		this.mContainer.appendChild(this.mImg);

		const close = document.createElement("span");
		close.className = "material-symbols-outlined close";
		close.style.position = "absolute";
		close.style.cursor = "pointer";
		close.style.fontSize = "48px";
		close.style.right = "16px";
		close.style.top = "16px";
		close.style.color = "var(--secondary-color)";
		close.innerHTML = "close";
		close.addEventListener("click", function () {
			self.close();
		});
		this.mContainer.appendChild(close);
	}

	public close(): void {
		const self = this;

		setTimeout(function () {
			self.removeClass("uix-popup-background");
			this.mImg.addClass("fade");
		}, 250);

		setTimeout(function () {
			self.mRoot.removeChild(self.mContainer);
		}, 500);

		if (this.mCloseCallback) {
			this.mCloseCallback();
		}
	}

}