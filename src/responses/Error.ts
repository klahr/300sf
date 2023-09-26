namespace Responses {
	export class ErrorCode {
		public static ErrorInternalServerError = 1000;
		public static ErrorRecordNotFound      = 1001;
		public static ErrorInsert              = 1002;
		public static ErrorDelete              = 1003;
	
		public static ErrorParse = 2001;
	
		public static ErrorLoggedOut         = 3001;
		public static ErrorMissingApiKey     = 3002;
		public static ErrorInvalidApiKey     = 3003;
		public static ErrorTooManyApiCalls   = 3004;
		public static ErrorMissingPermission = 3005;
	}

	export class Error {
		public message: string;
		public code: number;
	}
}
