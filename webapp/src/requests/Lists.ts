/// <reference path="../models/Translation.ts"/>

namespace Requests {
	export class Lists {
		public name: Models.Translation;
		public visibility: number;
		public template: string;
		public from: Date;
		public until: Date;
		public color: string;
	}
}