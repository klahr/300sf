/// <reference path="Language.ts"/>

namespace Models {
	export class User {
		uuid: string;
		display_name: string;
		permissions: number;
		imported_at: Date;
		language: Language;
		portrait: string;
	}
}