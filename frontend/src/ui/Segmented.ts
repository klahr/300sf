namespace Ui {
	// A row of buttons of which one is chosen; changed is called when the user picks another.
	export function segmented(options: Array<[string, string, string]>, selected: string, changed: (value: string) => void): HTMLDivElement {
		const group = document.createElement("div");
		group.className = "segmented";
		group.setAttribute("role", "radiogroup");
		const select = function (value: string, notify: boolean) {
			Array.from(group.children).forEach(b => {
				const on = (b as HTMLElement).dataset.value === value;
				b.classList.toggle("selected", on);
				b.setAttribute("aria-checked", String(on));
			});
			if (notify) {
				changed(value);
			}
		};
		options.forEach(o => {
			const b = document.createElement("button");
			b.type = "button";
			b.setAttribute("role", "radio");
			b.dataset.value = o[0];
			b.innerHTML = "<span class='material-symbols-outlined'>" + o[1] + "</span>" + Utils.escapeHtml(o[2]);
			b.addEventListener("click", function () {
				select(o[0], true);
			});
			group.appendChild(b);
		});
		select(selected, false);
		return group;
	}
}
