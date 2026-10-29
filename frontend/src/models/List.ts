/// <reference path="Species.ts"/>

namespace Models {
	export class List {
		uuid: string;
		name: Translation;
		species: Array<Species>;
		color: string;
		from: Date;
		until: Date;
		template?: string;
	}
}
