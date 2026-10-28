/// <reference path="Translation.ts"/>

namespace Models {
	export class Zone {
		uuid: string;
		name: Translation;
		polygon: string;
		admin_level: number;
		parent_zone: Zone;
	}
}
