namespace Ui {
	// The bar along the top edge while requests are running. It shows only for requests slower than a
	// moment, creeps towards the end while waiting, and fills and fades when the last one finishes.
	export namespace Progress {
		let running = 0;
		let value = 0;
		let showTimer: number | null = null;
		let trickleTimer: number | null = null;

		function bar(): HTMLElement | null {
			return document.getElementById("loading");
		}

		function set(v: number): void {
			value = v;
			const b = bar();
			if (b) {
				b.style.transform = "scaleX(" + v + ")";
			}
		}

		export function start(): void {
			++running;
			if (running > 1 || showTimer !== null) {
				return;
			}
			showTimer = window.setTimeout(function () {
				showTimer = null;
				const b = bar();
				if (!b || running === 0) {
					return;
				}
				b.classList.remove("done");
				b.classList.add("active");
				set(0.08);
				trickleTimer = window.setInterval(function () {
					// Smaller steps the further it gets; it never reaches the end on its own.
					set(value + (0.92 - value) * 0.08);
				}, 200);
			}, 150);
		}

		export function done(): void {
			running = Math.max(0, running - 1);
			if (running > 0) {
				return;
			}
			if (showTimer !== null) {
				clearTimeout(showTimer);
				showTimer = null;
			}
			if (trickleTimer !== null) {
				clearInterval(trickleTimer);
				trickleTimer = null;
			}
			const b = bar();
			if (!b || !b.classList.contains("active")) {
				return;
			}
			set(1);
			b.classList.add("done");
			window.setTimeout(function () {
				if (running === 0) {
					b.classList.remove("active", "done");
					set(0);
				}
			}, 400);
		}
	}

	// Disables a button and shows a spinner on it while something is in progress.
	export function setBusy(button: HTMLButtonElement, busy: boolean): void {
		button.classList.toggle("busy", busy);
		button.disabled = busy;
		if (busy) {
			button.setAttribute("aria-busy", "true");
		} else {
			button.removeAttribute("aria-busy");
		}
	}

	export function prefersReducedMotion(): boolean {
		return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
	}

	// Counts a number up from zero.
	export function countUp(element: HTMLElement, target: number, duration: number = 800): void {
		const format = (n: number) => n.toLocaleString(App.language);
		if (target <= 0 || prefersReducedMotion()) {
			element.textContent = format(target);
			return;
		}
		const start = performance.now();
		const step = function (now: number) {
			const t = Math.min(1, (now - start) / duration);
			const eased = 1 - Math.pow(1 - t, 3);
			element.textContent = format(Math.round(target * eased));
			if (t < 1) {
				requestAnimationFrame(step);
			}
		};
		element.textContent = format(0);
		requestAnimationFrame(step);
	}
}
