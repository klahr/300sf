/// <reference path="Translation.ts"/>

namespace Models {
	export class Announcement {
		uuid: string;
		name: string;
		content: Translation;
		from: Date;
		until: Date;
		color: string;
		icon: string;
		on_hold: Date;
	}
}