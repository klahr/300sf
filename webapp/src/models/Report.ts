/// <reference path="Zone.ts"/>
/// <reference path="Activity.ts"/>
/// <reference path="Age.ts"/>
/// <reference path="Gender.ts"/>
/// <reference path="Method.ts"/>
/// <reference path="Tag.ts"/>
/// <reference path="Species.ts"/>
/// <reference path="List.ts"/>
/// <reference path="Comment.ts"/>

namespace Models {
	export class Report {
		uuid: string;
		species: Species;
		from: Date;
		until: Date;
		latitude: number;
		longitude: number;
		zone: Zone;
		quantity: number;
		age: Age;
		activity: Activity;
		gender: Gender;
		method: Method;
		comment: Comment;
		lists: Array<List>;
		tags: Array<Tag>;
		site: string;
//		notification: Notification;
	}
}
