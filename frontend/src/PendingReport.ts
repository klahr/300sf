/// <reference path="HttpRequest.ts"/>
/// <reference path="ReportPopUp.ts"/>

// A sighting started from the bird search on the landing page before signing in: the species and when
// it was seen, kept in the browser until the account exists. Opened once, within a day.
namespace PendingReport {
	const key = "pending-report";

	export function resume(): void {
		let pending: { uuid: string, at: string } | null = null;
		try {
			pending = JSON.parse(localStorage.getItem(key) || "null");
			localStorage.removeItem(key);
		} catch (e) {
			return;
		}
		if (!pending || !pending.uuid) {
			return;
		}
		const at = new Date(pending.at);
		if (isNaN(at.getTime()) || Date.now() - at.getTime() > 24 * 60 * 60 * 1000) {
			return;
		}
		new HttpRequest("GET", "/v1/species/" + encodeURIComponent(pending.uuid)).process(function (status: any, data: any) {
			openReportPopUp(data.result, null, function () { }, null, at);
		});
	}
}
