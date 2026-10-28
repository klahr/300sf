// The colour theme: "system" follows the device, "light" and "dark" are the user's choice. Kept in
// localStorage under "theme", which app-head.html reads before the first paint.
namespace Theme {
	const themes = ["system", "light", "dark"];

	export function get(): string {
		try {
			const theme = localStorage.getItem("theme");
			if (theme === "light" || theme === "dark") {
				return theme;
			}
		} catch (e) { }
		return "system";
	}

	export function set(theme: string): void {
		try {
			if (theme === "system") {
				localStorage.removeItem("theme");
			} else {
				localStorage.setItem("theme", theme);
			}
		} catch (e) { }
		if (theme === "system") {
			delete document.documentElement.dataset.theme;
		} else {
			document.documentElement.dataset.theme = theme;
		}
		const color = getComputedStyle(document.documentElement).getPropertyValue("--paper").trim();
		document.querySelectorAll("meta[name=theme-color]").forEach(e => {
			if (theme === "system") {
				e.setAttribute("content", e.getAttribute("media")?.includes("light") ? "#f5f3f7" : "#0b0a0d");
			} else {
				e.setAttribute("content", color);
			}
		});		window.dispatchEvent(new Event("themechange"));
	}

	export function cycle(): void {
		set(themes[(themes.indexOf(get()) + 1) % themes.length]);
	}
}
