/// <reference path="Translation.ts"/>

namespace Models {
	export class CompetitionPerson {
		display_name: string;
		username: string;
		portrait: string;
	}

	export class CompetitionStanding {
		rank: number;
		user: CompetitionPerson;
		count: number;
		latest: string | null;
		me: boolean;
	}

	export class Competition {
		uuid: string;
		name: Translation;
		description: Translation;
		color: string;
		from: string;
		until: string;
		join_until: string;
		end_on_reports: number | null;
		max_competitors: number | null;
		visibility: number;
		has_password: boolean;
		list_name: Translation | null;
		list_size: number;
		zones: Array<Translation>;
		creator: CompetitionPerson;
		participants: number;
		status: "upcoming" | "ongoing" | "finished";
		ended_at: string | null;
		joined: boolean;
		is_creator: boolean;
		invited: boolean;
		can_join: boolean;
		my_rank: number;
		my_count: number;
		standings: Array<CompetitionStanding> | null;
	}
}
