/// <reference path="LocalStorage.ts"/>
/// <reference path="responses/Auth.ts"/>

// Stores the tokens from /v1/auth and refreshes the access token when it runs
// out. The access token always expires when the API says it does; "stay signed
// in" decides how long the refresh token is kept.
class Session {
	private static readonly StaySignedInSeconds = 3600 * 24 * 30;
	private static readonly SessionSeconds = 3600 * 12;
	// Refresh a little early so a request doesn't leave with a token that
	// expires on the way.
	private static readonly SkewSeconds = 30;
	private static sRefreshing: Array<(ok: boolean) => void> | null = null;

	public static store(token: Responses.AuthenticationResponse, staySignedIn: boolean): void {
		const ks = new LocalStorage();
		const accessSeconds = Math.max(token.expires_in - Session.SkewSeconds, 1);
		ks.set("x300sf_at", token.access_token, accessSeconds);
		ks.set("x300sf_rt", token.refresh_token, staySignedIn ? Session.StaySignedInSeconds : Session.SessionSeconds);
		ks.set("stay", staySignedIn ? "1" : "0", 0);
		ks.set("lang", token.lang, 0);
		ks.set("scope", token.scope, 0);
	}

	public static get accessToken(): string | null {
		return new LocalStorage().get("x300sf_at");
	}

	public static get refreshToken(): string | null {
		return new LocalStorage().get("x300sf_rt");
	}

	public static clear(): void {
		const ks = new LocalStorage();
		for (const key of ["x300sf_at", "x300sf_rt", "stay", "scope"]) {
			ks.remove(key);
		}
	}

	private static refreshBody(): object {
		return {
			grant_type: "refresh_token",
			refresh_token: Session.refreshToken
		};
	}

	// Used once at startup, before any page is built, so pages see the right
	// signed-in state.
	public static refreshSync(): boolean {
		if (Session.accessToken !== null || Session.refreshToken === null) {
			return Session.accessToken !== null;
		}
		let ok = false;
		const r = new HttpRequest("POST", "/v1/auth", Session.refreshBody());
		r.setSync();
		r.setAnonymous();
		r.process(function (status: any, token: Responses.AuthenticationResponse) {
			Session.store(token, new LocalStorage().get("stay") === "1");
			ok = true;
		}, function () {
			Session.clear();
		});
		return ok;
	}

	// Refreshes once even when many requests fail at the same time; every
	// caller is told whether it worked.
	public static refresh(done: (ok: boolean) => void): void {
		if (Session.refreshToken === null) {
			done(false);
			return;
		}
		if (Session.sRefreshing !== null) {
			Session.sRefreshing.push(done);
			return;
		}
		Session.sRefreshing = [done];
		const finish = function (ok: boolean) {
			const waiting = Session.sRefreshing!;
			Session.sRefreshing = null;
			waiting.forEach(cb => cb(ok));
		};
		const r = new HttpRequest("POST", "/v1/auth", Session.refreshBody());
		r.setAnonymous();
		r.process(function (status: any, token: Responses.AuthenticationResponse) {
			Session.store(token, new LocalStorage().get("stay") === "1");
			finish(true);
		}, function () {
			Session.clear();
			finish(false);
		});
	}

	// Sends the user to sign in, then back to where they were.
	public static signInAgain(): void {
		Session.clear();
		const here = window.location.pathname + window.location.search;
		const redirect = here === "/" || here.startsWith("/?") || here.startsWith("/home") ? "" : "?redirect=" + encodeURIComponent(here);
		window.open("/" + redirect, "_self");
	}
}
