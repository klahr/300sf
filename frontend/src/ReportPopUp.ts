/// <reference path="responses/Reports.ts"/>
/// <reference path="responses/Lists.ts"/>
/// <reference path="requests/Reports.ts"/>
/// <reference path="models/Species.ts"/>
/// <reference path="models/List.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="ui/Sheet.ts"/>
/// <reference path="ui/Toast.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="utils/Sort.ts"/>
/// <reference path="NewTagPopUp.ts"/>
/// <reference path="Onboarding.ts"/>
/// <reference path="SpeciesCatalog.ts"/>
/// <reference path="AnalyzePopUp.ts"/>
/// <reference path="ui/SpeciesSearch.ts"/>

declare const L;

class ReportPopUp {
	private mReportButton: HTMLButtonElement;
	private mSpeciesList: Ui.Input;
	private mChecklistContainer: Ui.Div;
	private mTagsContainer: Ui.Div;
	// Null until the user taps the map or their position is found.
	private mLocation: { lat: number, lng: number } | null = null;
	private mPopUp: Ui.Sheet;

	constructor(species: Models.Species, list: Models.List | null, closeCallback: (speciesAdded: boolean) => void) {
		const self = this;

		var speciesAdded = false;

		const sheet = new Ui.Sheet({ eyebrow: App.translations.SubmitSighting, title: App.getTranslation(species.name), subtitle: species.name.la, latin: true });
		this.mPopUp = sheet;

		this.mChecklistContainer = new Ui.Div();
		this.mChecklistContainer.style.position = "relative";

		// Which lists. Optional: a sighting on no list still counts in competitions.
		const listsSection = sheet.section(App.translations.AvailableChecklists);
		listsSection.appendChild(self.mChecklistContainer.container);

		// The map looks up the position after a moment; stop that if the sheet closes first.
		let closed = false;
		let map: any = null;

		// Where: the map, centred on the user's position when it can be found.
		const where = sheet.section(App.translations.ObservationLocation);
		const mapDiv = new Ui.Div();
		mapDiv.container.classList.add("map");
		mapDiv.style.height = "240px";
		mapDiv.style.position = "relative";
		where.appendChild(mapDiv.container);

		const locateLabel = "<span class='material-symbols-outlined'>my_location</span>" + App.translations.FindMyLocation;
		const locateButton = document.createElement("button");
		locateButton.className = "secondary";
		locateButton.innerHTML = locateLabel;
		where.appendChild(locateButton);

		const positionNote = document.createElement("p");
		positionNote.className = "details hidden";
		positionNote.textContent = App.translations.NoPositionYet;
		where.appendChild(positionNote);

		setTimeout(function () {
			if (closed) {
				return;
			}
			map = L.map(mapDiv.container).setView([58.1801188016368, 13.959220108737176], 13);
			L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
				maxZoom: 19,
				attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
			}).addTo(map);
			const markerGroup = L.layerGroup().addTo(map);
			map.on("click", function (e) {
				markerGroup.clearLayers();
				L.marker(e.latlng).addTo(markerGroup);
				self.mLocation = e.latlng;
				positionNote.classList.add("hidden");
			});
			map.on('locationfound', function (e) {
				markerGroup.clearLayers();
				L.marker(e.latlng).addTo(markerGroup);
				self.mLocation = e.latlng;
				map.setZoom(16);
				Ui.setBusy(locateButton, false);
				positionNote.classList.add("hidden");
			});
			// Not an error worth a dialog: the position is optional and can be set on the map.
			map.on('locationerror', function () {
				Ui.setBusy(locateButton, false);
				if (self.mLocation === null) {
					positionNote.classList.remove("hidden");
				}
			});

			locateButton.addEventListener("click", function () {
				Ui.setBusy(locateButton, true);
				map.locate({ setView: true, maxZoom: 19 });
			});
			setTimeout(function() {
				if (!closed) {
					locateButton.click();
				}
			}, 1000);
		}, 10);

		// The species is fixed; the input only carries it to the request and the toast.
		this.mSpeciesList = new Ui.Input("text");
		this.mSpeciesList.value = App.getTranslation(species.name);
		this.mSpeciesList.setData(species);

		const loadLists = function () {
			new HttpRequest("GET", "/v1/lists/species/" + species.uuid).process(function (status: any, data: any) {
				const d = data as Responses.ListsSpecies;
				self.mChecklistContainer.detachAll();
				self.mChecklistContainer.container.innerHTML = "";

				// No checklist at all: offer to start one right here.
				if (!d.result || d.result.length === 0) {
					const intro = document.createElement("p");
					intro.className = "details";
					intro.textContent = App.translations.NoChecklistForReport;
					self.mChecklistContainer.container.appendChild(intro);
					self.mChecklistContainer.container.appendChild(Onboarding.starterCard(function () {
						loadLists();
					}));
					self.validate();
					return;
				}
				if (!d.result.some(e => e.available)) {
					const none = document.createElement("p");
					none.className = "details";
					none.textContent = App.translations.NotOnYourLists;
					self.mChecklistContainer.container.appendChild(none);
				}

				(d.result || []).forEach(e => {
					const dd = new Ui.Div();
					dd.addClass("check-row");
					self.mChecklistContainer.attach(dd);
					const c = new Ui.Input("checkbox");
					c.container.setAttribute("name", e.list.uuid);
					c.container.id = "report-list-" + e.list.uuid;
					c.container.addEventListener("click", function () {
						self.validate();
					});

					dd.attach(c);
					const cLbl = document.createElement("label");
					cLbl.setAttribute("for", c.container.id);
					cLbl.textContent = App.getTranslation(e.list.name);
					dd.container.appendChild(cLbl);

					if (!e.available) {
						c.container.setAttribute("disabled", "");
						dd.addClass("unavailable");
					} else if (list && e.list.uuid === list.uuid) {
						(c.container as HTMLInputElement).checked = true;
						(c.container as HTMLInputElement).disabled = true;
					}
				});

				// With only one list that can hold the species, choose it for the user.
				const open = Array.from(self.mChecklistContainer.container.querySelectorAll("input")).filter(c => !c.disabled);
				if (open.length === 1 && !self.mChecklistContainer.container.querySelector("input:checked")) {
					open[0].checked = true;
				}

				self.validate();
			});
		};
		loadLists();

		// When and how many.
		const when = sheet.section(App.translations.Sighting);
		const whenGrid = document.createElement("div");
		whenGrid.className = "sheet-grid";
		when.appendChild(whenGrid);

		const from = new Ui.Input("datetime-local");
		from.container.id = "report-from";
		sheet.field(whenGrid, App.translations.ObservedFrom, from.container);

		const now = new Date();
		now.setMinutes(now.getMinutes() - now.getTimezoneOffset());
		now.setSeconds(now.getMinutes());
		from.value = now.toISOString().slice(0, -8);

		const until = new Ui.Input("datetime-local");
		until.container.id = "report-until";
		sheet.field(whenGrid, App.translations.ObservedUntil, until.container);

		const quantity = new Ui.Input("number");
		quantity.container.id = "report-quantity";
		quantity.container.setAttribute("min", "1");
		quantity.container.setAttribute("step", "1");
		quantity.value = "1";
		sheet.field(whenGrid, App.translations.Quantity, quantity.container);

		// Details and labels are optional; folded away so a quick tick stays short.
		const moreToggle = document.createElement("button");
		moreToggle.className = "text-button details-toggle";
		moreToggle.innerHTML = "<span class='material-symbols-outlined'>expand_more</span>" + App.translations.MoreDetails;
		when.appendChild(moreToggle);

		const details = sheet.section(App.translations.Details);
		details.classList.add("hidden");
		const detailGrid = document.createElement("div");
		detailGrid.className = "sheet-grid";
		details.appendChild(detailGrid);

		const ageList = new Ui.Select(App.translations.SelectAge + "...", App.translations.Search);
		sheet.field(detailGrid, App.translations.Age, ageList.container);

		const genderList = new Ui.Select(App.translations.SelectGender + "...", App.translations.Search);
		sheet.field(detailGrid, App.translations.Gender, genderList.container);

		const methodList = new Ui.Select(App.translations.SelectMethod + "...", App.translations.Search);
		sheet.field(detailGrid, App.translations.Method, methodList.container);

		const activityList = new Ui.Select(App.translations.SelectActivity + "...", App.translations.Search);
		sheet.field(detailGrid, App.translations.Activity, activityList.container);

		const site = new Ui.Input("text");
		site.container.id = "report-site";
		sheet.field(details, App.translations.Site, site.container);

		const privateComment = document.createElement("textarea");
		privateComment.id = "report-comment";
		privateComment.rows = 3;
		sheet.field(details, App.translations.PrivateComment, privateComment);

		const tagsSection = sheet.section(App.translations.Tags);
		tagsSection.classList.add("hidden");
		moreToggle.addEventListener("click", function () {
			const show = details.classList.contains("hidden");
			details.classList.toggle("hidden", !show);
			tagsSection.classList.toggle("hidden", !show);
			moreToggle.innerHTML = "<span class='material-symbols-outlined'>" + (show ? "expand_less" : "expand_more") + "</span>" + (show ? App.translations.FewerDetails : App.translations.MoreDetails);
		});
		this.mTagsContainer = new Ui.Div();
		this.mTagsContainer.addClass("chips");
		tagsSection.appendChild(this.mTagsContainer.container);

		const r = new HttpRequest("PUT", "/v1/reports/basis", { species_uuid: species.uuid });
		r.process(function (status: any, data: any) {
			const d = data as Responses.ReportsBasis;

			if (list) {
				let found = false;
				for (let i = 0; i < list.species.length; ++i) {
					if (list.species[i].uuid === d.species.uuid) {
						found = true;
						break;
					}
				}
				if (!found) {
					return;
				}
			}

			d.ages.sort(Utils.sortByNameOnCurrentLanguage);
			d.ages.forEach(e => {
				ageList.addOption(App.getTranslation(e.name), e.uuid);
			});

			d.genders.sort(Utils.sortByNameOnCurrentLanguage);
			d.genders.forEach(e => {
				genderList.addOption(App.getTranslation(e.name), e.uuid);
			});

			d.methods.sort(Utils.sortByNameOnCurrentLanguage);
			d.methods.forEach(e => {
				methodList.addOption(App.getTranslation(e.name), e.uuid);
			});

			d.activities.sort(Utils.sortByNameOnCurrentLanguage);
			d.activities.forEach(e => {
				activityList.addOption(App.getTranslation(e.name), e.uuid);
			});

			d.tags.forEach(e => {
				self.addTag(e);
			});

			self.addNewTagTag();
		});

		this.mReportButton = sheet.addAction(Ui.actionButton(App.translations.SubmitSighting, "check", "success"));
		this.mReportButton.disabled = true;
		this.mReportButton.addEventListener("click", function () {
			const r = new Requests.Reports();
			r.species_uuid = self.mSpeciesList.data.uuid;
			if (from.value !== "") {
				r.from = Utils.toISODate(new Date(from.value));
			}
			if (until.value !== "") {
				r.until = Utils.toISODate(new Date(until.value));
			}

			if (self.mLocation !== null) {
				r.latitude = self.mLocation.lat;
				r.longitude = self.mLocation.lng;
			}
			r.quantity = parseInt(quantity.value);
			if (ageList.value) {
				r.age_uuid = ageList.value;
			}
			if (activityList.value) {
				r.activity_uuid = activityList.value;
			}
			if (genderList.value) {
				r.gender_uuid = genderList.value;
			}
			if (methodList.value) {
				r.method_uuid = methodList.value;
			}
			r.site = site.value;
			r.comment = privateComment.value;
			r.silent = true;
			const checklists = self.mChecklistContainer.container.getElementsByTagName("input");
			r.list_uuids = Array();
			for (let i = 0; i < checklists.length; ++i) {
				if (checklists[i].checked) {
					r.list_uuids.push(checklists[i].name);
				}
			}

			const tags = self.mTagsContainer.container.getElementsByClassName("tag");
			r.tag_uuids = Array();
			for (let i = 0; i < tags.length; ++i) {
				if (tags[i].classList.contains("selected")) {
					r.tag_uuids.push(tags[i].getAttribute("uuid")!);
				}
			}

			// TODO: Tags and notifications.

			// Disabled while saving so a double tap can't create two sightings.
			Ui.setBusy(self.mReportButton, true);
			const rr = new HttpRequest("POST", "/v1/reports", r);
			rr.process(function (status: any, data: any) {
				new Ui.Toast(App.translations.Congratulations, App.translations.YouFoundA.replace("{0}", Utils.escapeHtml(self.mSpeciesList.value)), null, "var(--success-color)", function () { });
				speciesAdded = true;
				self.mPopUp.close();
				// Every page that shows sightings or progress updates from this, wherever the form was
				// opened from.
				window.dispatchEvent(new Event("sightingadded"));
			}, function () {
				Ui.setBusy(self.mReportButton, false);
				new Ui.MessageBox(App.translations.Error, App.translations.SomethingWentWrong, "/img/messagebox-256.webp", "var(--error-color)", function () { });
			}, function () {
				// Also after network errors, which skip the handler above.
				Ui.setBusy(self.mReportButton, false);
			});
		});
		sheet.open(function () {
			closed = true;
			if (map) {
				map.stopLocate();
				map.remove();
			}
			closeCallback(speciesAdded);
		});
	}

	private addNewTagTag(): void {
		const self = this;
		const newTag = document.createElement("div");
		newTag.id = "add-tag";
		newTag.className = "tag";
		newTag.addEventListener("click", () => {
			new NewTagPopUp((tag: Models.Tag | null) => {
				if (tag) {
					document.getElementById("add-tag")?.remove();
					self.addTag(tag).click();
					self.addNewTagTag();
				}
			});
		});

		const c = document.createElement("p");
		c.innerHTML = "+";
		newTag.appendChild(c);
		this.mTagsContainer.container.appendChild(newTag);
	}

	private addTag(e: Models.Tag): HTMLDivElement {
		const tag = document.createElement("div");
		tag.className = "tag";
		tag.setAttribute("uuid", e.uuid);
		tag.addEventListener("click", () => {
			if (tag.classList.contains("selected")) {
				tag.classList.remove("selected");
			} else {
				tag.classList.add("selected");
			}
		});
		if (e.color !== "") {
			tag.style.setProperty("--tag", e.color);
		}
		const c = document.createElement("p");
		c.textContent = e.name;
		tag.appendChild(c);
		this.mTagsContainer.container.appendChild(tag);

		return tag;
	}

	private validate(): void {
		this.mReportButton.disabled = false;
	}
}

function openReportPopUp(species: Models.Species, list: Models.List | null, closeCallback: (speciesAdded: boolean) => void = function () { }): void {
	new ReportPopUp(species, list, closeCallback);
}

// The "Report sighting" button in the menu: pick the bird, then the report form.
function openQuickReport(): void {
	const sheet = new Ui.Sheet({ eyebrow: App.translations.SubmitSighting, title: App.translations.WhichBird, size: "tall" });
	const section = sheet.section();
	const search = new Ui.SpeciesSearch(App.translations.SearchSpecies, function (species: Models.Species) {
		sheet.close();
		Ui.PopUp.whenHistorySettled(function () {
			openReportPopUp(species, null);
		});
	}, true, App.translations.WhichBirdHint);
	section.appendChild(search.container);

	// Not sure what it was: let the bird's song tell.
	const identify = document.createElement("button");
	identify.className = "secondary identify-entry";
	identify.innerHTML = "<span class='material-symbols-outlined'>mic</span>" + App.translations.IdentifyBySound;
	identify.addEventListener("click", function () {
		sheet.close();
		Ui.PopUp.whenHistorySettled(function () {
			openIdentifyBySound(function (species: Models.Species) {
				openReportPopUp(species, null);
			});
		});
	});
	section.insertBefore(identify, search.container.nextSibling);

	SpeciesCatalog.get(function (c) {
		search.setSpecies(c.species, c.preferred, c.swedish);
	});
	sheet.open();
	// After the sheet has slid in, so the keyboard does not push it around on phones.
	setTimeout(function () {
		search.focus();
	}, 350);
}
