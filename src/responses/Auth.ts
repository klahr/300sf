namespace Responses {
	export class AuthenticationResponse {
		token_type: string;
		access_token: string;
		refresh_token: string;
		expires_in: number;
		lang: string;
	}
}