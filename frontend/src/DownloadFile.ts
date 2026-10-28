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
