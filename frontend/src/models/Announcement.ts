namespace Models {
	export class LocalizedText {
		sv: string | null;
		en: string | null;
		de: string | null;
	}

	// A post in the news feed, as a reader sees it.
	export class NewsPost {
		uuid: string;
		title: LocalizedText;
		body: LocalizedText;
		icon: string;
		color: string;
		published_at: Date;
		read: boolean;
	}

	// A post as the admin tool sees it: drafts and old posts too.
	export class AdminNewsPost {
		uuid: string;
		title: LocalizedText;
		body: LocalizedText;
		icon: string;
		color: string;
		from: Date | null;
		until: Date | null;
		on_hold: Date | null;
		published_at: Date;
		reads: number;
	}
}
