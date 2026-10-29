namespace Utils {
	// Escapes text for insertion into innerHTML. Use for anything that comes
	// from users, the API or third parties (Wikipedia, xeno-canto).
	export function escapeHtml(value: any): string {
		return String(value ?? "")
			.replace(/&/g, "&amp;")
			.replace(/</g, "&lt;")
			.replace(/>/g, "&gt;")
			.replace(/"/g, "&quot;")
			.replace(/'/g, "&#39;");
	}

	// Returns url if it is an http(s) URL, otherwise null.
	export function safeUrl(url: any): string | null {
		try {
			const u = new URL(String(url), window.location.href);
			return u.protocol === "https:" || u.protocol === "http:" ? u.href : null;
		} catch (e) {
			return null;
		}
	}

	// Only same-origin paths are allowed as post-login redirects.
	// The browser decides where a path leads (it drops tabs and newlines and
	// treats "\\" as "/"), so resolve it the same way and check the origin.
	export function safeRedirect(path: string | null): string | null {
		if (!path || !path.startsWith("/")) {
			return null;
		}
		try {
			const u = new URL(path, window.location.origin);
			return u.origin === window.location.origin ? u.pathname + u.search + u.hash : null;
		} catch (e) {
			return null;
		}
	}
}
