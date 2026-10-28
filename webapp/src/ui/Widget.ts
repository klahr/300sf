namespace Ui {
	export class Widget {
		protected mChildren: Array<Widget> = Array();
		protected mContainer: HTMLElement;
		protected mData: any;

		protected constructor() { }

		public onDestroy(): void {
			this.mChildren.forEach(e => {
				e.onDestroy();
			});
			this.mChildren = Array();
		}

		public attach(child: Widget): void {
			this.mChildren.push(child);
			this.mContainer.appendChild(child.mContainer);
			child.onAttachedToParent(this);
		}

		public detach(child: Widget): void {
			const index = this.mChildren.indexOf(child);
			if (index > -1) {
				this.mChildren[index].onDetachFromParent();
				this.mContainer.removeChild(child.mContainer);
				this.mChildren.splice(index, 1);
			}
		}

		public detachAll(): void {
			const self = this;
			this.mChildren.forEach(child => {
				child.onDetachFromParent();
				self.mContainer.removeChild(child.mContainer);
			});
			this.mChildren.splice(0, this.mChildren.length);
		}

		protected onDetachFromParent(): void { }

		protected onAttachedToParent(_parent: Widget): void { }

		public get style(): CSSStyleDeclaration {
			return this.container.style;
		}

		public addClass(name: string): void {
			this.container.classList.add(name);
		}

		public removeClass(name: string): void {
			this.container.classList.remove(name);
		}

		public setClass(name: string): void {
			this.container.className = name;
		}

		public hasClass(name: string): boolean {
			return this.container.classList.contains(name);
		}

		public get container(): HTMLElement {
			return this.mContainer;
		}

		public get children(): Array<Widget> {
			return this.mChildren;
		}

		public click(): void {
			this.mContainer.click();
		}

		public clear(): void {
			this.mChildren = [];
			this.mContainer.innerHTML = "";
		}

		public setData(data: any): void {
			this.mData = data;
		}

		public get data(): any {
			return this.mData;
		}

		protected onClick(_event: any): void { }

		public onEnterForground(): void {
			this.mChildren.forEach(child => {
				child.onEnterForground();
			});
		}

		public onEnterBackground(): void {
			this.mChildren.forEach(child => {
				child.onEnterBackground();
			});
		}

		public setVisible(visible: boolean): void {
			if (visible) {
				this.removeClass("hidden");
				this.addClass("visible");
			} else {
				this.removeClass("visible");
				this.addClass("hidden");
			}
		}

		public isVisible(): boolean {
			return !this.hasClass("hidden");
		}
	}
}
