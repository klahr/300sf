namespace Responses {
	export class AuthenticationResponse {
		token_type: string;
		access_token: string;
		expires_at: number;
		max_age: number;
		lang: string;
	}
}