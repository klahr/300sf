/// <reference path="Layer.ts"/>

namespace Ui {
	export class LayerManager {
		private mLayers: Array<Layer> = Array();
		private mRoot: HTMLDivElement;

		constructor() {
			this.mRoot = document.getElementById("root") as HTMLDivElement;
		}

		public addLayer(id: string): boolean {
			if (this.getIndexOfLayer(id) === -1) {
				const layer = new Layer(id, this.mLayers.length.toString());
				this.mRoot.appendChild(layer.container);
				this.mLayers.push(layer);
				return true;
			} else {
				return false;
			}
		}

		public clearLayer(id: string): void {
			const layer = this.getLayer(id);
			if (layer) {
				layer.removeAll();
			}
		}

		public addToLayer(id: string, widget: Ui.Widget): boolean {
			const layer = this.getLayer(id);
			if (layer) {
				layer.add(widget);
				return true;
			}
			return false;
		}

		public removeFromLayer(id: string, widget: Ui.Widget): boolean {
			const layer = this.getLayer(id);
			if (layer) {
				layer.remove(widget);
				return true;
			}
			return false;
		}

		public showLayer(id: string): void {
			let found = false;
			for (let i = 0; i < this.mLayers.length; ++i) {
				if (this.mLayers[i].id === id) {
					this.mLayers[i].show();
					found = true;
				} else {
					if (found) {
						this.mLayers[i].hideSide();
					} else {
						this.mLayers[i].hideScale();
					}
				}
			}
		}

		private getIndexOfLayer(id: string): number {
			for (let i = 0; i < this.mLayers.length; ++i) {
				if (this.mLayers[i].id === id) {
					return i;
				}
			}
			return -1;
		}

		private getLayer(id: string): Layer | null {
			for (let i = 0; i < this.mLayers.length; ++i) {
				if (this.mLayers[i].id === id) {
					return this.mLayers[i];
				}
			}
			return null;
		}
	}
}
