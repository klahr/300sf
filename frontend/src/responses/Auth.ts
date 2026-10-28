namespace Responses {
	// The app's sign-in answer; the tokens themselves are set as cookies.
	export class AuthenticationResponse {
		expires_in: number;
		lang: string;
		scope: string;
	}
}