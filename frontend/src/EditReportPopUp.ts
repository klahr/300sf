/// <reference path="responses/Reports.ts"/>
/// <reference path="responses/Lists.ts"/>
/// <reference path="requests/Reports.ts"/>
/// <reference path="models/Species.ts"/>
/// <reference path="models/List.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="ui/Sheet.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="utils/Sort.ts"/>
/// <reference path="SpeciesPopUp.ts"/>

// One sighting, read only: the species photo, what was seen when and where, and the lists and labels
// it counts for.
class EditReportPopUp {
	private mSheet: Ui.Sheet;

	constructor(report: Models.Report, closeCallback: () => void) {
		const self = this;

		const sheet = new Ui.Sheet({ eyebrow: App.translations.Sighting, title: App.getTranslation(report.species.name), subtitle: report.species.name.la, latin: true, hero: true });
		this.mSheet = sheet;

		// The species page, for the photos and the article.
		const speciesChip = document.createElement("button");
		speciesChip.className = "sheet-chip";
		speciesChip.innerHTML = App.translations.AboutSpecies;
		speciesChip.addEventListener("click", function () {
			openSpeciesPopUp(report.species, null, function () { });
		});
		sheet.names.insertBefore(speciesChip, sheet.title);
		sheet.names.querySelector("p.sheet-eyebrow")?.remove();

		// Label and value pairs; empty values are left out.
		const facts = sheet.section(App.translations.Sighting);
		const details = document.createElement("div");
		details.className = "detail-grid";
		facts.appendChild(details);
		const addDetail = function (label: string, value: string, icon: string) {
			if (!value) {
				return;
			}
			const item = document.createElement("div");
			const l = document.createElement("p");
			l.className = "detail-label";
			l.innerHTML = "<span class='material-symbols-outlined'>" + icon + "</span>" + label;
			item.appendChild(l);
			const v = document.createElement("p");
			v.className = "detail-value";
			v.textContent = value;
			item.appendChild(v);
			details.appendChild(item);
		};
		if (Utils.isValidDate(report.from)) {
			addDetail(App.translations.ObservedFrom, Utils.toYYYY_MM_DD_hh_mm(report.from), "schedule");
		}
		if (Utils.isValidDate(report.until)) {
			addDetail(App.translations.ObservedUntil, Utils.toYYYY_MM_DD_hh_mm(report.until), "schedule");
		}
		if (report.zone) {
			addDetail(App.translations.ObservationLocation, App.getTranslation(report.zone.name), "location_on");
		}
		addDetail(App.translations.Site, report.site, "pin_drop");
		addDetail(App.translations.Quantity, report.quantity.toString(), "tag");
		if (report.age) {
			addDetail(App.translations.Age, App.getTranslation(report.age.name), "cake");
		}
		if (report.gender) {
			addDetail(App.translations.Gender, App.getTranslation(report.gender.name), "wc");
		}
		if (report.method) {
			addDetail(App.translations.Method, App.getTranslation(report.method.name), "visibility");
		}
		if (report.activity) {
			addDetail(App.translations.Activity, App.getTranslation(report.activity.name), "directions_run");
		}

		if (report.comment.text !== "") {
			const comment = document.createElement("p");
			comment.className = "comment";
			comment.textContent = report.comment.text;
			facts.appendChild(comment);
		}

		if (report.latitude !== 0 && report.longitude !== 0) {
			const where = sheet.section(App.translations.ObservationLocation);
			const mapDiv = new Ui.Div();
			mapDiv.container.classList.add("map");
			mapDiv.style.height = "220px";
			mapDiv.style.position = "relative";
			where.appendChild(mapDiv.container);

			setTimeout(function () {
				const latLng = [report.latitude, report.longitude];
				const map = L.map(mapDiv.container).setView(latLng, 13);
				L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
					maxZoom: 19,
					attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
				}).addTo(map);
				const markerGroup = L.layerGroup().addTo(map);
				L.marker(latLng).addTo(markerGroup);
			}, 10);
		}

		const lists = sheet.section(App.translations.Checklists);
		const listChips = document.createElement("div");
		listChips.className = "chips";
		lists.appendChild(listChips);
		report.lists.forEach(e => {
			const chip = document.createElement("span");
			chip.className = "chip";
			chip.innerHTML = "<span class='list-dot' style='--dot: " + Utils.escapeHtml(e.color !== "" ? e.color : "var(--blue)") + ";'></span>" + Utils.escapeHtml(App.getTranslation(e.name));
			listChips.appendChild(chip);
		});

		if (report.tags.length > 0) {
			const tags = sheet.section(App.translations.Tags);
			const tagChips = document.createElement("div");
			tagChips.className = "chips";
			tags.appendChild(tagChips);
			report.tags.forEach(e => {
				const tag = document.createElement("div");
				tag.className = "tag";
				if (e.color !== "") {
					tag.style.backgroundColor = e.color;
				}
				const c = document.createElement("p");
				c.textContent = e.name;
				tag.appendChild(c);
				tagChips.appendChild(tag);
			});
		}

		sheet.addAction(Ui.actionButton(App.translations.ArchiveSighting, "archive", "secondary")).addEventListener("click", function () {
			new Ui.ConfirmBox(App.translations.ArchiveSighting, App.translations.ConfirmArchiveSighting, "/img/messagebox-256.webp", "var(--error-color)", function () {
				const r = new HttpRequest("POST", "/v1/reports/" + report.uuid + "/archive");
				r.process(function (status: any, data: any) {
					self.mSheet.close();
					setTimeout(() => {
						location.reload();
					}, 150);
				});
			}, function () {
			}, App.translations.ArchiveSighting, App.translations.Cancel);
		});

		// The species photo for the header.
		new HttpRequest("GET", "/v1/species/" + report.species.uuid).process(function (status: any, data: any) {
			sheet.setHeroImage(wikipediaHeroUrl(data.wikipedia || {}));
		}, function () {
			sheet.setHeroImage("");
		});

		sheet.open(function () {
			closeCallback();
		});
	}
}

function openEditReportPopUp(report: Models.Report, closeCallback: () => void = function () { }): void {
	new EditReportPopUp(report, closeCallback);
}
