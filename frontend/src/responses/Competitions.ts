/// <reference path="../models/Zone.ts"/>
/// <reference path="../models/Translation.ts"/>

namespace Responses {
	export class CompetitionsBasisList {
		uuid: string;
		name: Models.Translation;
		description: Models.Translation;
		size: number;
	}

	export class CompetitionsBasis {
		zones: Array<Models.Zone>;
		list_templates: Array<CompetitionsBasisList>;
	}
}
