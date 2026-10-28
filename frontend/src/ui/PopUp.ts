/// <reference path="Div.ts"/>
/// <reference path="Button.ts"/>

namespace Ui {
	export class PopUp extends Div {
		// Open dialogs, newest last. Each has a history entry so the back button closes it.
		private static sOpen: Array<PopUp> = [];
		// Set when a dialog closed by itself removes its history entry, so that pop is not handled.
		private static sIgnoreNextPop = 0;
		// Waiting for those pops, so they do not land after a navigation made in the meantime.
		private static sSettled: Array<() => void> = [];
		private mRoot: HTMLDivElement;
		private mPopup: Ui.Div;
		private mCloseCallback: () => void;
		private mClosed = false;
		private mKeyListener: (event: KeyboardEvent) => void;

		constructor(content: HTMLElement, closeCallback: () => void) {
			super();

			this.mCloseCallback = closeCallback;

			this.addClass("uix-popup-base");

			this.mRoot = document.getElementById("root") as HTMLDivElement;
			this.mRoot.appendChild(this.mContainer);

			const self = this;

			this.mPopup = new Ui.Div();
			this.mPopup.addClass("uix-popup");
			this.mPopup.container.setAttribute("role", "dialog");
			this.mPopup.container.setAttribute("aria-modal", "true");
			this.attach(this.mPopup);

			const grabber = new Ui.Div();
			grabber.addClass("uix-popup-grabber");
			this.mPopup.attach(grabber);

			const c = new Ui.Div();
			c.addClass("uix-popup-body");
			this.mPopup.attach(c);

			const close = document.createElement("button");
			close.className = "uix-close";
			close.setAttribute("aria-label", App.translations.Close);
			close.innerHTML = "<span class='material-symbols-outlined'>close</span>";
			close.addEventListener("click", function () {
				self.close();
			});
			c.container.appendChild(close);

			c.container.appendChild(content);

			this.mKeyListener = function (event: KeyboardEvent) {
				if (event.key === "Escape" && self.isTopmost()) {
					self.close();
				}
			};
			document.addEventListener("keydown", this.mKeyListener);

			// A click on the dimmed area around the dialog closes it. Both the press and the release
			// must be outside: dragging the map or selecting text can end outside, and the browser
			// reports that as a click on the backdrop too.
			let pressedOutside = false;
			this.mContainer.addEventListener("pointerdown", function (event) {
				pressedOutside = event.target === self.mContainer;
			});
			this.mContainer.addEventListener("click", function (event) {
				if (pressedOutside && event.target === self.mContainer && self.isTopmost()) {
					self.close();
				}
				pressedOutside = false;
			});

			PopUp.sOpen.push(this);
			history.pushState({ popup: PopUp.sOpen.length }, "", window.location.href);

			requestAnimationFrame(function () {
				requestAnimationFrame(function () {
					self.addClass("uix-popup-background");
					self.mPopup.addClass("open");
				});
			});
		}

		private isTopmost(): boolean {
			const popups = this.mRoot.getElementsByClassName("uix-popup-base");
			return popups[popups.length - 1] === this.mContainer;
		}

		// Runs the callback once a dialog that just closed has left history, so a page change made
		// in it is not undone by that dialog's own step back.
		public static whenHistorySettled(callback: () => void): void {
			if (PopUp.sIgnoreNextPop > 0) {
				PopUp.sSettled.push(callback);
			} else {
				callback();
			}
		}

		// Handles the back button:closes the newest dialog, or returns false when none is open. A pop
		// caused by a dialog closing itself is swallowed here too.
		public static handleBack(): boolean {
			if (PopUp.sIgnoreNextPop > 0) {
				--PopUp.sIgnoreNextPop;
				if (PopUp.sIgnoreNextPop === 0) {
					const settled = PopUp.sSettled;
					PopUp.sSettled = [];
					settled.forEach(cb => cb());
				}
				return true;
			}
			const top = PopUp.sOpen[PopUp.sOpen.length - 1];
			if (!top) {
				return false;
			}
			top.close(true);
			return true;
		}

		public close(fromHistory: boolean = false): void {
			if (this.mClosed) {
				return;
			}
			this.mClosed = true;

			const index = PopUp.sOpen.indexOf(this);
			if (index >= 0) {
				PopUp.sOpen.splice(index, 1);
			}
			// Closed with its own button: take its history entry away again.
			if (!fromHistory && history.state && history.state.popup) {
				++PopUp.sIgnoreNextPop;
				history.back();
			}
			document.removeEventListener("keydown", this.mKeyListener);

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
	}
}
