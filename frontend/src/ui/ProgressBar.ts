/// <reference path="Div.ts"/>

namespace Ui {
	export class ProgressBar extends Div {
		private mProgression: Div;

		constructor() {
			super();

			this.addClass("uix-progress-bar");

			this.mProgression = new Div();
			this.mProgression.addClass("uix-progress-fill");

			this.attach(this.mProgression);
		}

		public setProgress(percent: number): void {
			this.mProgression.style.width = percent + "%";
		}

		public addProgressClass(name: string): void {
			this.mProgression.container.classList.add(name);
		}

		public removeProgressClass(name: string): void {
			this.mProgression.container.classList.remove(name);
		}
	}
}
