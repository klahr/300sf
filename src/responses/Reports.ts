/// <reference path="../models/Species.ts"/>
/// <reference path="../models/Age.ts"/>
/// <reference path="../models/Gender.ts"/>
/// <reference path="../models/Method.ts"/>
/// <reference path="../models/Activity.ts"/>
/// <reference path="../models/List.ts"/>
/// <reference path="../models/Tag.ts"/>

namespace Responses {
	export class ReportsBasis {
		public species: Array<Models.Species>;
		public ages: Array<Models.Age>;
		public genders: Array<Models.Gender>;
		public methods: Array<Models.Method>;
		public activities: Array<Models.Activity>;
		public lists: Array<Models.List>;
		public tags: Array<Models.Tag>;
	}
}
