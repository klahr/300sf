namespace Models {
	export class ApiKey {
		name: string;
		prefix: string;
		client_id: string;
		scopes: string;
		soft_limit: number;
		hard_limit: number;
		revoked_at: string | null;
		website: boolean;
	}
}
