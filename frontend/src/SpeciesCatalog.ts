/// <reference path="HttpRequest.ts"/>
/// <reference path="models/Species.ts"/>

// All species, plus what search ranks first: species on the user's lists, then Swedish species (every
// species on a list template). Loaded once and shared by the overview search and the report sheet.
namespace SpeciesCatalog {
	export interface Catalog {
		species: Array<Models.Species>;
		preferred: Set<string>;
		swedish: Set<string>;
	}

	let catalog: Catalog | null = null;
	let loading = false;
	let waiting: Array<(c: Catalog) => void> = [];

	export function get(callback: (c: Catalog) => void): void {
		if (catalog) {
			callback(catalog);
			return;
		}
		waiting.push(callback);
		if (loading) {
			return;
		}
		loading = true;

		const result: Catalog = { species: [], preferred: new Set<string>(), swedish: new Set<string>() };
		let pending = 3;
		const finish = function () {
			if (--pending > 0) {
				return;
			}
			catalog = result;
			loading = false;
			const callbacks = waiting;
			waiting = [];
			callbacks.forEach(cb => cb(result));
		};

		// Counted on success or failure rather than completion: a request retried after refreshing
		// the session completes twice.
		new HttpRequest("GET", "/v1/species/list").process(function (status: any, data: any) {
			result.species = data.result || [];
			finish();
		}, finish);
		new HttpRequest("GET", "/v1/list_templates/list").process(function (status: any, data: any) {
			(data.result || []).forEach(t => (t.species || []).forEach(s => result.swedish.add(s.uuid)));
			finish();
		}, finish);
		loadPreferred(result, finish);
	}

	function loadPreferred(target: Catalog, done: () => void): void {
		new HttpRequest("GET", "/v1/lists/list").process(function (status: any, data: any) {
			// Updated in place: search holds on to this set.
			target.preferred.clear();
			(data.result || []).forEach(l => (l.species || []).forEach(s => target.preferred.add(s.uuid)));
			done();
		}, done);
	}

	// The user's lists changed, so the species to rank first may have too.
	window.addEventListener("listschanged", function () {
		if (catalog) {
			loadPreferred(catalog, function () { });
		}
	});
}
