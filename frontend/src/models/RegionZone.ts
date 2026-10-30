/// <reference path="Translation.ts"/>

namespace Models {
	// A county or region to choose for a competition. From the map it is a zone (uuid); from a place
	// search it may not be one yet (osm_id only) until chosen.
	export class RegionZone {
		uuid?: string;
		osm_id?: string;
		name: Translation;
		country?: string;
		outline?: any;
	}
}
