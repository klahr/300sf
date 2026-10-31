/// <reference path="Maps.ts"/>
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
/// <reference path="Outbox.ts"/>

class ReportPopUp {
	private mReportButton: HTMLButtonElement;
	private mSpeciesList: Ui.Input;
	private mChecklistContainer: Ui.Div;
	private mTagsContainer: Ui.Div;
	// Null until the user taps the map or their position is found.
	private mLocation: { lat: number, lng: number } | null = null;
	private mPopUp: Ui.Sheet;

	// With existing, the form edits that sighting instead of reporting a new one. seenAt is when a new
	// sighting was made, if not now.
	constructor(species: Models.Species, list: Models.List | null, closeCallback: (speciesAdded: boolean) => void, existing: Models.Report | null = null, seenAt: Date | null = null) {
		const self = this;

		var speciesAdded = false;

		const sheet = new Ui.Sheet({ eyebrow: existing ? App.translations.EditSighting : App.translations.SubmitSighting, title: App.getTranslation(species.name), subtitle: species.name.la, latin: true });
		this.mPopUp = sheet;

		// A wrong bird is the usual mistake; pick the right one and carry on editing.
		if (existing) {
			const changeSpecies = document.createElement("button");
			changeSpecies.className = "text-button";
			changeSpecies.innerHTML = "<span class='material-symbols-outlined'>swap_horiz</span>" + App.translations.ChangeSpecies;
			changeSpecies.addEventListener("click", function () {
				sheet.close();
				Ui.PopUp.whenHistorySettled(function () {
					openSpeciesPicker(function (other: Models.Species) {
						openReportPopUp(other, null, closeCallback, existing);
					}, App.translations.EditSighting);
				});
			});
			sheet.names.appendChild(changeSpecies);
		}

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
		where.appendChild(positionNote);
		const showNote = function (text: string | null) {
			positionNote.textContent = text || "";
			positionNote.classList.toggle("hidden", text === null);
		};

		setTimeout(function () {
			if (closed) {
				return;
			}
			// Where the user last was, as a better first guess than the middle of Sweden.
			const ks = new LocalStorage();
			const last = ks.get("last-position") as any;
			map = L.map(mapDiv.container).setView(last ? [last.lat, last.lng] : [62.0, 16.0], last ? 13 : 5);
			Maps.tileLayer().addTo(map);
			const markerGroup = L.layerGroup().addTo(map);
			if (self.mLocation !== null) {
				map.setView([self.mLocation.lat, self.mLocation.lng], 13);
				L.marker([self.mLocation.lat, self.mLocation.lng]).addTo(markerGroup);
			}

			// The position keeps improving while the GPS settles; stop once it is good enough, or
			// when the user puts the marker where they want it.
			const goodEnoughMeters = 25;
			let watching = false;
			let fixes = 0;
			let watchTimer = 0;
			const stopWatching = function () {
				if (watching) {
					map.stopLocate();
					watching = false;
				}
				clearTimeout(watchTimer);
				Ui.setBusy(locateButton, false);
			};
			const offlineNote = function (): string | null {
				return navigator.onLine ? null : App.translations.MapOffline;
			};

			map.on("click", function (e) {
				stopWatching();
				markerGroup.clearLayers();
				L.marker(e.latlng).addTo(markerGroup);
				self.mLocation = e.latlng;
				showNote(offlineNote());
			});
			map.on('locationfound', function (e) {
				if (!watching) {
					return;
				}
				const first = fixes++ === 0;
				markerGroup.clearLayers();
				L.circle(e.latlng, { radius: e.accuracy, weight: 1, fillOpacity: 0.1 }).addTo(markerGroup);
				L.marker(e.latlng).addTo(markerGroup);
				self.mLocation = e.latlng;
				ks.set("last-position", { lat: e.latlng.lat, lng: e.latlng.lng }, 0);
				map.setView(e.latlng, first ? 16 : map.getZoom());
				const accuracy = App.translations.PositionAccuracy.replace("{0}", String(Math.round(e.accuracy)));
				const offline = offlineNote();
				showNote(offline ? accuracy + " " + offline : accuracy);
				if (e.accuracy <= goodEnoughMeters) {
					stopWatching();
				}
			});
			// Not an error worth a dialog: the position is optional and can be set on the map.
			// A later fix may still come once there is one; without, give up.
			map.on('locationerror', function () {
				if (!watching || fixes > 0) {
					return;
				}
				stopWatching();
				if (self.mLocation === null) {
					showNote(App.translations.NoPositionYet);
				}
			});

			locateButton.addEventListener("click", function () {
				stopWatching();
				fixes = 0;
				watching = true;
				Ui.setBusy(locateButton, true);
				map.locate({ watch: true, enableHighAccuracy: true, timeout: 30000, maximumAge: 10000 });
				// Whatever it has by then will do.
				watchTimer = setTimeout(stopWatching, 60000);
			});
			// An edited sighting keeps its place unless the user asks for where they are now.
			if (!existing) {
				setTimeout(function() {
					if (!closed) {
						locateButton.click();
					}
				}, 1000);
			}
		}, 10);

		if (existing && existing.latitude !== 0 && existing.longitude !== 0) {
			this.mLocation = { lat: existing.latitude, lng: existing.longitude };
		}

		// The species is fixed; the input only carries it to the request and the toast.
		this.mSpeciesList = new Ui.Input("text");
		this.mSpeciesList.value = App.getTranslation(species.name);
		this.mSpeciesList.setData(species);

		const showLists = function (d: Responses.ListsSpecies) {
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

				const onIt = existing !== null && (existing.lists || []).some(l => l.uuid === e.list.uuid);
				if (onIt) {
					// Left free even when the bird isn't on the list, so it can be taken off.
					(c.container as HTMLInputElement).checked = true;
				} else if (!e.available) {
					c.container.setAttribute("disabled", "");
					dd.addClass("unavailable");
				} else if (list && e.list.uuid === list.uuid) {
					(c.container as HTMLInputElement).checked = true;
					(c.container as HTMLInputElement).disabled = true;
				}
			});

			// With only one list that can hold the species, choose it for the user.
			const open = Array.from(self.mChecklistContainer.container.querySelectorAll("input")).filter(c => !c.disabled);
			if (!existing && open.length === 1 && !self.mChecklistContainer.container.querySelector("input:checked")) {
				open[0].checked = true;
			}

			self.validate();
		};
		const loadLists = function () {
			const r = new HttpRequest("GET", "/v1/lists/species/" + species.uuid);
			r.setQuiet();
			r.process(function (status: any, data: any) {
				showLists(data as Responses.ListsSpecies);
			}, function (status: number) {
				if (!Outbox.isTemporary(status)) {
					return;
				}
				// Offline, and this bird's lists were never fetched: work them out from the user's
				// lists, if the app kept those.
				const all = new HttpRequest("GET", "/v1/lists/list");
				all.setQuiet();
				all.process(function (status: any, data: any) {
					const d = new Responses.ListsSpecies();
					d.result = (data.result || []).map(function (l: Models.List) {
						const a = new Responses.ListAvailable();
						a.list = l;
						a.available = (l.species || []).some(s => s.uuid === species.uuid);
						return a;
					});
					showLists(d);
				}, function () {
					self.mChecklistContainer.container.innerHTML = "";
					const note = document.createElement("p");
					note.className = "details";
					note.textContent = App.translations.ListsUnavailableOffline;
					self.mChecklistContainer.container.appendChild(note);
					self.validate();
				});
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

		// For the date inputs, in local time.
		const toLocalInput = function (d: Date): string {
			d.setMinutes(d.getMinutes() - d.getTimezoneOffset());
			return d.toISOString().slice(0, -8);
		};
		const now = seenAt ? new Date(seenAt.getTime()) : new Date();
		now.setSeconds(now.getMinutes());
		from.value = toLocalInput(now);

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

		if (existing) {
			from.value = Utils.isValidDate(existing.from as any) ? toLocalInput(new Date(existing.from as any)) : "";
			if (Utils.isValidDate(existing.until as any)) {
				until.value = toLocalInput(new Date(existing.until as any));
			}
			quantity.value = existing.quantity ? existing.quantity.toString() : "";
			site.value = existing.site || "";
			privateComment.value = existing.comment ? existing.comment.text : "";
		}

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

		const fillBasis = function (d: Responses.ReportsBasis) {

			if (existing) {
				// An edit can take a detail away again.
				[ageList, genderList, methodList, activityList].forEach(l => l.addOption(App.translations.None, ""));
			}

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
				const tag = self.addTag(e);
				if (existing && (existing.tags || []).some(t => t.uuid === e.uuid)) {
					tag.classList.add("selected");
				}
			});

			self.addNewTagTag();

			if (existing) {
				const chosen: Array<[Ui.Select, { uuid: string } | null]> = [[ageList, existing.age], [genderList, existing.gender], [methodList, existing.method], [activityList, existing.activity]];
				chosen.forEach(([l, v]) => {
					if (v) {
						l.setSelected(v.uuid);
					}
				});
				// Show the details the sighting has.
				const hasDetails = chosen.some(([, v]) => !!v) || site.value !== "" || privateComment.value !== "" || (existing.tags || []).length > 0;
				if (hasDetails && details.classList.contains("hidden")) {
					moreToggle.click();
				}
			}
		};
		// The choices are the same for every bird, so the last ones fetched do offline.
		const basis = new HttpRequest("PUT", "/v1/reports/basis", { species_uuid: species.uuid });
		basis.setQuiet();
		basis.process(function (status: any, data: any) {
			const d = data as Responses.ReportsBasis;
			new LocalStorage().set("report-basis", { ages: d.ages, genders: d.genders, methods: d.methods, activities: d.activities, tags: d.tags }, 0);
			fillBasis(d);
		}, function (status: number) {
			const kept = new LocalStorage().get("report-basis") as any;
			if (kept && Outbox.isTemporary(status)) {
				kept.species = species;
				fillBasis(kept as Responses.ReportsBasis);
			}
		});

		this.mReportButton = sheet.addAction(Ui.actionButton(existing ? App.translations.Save : App.translations.SubmitSighting, "check", "success"));
		this.mReportButton.disabled = true;
		this.mReportButton.addEventListener("click", function () {
			const r = new Requests.Reports();
			r.uuid = Outbox.newUuid();
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

			if (existing) {
				self.save(existing, r);
				return;
			}

			// No connection: keep it on this device and send it later.
			const keep = function () {
				Outbox.add(r, species.name);
				new Ui.Toast(App.translations.SavedOffline, App.translations.SavedOfflineDescription.replace("{0}", Utils.escapeHtml(self.mSpeciesList.value)), null, "var(--success-color)", function () { });
				self.mPopUp.close();
			};
			if (!navigator.onLine && App.user) {
				keep();
				return;
			}

			// Disabled while saving so a double tap can't create two sightings.
			Ui.setBusy(self.mReportButton, true);
			const rr = new HttpRequest("POST", "/v1/reports", r);
			rr.setQuiet();
			// A weak signal in the field shouldn't keep the user waiting a minute.
			rr.setTimeLimit(20000);
			rr.process(function (status: any, data: any) {
				new Ui.Toast(App.translations.Congratulations, App.translations.YouFoundA.replace("{0}", Utils.escapeHtml(self.mSpeciesList.value)), null, "var(--success-color)", function () { });
				speciesAdded = true;
				self.mPopUp.close();
				// Every page that shows sightings or progress updates from this, wherever the form was
				// opened from.
				window.dispatchEvent(new Event("sightingadded"));
			}, function (status: number) {
				Ui.setBusy(self.mReportButton, false);
				if (Outbox.isTemporary(status) && App.user) {
					keep();
					return;
				}
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

	// Sends an edit: every field is replaced, so the request carries all of them.
	private save(existing: Models.Report, r: Requests.Reports): void {
		const self = this;
		const body: any = Object.assign({}, r);
		delete body.uuid;
		delete body.silent;
		["age_uuid", "activity_uuid", "gender_uuid", "method_uuid"].forEach(k => {
			if (!body[k]) {
				delete body[k];
			}
		});
		Ui.setBusy(self.mReportButton, true);
		new HttpRequest("PUT", "/v1/reports/" + encodeURIComponent(existing.uuid), body).process(function () {
			new Ui.Toast(App.translations.Saved, "", null, "var(--success-color)", function () { });
			self.mPopUp.close();
			window.dispatchEvent(new Event("sightingadded"));
		}, function () {
			new Ui.MessageBox(App.translations.Error, App.translations.SomethingWentWrong, "/img/messagebox-256.webp", "var(--error-color)", function () { });
		}, function () {
			Ui.setBusy(self.mReportButton, false);
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

function openReportPopUp(species: Models.Species, list: Models.List | null, closeCallback: (speciesAdded: boolean) => void = function () { }, existing: Models.Report | null = null, seenAt: Date | null = null): void {
	new ReportPopUp(species, list, closeCallback, existing, seenAt);
}

// The "Report sighting" button in the menu: pick the bird, then the report form.
function openQuickReport(): void {
	openSpeciesPicker(function (species: Models.Species) {
		openReportPopUp(species, null);
	});
}

// Asks which bird, by name or by its song.
function openSpeciesPicker(chosen: (species: Models.Species) => void, eyebrow: string = App.translations.SubmitSighting): void {
	const sheet = new Ui.Sheet({ eyebrow: eyebrow, title: App.translations.WhichBird, size: "tall" });
	const section = sheet.section();
	const search = new Ui.SpeciesSearch(App.translations.SearchSpecies, function (species: Models.Species) {
		sheet.close();
		Ui.PopUp.whenHistorySettled(function () {
			chosen(species);
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
			openIdentifyBySound(chosen);
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
