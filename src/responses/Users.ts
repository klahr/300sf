namespace Responses {
	export class ListReport {
		list: Models.List;
		reports: Array<Models.Report>;
	}

	export class UsersMe {
		list_reports: Array<ListReport>;
		last_reports: Array<Models.Report>;
		user: Models.User;
	}

	export class UsersWhoAmI {
		message: string;
		data: Models.User;
	}
}