namespace Responses {
	export class ErrorCode {
		public static ErrorInternalServerError = 1000;
		public static ErrorRecordNotFound = 1001;
		public static ErrorInsert = 1002;
		public static ErrorDelete = 1003;

		public static ErrorParse = 2001;
		public static ErrorLinkExpired = 2002;

		public static ErrorLoggedOut = 3001;
		public static ErrorMissingApiKey = 3002;
		public static ErrorInvalidApiKey = 3003;
		public static ErrorTooManyApiCalls = 3004;
		public static ErrorMissingPermission = 3005;

		public static ErrorSignUpInvalidDisplayName = 4001;
		public static ErrorSignUpInvalidEmail = 4002;
		public static ErrorSignUpInvalidPassword = 4003;
		public static ErrorSignUpTermsNotAccepted = 4004;

		public static ErrorMigrateCompleted = 5001
		public static ErrorMigrateInvalidCredentials = 5002

		public static ErrorAccountDeactivated = 6001;
	}
}
