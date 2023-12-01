/// <reference path="Widget.ts"/>

namespace Ui {
	export class MultiInput extends Widget {
		private mSelect: Select;
		private mType: string;
	
		constructor(type: string) {
			super();
			this.mType = type;

			this.mContainer = document.createElement("div");
			this.mContainer.className = "uix-div";
			this.mContainer.style.width = "100%";
			this.mContainer.style.position = "relative";

			const self = this;
			this.mSelect = new Select("", "", (value: string | null) => {
				self.setSelected(value!);
			});

			this.attach(this.mSelect);
		}

		public setSelected(value: string): void {
			this.mSelect.setSelected(value);
			this.children.forEach(e => {
				if (e instanceof Input) {
					const i = e as Input;
					i.setVisible(i.data === value);
				}
			});
		}

		public addOption(name: string, optionName: string, value: string): void {
			const i = new Input(this.mType);
			i.setData(value);
			i.setPlaceHolder(optionName);
			i.setVisible(false);
			this.attach(i);
			this.mSelect.addOption(name, value);
		}
	}
}