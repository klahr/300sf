function DownloadFile(filename: string, data: string, mimeType: string = "application/json"): void {
	let d = data.replace(/\\n/g, "\n");
	if (mimeType == "text/csv") {
		if (d.startsWith("\"") && d.endsWith("\"")) {
			d = d.slice(1, -1);
		}
	}

	const utf8BOM = new Uint8Array([0xEF, 0xBB, 0xBF]);

	const utf8Encoder = new TextEncoder();
	const utf8Data = utf8Encoder.encode(d);
	const blob = new Blob([utf8BOM, utf8Data], { type: mimeType + ";charset=utf-8" });
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