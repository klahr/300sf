namespace Requests {
	export class Reports {
		public species_uuid: string;
		public from: string;
		public until: string;
		public latitude: number;
		public longitude: number;
		public quantity: number;
		public age_uuid: string;
		public activity_uuid: string;
		public gender_uuid: string;
		public method_uuid: string;
		public comment: string;
		public list_uuids: Array<string>;
		public tag_uuids: Array<string>;
		public silent: boolean;
		public notification_comment: string;
	}
}