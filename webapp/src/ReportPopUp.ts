/// <reference path="responses/Reports.ts"/>
/// <reference path="responses/Lists.ts"/>
/// <reference path="requests/Reports.ts"/>
/// <reference path="models/Species.ts"/>
/// <reference path="models/List.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="ui/Toast.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="utils/Sort.ts"/>
/// <reference path="NewTagPopUp.ts"/>

declare const L;

class ReportPopUp {
	private mReportButton: HTMLButtonElement;
	private mSpeciesList: Ui.Input;
	private mChecklistContainer: Ui.Div;
	private mTagsContainer: Ui.Div;
	// Null until the user taps the map or their position is found.
	private mLocation: { lat: number, lng: number } | null = null;
	private mPopUp: Ui.PopUp;

	constructor(species: Models.Species, list: Models.List | null, closeCallback: (speciesAdded: boolean) => void) {
		const self = this;

		var speciesAdded = false;
		const content = new Ui.Div();
		content.style.textAlign = "left";
		content.style.width = "100%";
		content.style.position = "relative";

		this.mChecklistContainer = new Ui.Div();
		this.mChecklistContainer.style.position = "relative";

		const name = document.createElement("h1");
		name.innerHTML = App.translations.SubmitSighting;
		name.style.margin = "unset";
		content.container.appendChild(name);

		const locationLbl = document.createElement("h3");
		locationLbl.className = "title";
		locationLbl.innerHTML = App.translations.ObservationLocation;
		content.container.appendChild(locationLbl);
		const mapDiv = new Ui.Div();
		mapDiv.container.classList.add("map");
		mapDiv.style.height = "300px";
		mapDiv.style.width = "100%";
		mapDiv.style.position = "relative";
		mapDiv.style.borderRadius = "var(--small-distance)";
		content.attach(mapDiv);

		const locateButton = document.createElement("button");
		locateButton.innerHTML = App.translations.FindMyLocation;
		locateButton.style.borderRadius = "8px";
		locateButton.style.backgroundColor = "var(--success-color)";
		content.container.appendChild(locateButton);

		setTimeout(function () {
			const map = L.map(mapDiv.container).setView([58.1801188016368, 13.959220108737176], 13);
			L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
				maxZoom: 19,
				attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
			}).addTo(map);
			const markerGroup = L.layerGroup().addTo(map);
			map.on("click", function (e) {
				markerGroup.clearLayers();
				L.marker(e.latlng).addTo(markerGroup);
				self.mLocation = e.latlng;
			});
			map.on('locationfound', function (e) {
				markerGroup.clearLayers();
				L.marker(e.latlng).addTo(markerGroup);
				self.mLocation = e.latlng;
				map.setZoom(16);
				locateButton.disabled = false;
				locateButton.innerHTML = App.translations.FindMyLocation;
			});
			map.on('locationerror', function () {
				locateButton.disabled = false;
				locateButton.innerHTML = App.translations.FindMyLocation;
				new Ui.MessageBox(App.translations.Error, App.translations.YourLocationNotFound, "/img/messagebox-256.webp", "var(--error-color)", function () { });
			});

			locateButton.addEventListener("click", function () {
				locateButton.innerHTML = App.translations.PleaseWait + "...";
				locateButton.disabled = true;
				map.locate({ setView: true, maxZoom: 19 });
			});
			setTimeout(function() {
				locateButton.click();
			}, 1000);
		}, 10);

		this.mSpeciesList = new Ui.Input("text");
		this.mSpeciesList.setDisabled(true);
		this.mSpeciesList.value = App.getTranslation(species.name);
		this.mSpeciesList.setData(species);
		content.attach(this.mSpeciesList);

		new HttpRequest("GET", "/v1/lists/species/" + species.uuid).process(function (status: any, data: any) {
			const d = data as Responses.ListsSpecies;
			self.mChecklistContainer.detachAll();
			self.mChecklistContainer.container.innerHTML = "";

			d.result.forEach(e => {
				const dd = new Ui.Div();
				dd.style.position = "relative";
				dd.style.width = "100%";
				self.mChecklistContainer.attach(dd);
				const c = new Ui.Input("checkbox");
				c.container.setAttribute("name", e.list.uuid);
				c.container.addEventListener("click", function () {
					self.validate();
				});

				dd.attach(c);
				const cLbl = document.createElement("label");
				cLbl.setAttribute("for", c.container.getAttribute("name")!);
				cLbl.textContent = App.getTranslation(e.list.name);
				dd.container.appendChild(cLbl);

				if (!e.available) {
					c.container.setAttribute("disabled", "");
					cLbl.style.color = "var(--secondary-color)";
				} else if (list && e.list.uuid === list.uuid) {
					(c.container as HTMLInputElement).checked = true;
					(c.container as HTMLInputElement).disabled = true;
				}
			});

			self.validate();
		});

		const ageList = new Ui.Select(App.translations.SelectAge + "...", App.translations.Search);
		content.attach(ageList);

		const genderList = new Ui.Select(App.translations.SelectGender + "...", App.translations.Search);
		content.attach(genderList);

		const methodList = new Ui.Select(App.translations.SelectMethod + "...", App.translations.Search);
		content.attach(methodList);

		const activityList = new Ui.Select(App.translations.SelectActivity + "...", App.translations.Search);
		content.attach(activityList);

		const siteLbl = document.createElement("h3");
		siteLbl.className = "title";
		siteLbl.innerHTML = App.translations.Site;
		content.container.appendChild(siteLbl);
		const site = new Ui.Input("text");
		content.attach(site);

		const quantityLbl = document.createElement("h3");
		quantityLbl.className = "title";
		quantityLbl.innerHTML = App.translations.Quantity;
		content.container.appendChild(quantityLbl);
		const quantity = new Ui.Input("number");
		quantity.container.setAttribute("min", "1");
		quantity.container.setAttribute("step", "1");
		quantity.value = "1";
		content.attach(quantity);

		const fromLbl = document.createElement("h3");
		fromLbl.className = "title";
		fromLbl.innerHTML = App.translations.ObservedFrom;
		content.container.appendChild(fromLbl);
		const from = new Ui.Input("datetime-local");
		content.attach(from);

		const now = new Date();
		now.setMinutes(now.getMinutes() - now.getTimezoneOffset());
		now.setSeconds(now.getMinutes());
		from.value = now.toISOString().slice(0, -8);

		const untilLbl = document.createElement("h3");
		untilLbl.className = "title";
		untilLbl.innerHTML = App.translations.ObservedUntil;
		content.container.appendChild(untilLbl);
		const until = new Ui.Input("datetime-local");
		content.attach(until);

		const privateCommentLbl = document.createElement("h3");
		privateCommentLbl.className = "title";
		privateCommentLbl.innerHTML = App.translations.PrivateComment;
		content.container.appendChild(privateCommentLbl);
		const privateComment = document.createElement("textarea");
		privateComment.style.height = "128px";
		content.container.appendChild(privateComment);

		const tagsLbl = document.createElement("h3");
		tagsLbl.className = "title";
		tagsLbl.innerHTML = App.translations.Tags;
		content.container.appendChild(tagsLbl);

		this.mTagsContainer = new Ui.Div();
		this.mTagsContainer.style.position = "relative";
		this.mTagsContainer.style.margin = "var(--small-distance)";
		content.attach(this.mTagsContainer);

		const listsLbl = document.createElement("h3");
		listsLbl.className = "title";
		listsLbl.innerHTML = App.translations.AvailableChecklists;
		content.container.appendChild(listsLbl);
		content.attach(self.mChecklistContainer);

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

		this.mReportButton = document.createElement("button");
		this.mReportButton.innerHTML = App.translations.SubmitSighting;
		this.mReportButton.style.width = "auto";
		this.mReportButton.style.cssFloat = "right";
		this.mReportButton.style.borderRadius = "8px";
		this.mReportButton.style.backgroundColor = "var(--success-color)";
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
			self.mReportButton.disabled = true;
			const rr = new HttpRequest("POST", "/v1/reports", r);
			rr.process(function (status: any, data: any) {
				new Ui.Toast(App.translations.Congratulations, App.translations.YouFoundA.replace("{0}", Utils.escapeHtml(self.mSpeciesList.value)), null, "var(--success-color)", function () { });
				speciesAdded = true;
				self.mPopUp.close();
			}, function () {
				self.mReportButton.disabled = false;
				new Ui.MessageBox(App.translations.Error, App.translations.SomethingWentWrong, "/img/messagebox-256.webp", "var(--error-color)", function () { });
			});
		});
		content.container.appendChild(this.mReportButton);

		this.mPopUp = new Ui.PopUp(content.container, function () {
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
			tag.style.backgroundColor = e.color;
		}
		const c = document.createElement("p");
		c.textContent = e.name;
		tag.appendChild(c);
		this.mTagsContainer.container.appendChild(tag);

		return tag;
	}

	private validate(): void {
		this.mReportButton.disabled = false;

		const checklists = this.mChecklistContainer.container.getElementsByTagName("input");
		let isAnyChecked = false;
		for (let i = 0; i < checklists.length; ++i) {
			if (checklists[i].checked) {
				isAnyChecked = true;
				break;
			}
		}

		if (!isAnyChecked) {
			this.mReportButton.disabled = true;
			return;
		}
	}
}

function openReportPopUp(species: Models.Species, list: Models.List | null, closeCallback: (speciesAdded: boolean) => void = function () { }): void {
	new ReportPopUp(species, list, closeCallback);
}
