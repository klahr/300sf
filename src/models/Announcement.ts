/// <reference path="Translation.ts"/>

namespace Models {
	export class Announcement {
		uuid: string;
		name: string;
		content: Translation;
		from: Date;
		until: Date;
		priority: number;
		on_hold: Date;
	}
}