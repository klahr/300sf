/// <reference path="LocalStorage.ts"/>
/// <reference path="responses/Auth.ts"/>

// The session lives in cookies set by the server. The tokens themselves are
// HttpOnly, so this code (and anything injected into the page) never sees
// them. The readable "sf_session" cookie only says that a session exists and
// when its access token expires.
class Session {
	// Refresh a little early so a request doesn't leave with a token that
	// expires on the way.
	private static readonly SkewMs = 30 * 1000;
	private static sRefreshing: Array<(ok: boolean) => void> | null = null;

	private static sessionCookie(): string | null {
		const match = document.cookie.match(/(?:^|;\s*)sf_session=([^;]*)/);
		return match ? decodeURIComponent(match[1]) : null;
	}

	// True while there is a session, even if its access token needs renewing.
	public static get exists(): boolean {
		return Session.sessionCookie() !== null;
	}

	private static get accessExpired(): boolean {
		const expiresAt = parseInt(Session.sessionCookie() || "0");
		return isNaN(expiresAt) || Date.now() > expiresAt - Session.SkewMs;
	}

	// Called after signing in or refreshing; the tokens are already cookies.
	public static store(response: Responses.AuthenticationResponse): void {
		const ks = new LocalStorage();
		ks.set("lang", response.lang, 0);
		ks.set("scope", response.scope, 0);
	}

	// Removes what older versions kept in localStorage, including tokens.
	public static forgetLegacyTokens(): void {
		const ks = new LocalStorage();
		for (const key of ["x300sf_at", "x300sf_rt", "stay"]) {
			ks.remove(key);
		}
	}

	private static refreshRequest(): HttpRequest {
		const r = new HttpRequest("POST", "/v1/auth", { grant_type: "refresh_token" });
		r.setNoRetry();
		return r;
	}

	// Used once at startup, before any page is built, so pages see the right
	// signed-in state.
	public static refreshSync(): boolean {
		if (!Session.exists) {
			return false;
		}
		if (!Session.accessExpired) {
			return true;
		}
		let ok = false;
		const r = Session.refreshRequest();
		r.setSync();
		r.process(function (status: any, response: Responses.AuthenticationResponse) {
			Session.store(response);
			ok = true;
		});
		return ok;
	}

	// Refreshes once even when many requests fail at the same time; every
	// caller is told whether it worked.
	public static refresh(done: (ok: boolean) => void): void {
		if (!Session.exists) {
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
		Session.refreshRequest().process(function (status: any, response: Responses.AuthenticationResponse) {
			Session.store(response);
			finish(true);
		}, function () {
			finish(false);
		});
	}

	// Ends the session on the server, then runs done.
	public static signOut(done: () => void): void {
		const r = new HttpRequest("POST", "/v1/auth/logout");
		r.setNoRetry();
		r.process(function () { }, function () { }, function () {
			new LocalStorage().remove("scope");
			done();
		});
	}

	// Sends the user to sign in, then back to where they were.
	public static signInAgain(): void {
		new LocalStorage().remove("scope");
		const here = window.location.pathname + window.location.search;
		const redirect = here.startsWith("/home") ? "" : "?redirect=" + encodeURIComponent(here);
		window.open("/home" + redirect, "_self");
	}
}
