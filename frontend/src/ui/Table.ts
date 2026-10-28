/// <reference path="Widget.ts"/>

namespace Ui {
	export class Table extends Widget {
		constructor() {
			super();

			this.mContainer = document.createElement("table");
			this.mContainer.className = "uix-table";
		}

		public clear(): void {
			this.mContainer.innerHTML = "";
		}

		// Grey rows shaped like content, shown until the real rows replace them.
		public showPlaceholders(count: number): void {
			this.clear();
			for (let i = 0; i < count; ++i) {
				const p = document.createElement("div");
				p.className = "placeholder-row";
				p.innerHTML = "<span class='skeleton'></span><span class='skeleton'></span>";
				this.addRow([p]).classList.add("placeholder");
			}
		}

		public addRow(elements: Array<any>): HTMLTableRowElement {			const row = document.createElement("tr");

			elements.forEach(element => {
				const column = document.createElement("td");
				column.appendChild(element);
				row.appendChild(column);
			});

			this.mContainer.appendChild(row);
			return row;
		}
	}
}
