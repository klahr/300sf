/// <reference path="Language.ts"/>

namespace Models {
	export class User {
		uuid: string;
		display_name: string;
		imported_at: Date;
		language: Language;
		portrait: string;
	}
}