/// <reference path="HttpRequest.ts"/>
/// <reference path="requests/Reports.ts"/>
/// <reference path="models/Translation.ts"/>
/// <reference path="ui/Toast.ts"/>

// Sightings reported without a connection wait here, in this browser, until they can be sent.
// Each carries its own uuid, so one that reached the server although the answer got lost is
// stored there only once.
namespace Outbox {
	interface Entry {
		user: string;
		species: Models.Translation;
		report: Requests.Reports;
	}

	const key = "outbox";
	let sending = false;

	function load(): Array<Entry> {
		try {
			return JSON.parse(localStorage.getItem(key) || "[]");
		} catch (e) {
			return [];
		}
	}

	function save(entries: Array<Entry>): void {
		localStorage.setItem(key, JSON.stringify(entries));
		window.dispatchEvent(new Event("outboxchanged"));
	}

	function mine(): Array<Entry> {
		return App.user ? load().filter(e => e.user === App.user.uuid) : [];
	}

	// 16 of the server's letters: long enough that no two devices pick the same.
	export function newUuid(): string {
		const letters = "abcdefghjkmnpqrstuvwxyz123456789";
		const bytes = crypto.getRandomValues(new Uint8Array(16));
		return Array.from(bytes, b => letters[b % letters.length]).join("");
	}

	export function add(report: Requests.Reports, species: Models.Translation): void {
		const entries = load();
		entries.push({ user: App.user.uuid, species: species, report: report });
		save(entries);
	}

	// Sightings of the signed-in user still to send.
	export function count(): number {
		return mine().length;
	}

	// Worth trying again later: no connection, a timeout or the server is busy.
	export function isTemporary(status: number): boolean {
		return status === 0 || status === 429 || status >= 500;
	}

	// Sends the signed-in user's sightings one at a time, oldest first, and stops at the first
	// that can't be sent now.
	export function send(): void {
		if (sending || !navigator.onLine || mine().length === 0) {
			return;
		}
		sending = true;
		let sent = 0;

		const finish = function () {
			sending = false;
			if (sent > 0) {
				const message = sent === 1 ? App.translations.OfflineSightingSent : App.translations.OfflineSightingsSent.replace("{0}", String(sent));
				new Ui.Toast(App.translations.BackOnline, message, null, "var(--success-color)", function () { });
				window.dispatchEvent(new Event("sightingadded"));
			}
		};

		const next = function () {
			const entry = mine()[0];
			if (!entry) {
				finish();
				return;
			}
			const remove = function () {
				save(load().filter(e => e.report.uuid !== entry.report.uuid));
			};
			const r = new HttpRequest("POST", "/v1/reports", entry.report);
			r.setQuiet();
			r.process(function () {
				remove();
				++sent;
				next();
			}, function (status: number) {
				if (isTemporary(status)) {
					finish();
					return;
				}
				// Refused, say because a choice in it no longer exists: sending again won't help.
				remove();
				new Ui.MessageBox(App.translations.Error, App.translations.OfflineSightingRefused.replace("{0}", Utils.escapeHtml(App.getTranslation(entry.species))), "/img/messagebox-256.webp", "var(--error-color)", function () { });
				next();
			});
		};
		next();
	}

	window.addEventListener("online", send);
}
