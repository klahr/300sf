declare const L;

// The maps' tiles, from the server's config (site.map-tiles), so the tile service can change without
// a new app. OpenStreetMap's by default; their policy asks for its attribution on every map.
namespace Maps {
	let tiles = "https://tile.openstreetmap.org/{z}/{x}/{y}.png";
	let attribution = "&copy; <a href=\"https://www.openstreetmap.org/copyright\">OpenStreetMap</a> contributors";

	export function configure(tileUrl: string, tileAttribution: string): void {
		if (tileUrl) {
			tiles = tileUrl;
			attribution = tileAttribution || "";
		}
	}

	export function tileLayer(): any {
		return L.tileLayer(tiles, { maxZoom: 19, attribution: attribution });
	}
}
