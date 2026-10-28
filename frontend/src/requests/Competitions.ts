namespace Requests {
	export class Competitions {
		name: Models.Translation;
		from: string;
		until: string;
		list_template_uuid: string;
		description: Models.Translation;
		end_on_reports: number;
		max_competitors: number;
		visibility: number;
		join_from: string;
		join_until: string;
		icon: string;
		color: string;
		competition_zones_uuids: Array<string>;
		password: string;
		creator_join: boolean;
	}
}