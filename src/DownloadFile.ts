function DownloadFile(filename: string, data: string, mimeType: string = "application/json"): void {
	let d = data.replace(/\\n/g, "\n");
	if (mimeType == "text/csv") {
		if (d.startsWith("\"") && d.endsWith("\"")) {
			d = d.slice(1, -1);
		}
	}
	const blob = new Blob([d], { type: mimeType });
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