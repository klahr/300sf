/// <reference path="Widget.ts"/>
/// <reference path="Div.ts"/>

namespace Ui {
	export class Layer {
		private mView = new Div();

		constructor(id: string, zIndex: string) {
			this.mView.container.id = id;
			this.mView.addClass("uix-layer");
			this.mView.addClass("hide-scale");
			this.mView.style.zIndex = zIndex;
		}

		public get id(): string {
			return this.mView.container.id;
		}

		public get container(): HTMLElement {
			return this.mView.container;
		}

		public add(widget: Widget): void {
			this.mView.attach(widget);
		}

		public remove(widget: Widget): void {
			this.mView.detach(widget);
		}

		public removeAll(): void {
			this.mView.detachAll();
		}

		public show(): void {
			this.mView.removeClass("hide-scale");
			this.mView.removeClass("hide-side");
			this.onEnterForground();
		}

		public hideScale(): void {
			this.mView.addClass("hide-scale");
			this.onEnterBackground();
		}

		public hideSide(): void {
			this.mView.addClass("hide-side");
			this.onEnterBackground();
		}

		private onEnterForground(): void {
			this.mView.onEnterForground();
		}

		private onEnterBackground(): void {
			this.mView.onEnterBackground();
		}
	}
}
