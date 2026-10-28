function DownloadFile(filename: string, data: string, mimeType: string = "application/json"): void {
	const blob = new Blob([new TextEncoder().encode(data)], { type: mimeType + ";charset=utf-8" });
	const url = URL.createObjectURL(blob);
	const a = document.createElement("a");
	a.href = url;
	a.download = filename;
	a.style.display = 'none';
	document.body.appendChild(a);
	a.click();
	document.body.removeChild(a);
	URL.revokeObjectURL(url);
}

namespace Utils {
	// A file name that says what it is and when it was made: "300 Svenska Fåglar" exported today
	// becomes 300sf-300-svenska-faglar-2026-10-28.csv.
	export function exportFileName(name: string, extension: string): string {
		// Accents are split off by NFD and removed: U+0300 to U+036F are the combining marks.
		const marks = new RegExp("[" + String.fromCharCode(0x300) + "-" + String.fromCharCode(0x36f) + "]", "g");
		const slug = name.toLowerCase().normalize("NFD").replace(marks, "").replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "");
		const now = new Date();
		const date = now.getFullYear() + "-" + String(now.getMonth() + 1).padStart(2, "0") + "-" + String(now.getDate()).padStart(2, "0");
		return "300sf-" + (slug !== "" ? slug + "-" : "") + date + "." + extension;
	}
}
