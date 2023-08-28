namespace Ui {
	export class ProgressBar extends Div {
		private mProgression: Div;
		constructor() {
			super();

			this.addClass("uix-progress-bar");

			this.mProgression = new Div();
			this.mProgression.style.width = "0%";
			this.mProgression.style.height = "100%";
			this.mProgression.style.backgroundColor = "var(--second-color)";
			this.mProgression.style.borderRadius = "1vw";
			this.mProgression.style.transitionDuration = "1s";

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
