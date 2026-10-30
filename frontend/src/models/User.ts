/// <reference path="Language.ts"/>

namespace Models {
	export class User {
		uuid: string;
		display_name: string;
		username: string;
		language: Language;
		portrait: string;
		scopes?: string;
		last_seen_at?: string;
	}
}