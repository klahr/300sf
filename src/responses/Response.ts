namespace Responses {
	export class SingleItemDataResponse {
		data: any;
	}

	export class GeneralErrorResponse {
		public message: string;
		public code: number;
	}
}