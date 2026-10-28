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

		public addRow(elements: Array<any>): void {
			const row = document.createElement("tr");

			elements.forEach(element => {
				const column = document.createElement("td");
				column.appendChild(element);
				row.appendChild(column);
			});

			this.mContainer.appendChild(row);
		}
	}
}
