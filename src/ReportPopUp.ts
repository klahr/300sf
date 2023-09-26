/// <reference path="responses/Reports.ts"/>
/// <reference path="responses/Lists.ts"/>
/// <reference path="models/Species.ts"/>
/// <reference path="models/List.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="utils/sort.ts"/>

declare const L;

class ReportPopup {
	private mReportButton: HTMLButtonElement;
	private mSpeciesList: Ui.Select;
	private mChecklistContainer: Ui.Div;

	constructor(species: Models.Species | null, list: Models.List | null) {
		const self = this;

		const content = new Ui.Div();
		content.style.textAlign = "left";
		content.style.width = "100%";
		content.style.position = "relative";
	
		this.mChecklistContainer = new Ui.Div();
		this.mChecklistContainer.style.position = "relative";
	
		const name = document.createElement("h1");
		name.innerHTML = App.translations.Submit;
		name.style.margin = "unset";
		content.container.appendChild(name);
	
		const locationLbl = document.createElement("h3");
		locationLbl.className = "title";
		locationLbl.innerHTML = App.translations.ObservationLocation;
		content.container.appendChild(locationLbl);
		const mapDiv = new Ui.Div();
		mapDiv.container.id = "map";
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
	
		setTimeout(function() {
			const map = L.map('map').setView([58.1801188016368, 13.959220108737176], 13);
			L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
				maxZoom: 19,
				attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
			}).addTo(map);
			map.locate({setView: true, maxZoom: 19});
			const markerGroup = L.layerGroup().addTo(map);
			map.on("click", function(e) {
				markerGroup.clearLayers();
				L.marker(e.latlng).addTo(markerGroup);
			});
			map.on('locationfound', function(e) {
				markerGroup.clearLayers();
				L.marker(e.latlng).addTo(markerGroup);
				map.setZoom(16);
			});
	
			locateButton.addEventListener("click", function() {
				map.locate({setView: true, maxZoom: 19});
			});
		}, 10);

		const speciesListLbl = document.createElement("h3");
		speciesListLbl.className = "title";
		speciesListLbl.innerHTML = App.translations.Species;
		content.container.appendChild(speciesListLbl);
		this.mSpeciesList = new Ui.Select(App.translations.SelectSpecies + "...", function(species: string | null) {
			const r = new HttpRequest("GET", "/v1/lists/species/" + species);
			r.process(function(status: any, data: any) {
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
					c.container.addEventListener("click", function() {
						self.validate();
					});
					if (!e.available) {
						c.container.setAttribute("disabled", "");
					}
					dd.attach(c);
					const cLbl = document.createElement("label");
					cLbl.setAttribute("for", c.container.getAttribute("name")!);
					cLbl.innerHTML = App.getTranslation(e.list.name);
					dd.container.appendChild(cLbl);
				});

				self.validate();
			});
		});
		content.attach(this.mSpeciesList);
	
		const ageListLbl = document.createElement("h3");
		ageListLbl.className = "title";
		ageListLbl.innerHTML = App.translations.Age;
		content.container.appendChild(ageListLbl);
		const ageList = new Ui.Select(App.translations.SelectAge + "...");
		content.attach(ageList);
	
		const genderListLbl = document.createElement("h3");
		genderListLbl.className = "title";
		genderListLbl.innerHTML = App.translations.Gender;
		content.container.appendChild(genderListLbl);
		const genderList = new Ui.Select(App.translations.SelectGender + "...");
		content.attach(genderList);
	
		const methodListLbl = document.createElement("h3");
		methodListLbl.className = "title";
		methodListLbl.innerHTML = App.translations.Method;
		content.container.appendChild(methodListLbl);
		const methodList = new Ui.Select(App.translations.SelectMethod + "...");
		content.attach(methodList);
	
		const activityListLbl = document.createElement("h3");
		activityListLbl.className = "title";
		activityListLbl.innerHTML = "Activity";
		content.container.appendChild(activityListLbl);
		const activityList = new Ui.Select(App.translations.SelectActivity + "...");
		content.attach(activityList);
	
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
		console.log(now.toISOString());
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
		privateComment.style.width = "100%";
		privateComment.style.height = "128px";
		privateComment.style.resize = "none";
		privateComment.style.padding = "var(--small-distance)";
		privateComment.style.boxSizing = "border-box";
		content.container.appendChild(privateComment);
	
		const listsLbl = document.createElement("h3");
		listsLbl.className = "title";
		listsLbl.innerHTML = "Available checklists";
		content.container.appendChild(listsLbl);
		content.attach(self.mChecklistContainer);
	
		const r = new HttpRequest("GET", "/v1/reports/basis");
		r.process(function(status: any, data: any) {
			const d = data as Responses.ReportsBasis;
	
			d.species.sort(Utils.sortByNameOnCurrentLanguage);
			d.species.forEach(e => {
				self.mSpeciesList.addOption(App.getTranslation(e.name), e.uuid);
			});
	
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
		});
	
		this.mReportButton = document.createElement("button");
		this.mReportButton.innerHTML = App.translations.Submit;
		this.mReportButton.style.width = "auto";
		this.mReportButton.style.cssFloat = "right";
		this.mReportButton.style.borderRadius = "8px";
		this.mReportButton.style.backgroundColor = "var(--success-color)";
		this.mReportButton.disabled = true;
		this.mReportButton.addEventListener("click", function() {
			console.log("Submit report");
		});
		content.container.appendChild(this.mReportButton);

		new Ui.PopUp(content.container, function() {
			console.log("closed");
		});
	}

	private validate(): void {
		this.mReportButton.disabled = false;
		if (this.mSpeciesList.value === "") {
			this.mReportButton.disabled = true;
			return;
		}

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

function openReportPopUp(species: Models.Species | null, list: Models.List | null): void {
	new ReportPopup(species, list);
}
