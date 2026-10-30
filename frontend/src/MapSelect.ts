/// <reference path="Maps.ts"/>
/// <reference path="ui/Widget.ts"/>
/// <reference path="models/RegionZone.ts"/>
/// <reference path="HttpRequest.ts"/>

// A map of the counties and regions in view, loaded as the map moves. Tapping one toggles it; the
// owner keeps what is selected and says so with setSelected.
class MapSelect extends Ui.Widget {
	private mMap: any = null;
	private mLayers = new Map<string, any>();
	private mSelected = new Set<string>();
	private mToggle: (zone: Models.RegionZone) => void;
	private mLoadTimer = 0;

	constructor(toggle: (zone: Models.RegionZone) => void) {
		super();
		this.mToggle = toggle;
		this.mContainer = document.createElement("div");
		this.mContainer.className = "map zone-map";
	}

	// Leaflet measures its container, so the map is made once it is on screen.
	public show(): void {
		if (this.mMap) {
			this.mMap.invalidateSize();
			return;
		}
		const self = this;
		let last: any = null;
		try {
			last = JSON.parse(localStorage.getItem("last-position") || "null")?.value;
		} catch (e) { }
		this.mMap = L.map(this.mContainer, { doubleClickZoom: false }).setView(last ? [last.lat, last.lng] : [62.0, 16.0], last ? 6 : 4);
		Maps.tileLayer().addTo(this.mMap);
		this.mMap.on("moveend", function () {
			clearTimeout(self.mLoadTimer);
			self.mLoadTimer = setTimeout(function () {
				self.loadInView();
			}, 250);
		});
		this.loadInView();
	}

	public remove(): void {
		clearTimeout(this.mLoadTimer);
		if (this.mMap) {
			this.mMap.remove();
			this.mMap = null;
		}
	}

	private loadInView(): void {
		if (!this.mMap) {
			return;
		}
		const self = this;
		const b = this.mMap.getBounds();
		const clamp = (v: number, lo: number, hi: number) => Math.max(lo, Math.min(hi, v)).toFixed(4);
		const bbox = [clamp(b.getWest(), -180, 180), clamp(b.getSouth(), -90, 90), clamp(b.getEast(), -180, 180), clamp(b.getNorth(), -90, 90)].join(",");
		const r = new HttpRequest("GET", "/v1/zones/regions?bbox=" + bbox);
		r.setQuiet();
		r.process(function (status: any, data: any) {
			(data.result || []).forEach((z: Models.RegionZone) => self.add(z));
		});
	}

	// Draws the region, unless it is drawn already.
	public add(zone: Models.RegionZone): void {
		if (!this.mMap || !zone.uuid || !zone.outline || this.mLayers.has(zone.uuid)) {
			return;
		}
		const self = this;
		const uuid = zone.uuid;
		const layer = L.geoJSON(zone.outline, {
			style: function () {
				return { className: "zone-area", weight: 1.5 };
			}
		});
		layer.bindTooltip(Utils.escapeHtml(App.getTranslation(zone.name)), { sticky: true });
		layer.on("click", function () {
			self.mToggle(zone);
		});
		layer.addTo(this.mMap);
		this.mLayers.set(uuid, layer);
		this.paint(uuid);
	}

	// Draws the region and moves the map to it.
	public showZone(zone: Models.RegionZone): void {
		this.add(zone);
		const layer = zone.uuid ? this.mLayers.get(zone.uuid) : null;
		if (layer) {
			this.mMap.fitBounds(layer.getBounds(), { padding: [16, 16] });
		}
	}

	public setSelected(uuid: string, selected: boolean): void {
		if (selected) {
			this.mSelected.add(uuid);
		} else {
			this.mSelected.delete(uuid);
		}
		this.paint(uuid);
	}

	private paint(uuid: string): void {
		const layer = this.mLayers.get(uuid);
		if (!layer) {
			return;
		}
		const selected = this.mSelected.has(uuid);
		layer.eachLayer(function (l: any) {
			l.getElement()?.classList.toggle("selected", selected);
		});
		if (selected) {
			layer.bringToFront();
		}
	}
}
