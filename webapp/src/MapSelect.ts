/// <reference path="ui/Widget.ts"/>
/// <reference path="models/Zone.ts"/>

class MapSelect extends Ui.Widget {
	private mMap: any;
	private mGeoJsonLayers = Array<any>();
	private mCallback: (zoneUuids: Array<string>) => void;

	constructor(callback: (zoneUuids: Array<string>) => void) {
		super();
		this.mCallback = callback;

		this.mContainer = document.createElement("div");
		this.mContainer.id = "map";
		this.mContainer.style.height = "500px";
		this.mContainer.style.width = "100%";
		this.mContainer.style.position = "relative";
		this.mContainer.style.borderRadius = "var(--small-distance)";

		const self = this;
		setTimeout(function () {
			self.mMap = L.map('map', {
				doubleClickZoom: false
			}).setView([62.0, 15.0], 4);
			L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
				maxZoom: 19,
				attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
			}).addTo(self.mMap);
			const markerGroup = L.layerGroup().addTo(self.mMap);
		}, 10);
	}

	public addZone(zone: Models.Zone): void {
		const self = this;
		const l = L.geoJSON(JSON.parse(Utils.decodeBase64(zone.polygon)), {
			onEachFeature: function (feature, layer) {
				feature.properties.uuid = zone.uuid;
				layer.on('click', function () {
					if (!feature.properties.hasOwnProperty('selected') || !feature.properties.selected) {
						feature.properties.selected = true;
						layer.setStyle({ color: "var(--success-color)", fillColor: 'var(--success-color)', fillOpacity: 0.5 }); // Change styling for selected state
					} else {
						feature.properties.selected = false;
						layer.setStyle({ color: "#777", fillColor: '#777', fillOpacity: 0.2 }); // Change styling for unselected state
					}
					self.mCallback(self.getSelectedZoneUuids());
				});
			},
			style: function () {
				return {
					color: '#777',
					fillColor: '#777',
					fillOpacity: 0.2
				};
			}
		}).addTo(this.mMap);
		this.mGeoJsonLayers.push(l);
	}

	public setZoneSelected(uuid: string): void {
		this.mGeoJsonLayers.forEach(l => {
			l.eachLayer(function (layer) {
				const f = layer.feature;
				if (f.properties.uuid === uuid && !f.properties.selected) {
					layer.fireEvent("click");
				}
			});
		});
	}

	public getSelectedZoneUuids(): Array<string> {
		const uuids = Array<string>();
		this.mGeoJsonLayers.forEach(l => {
			l.eachLayer(function (layer) {
				const f = layer.feature;
				if (f.properties.selected) {
					uuids.push(f.properties.uuid);
				}
			});
		});
		return uuids;
	}
}