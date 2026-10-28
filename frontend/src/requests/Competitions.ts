namespace Requests {
	export class Competitions {
		name: Models.Translation;
		description: Models.Translation;
		from: string;
		until: string;
		list_template_uuid: string;
		end_on_reports: number;
		max_competitors: number;
		visibility: number;
		color: string;
		competition_zones_uuids: Array<string>;
		password: string;
		creator_join: boolean;
	}
}
