/// <reference path="Species.ts"/>

namespace Models {
	export class List {
		uuid: string;
		name: Translation;
		species: Array<Species>;
		color: string;
	}
}
