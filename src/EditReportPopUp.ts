/// <reference path="responses/Reports.ts"/>
/// <reference path="responses/Lists.ts"/>
/// <reference path="requests/Reports.ts"/>
/// <reference path="models/Species.ts"/>
/// <reference path="models/List.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="utils/Sort.ts"/>

//declare const L;

class EditReportPopUp {
	private mReportButton: HTMLButtonElement;
	private mChecklistContainer: Ui.Div;d
	private mPopUp: Ui.PopUp;

	constructor(report: Models.Report, closeCallback: () => void) {
		const self = this;

		const content = new Ui.Div();
		content.style.textAlign = "left";
		content.style.width = "100%";
		content.style.position = "relative";
	
		this.mChecklistContainer = new Ui.Div();
		this.mChecklistContainer.style.position = "relative";
	
		const name = document.createElement("h1");
		name.innerHTML = App.translations.Sighting;
		name.style.margin = "unset";
		content.container.appendChild(name);
	
		if (report.latitude !== 0 && report.longitude !== 0) {
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

			setTimeout(function() {
				const latLng = [report.latitude, report.longitude];
				const map = L.map('map').setView(latLng, 13);
				L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
					maxZoom: 19,
					attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
				}).addTo(map);
				const markerGroup = L.layerGroup().addTo(map);
				L.marker(latLng).addTo(markerGroup);
			}, 10);
		}

		const speciesList = new Ui.Input("text")
		speciesList.value = App.getTranslation(report.species.name);
		speciesList.setDisabled(true);
		content.attach(speciesList);

		if (report.age) {
			const ageList = new Ui.Input("text");
			ageList.value = App.getTranslation(report.age.name);
			ageList.setDisabled(true);
			content.attach(ageList);	
		}

		if (report.gender) {
			const genderList = new Ui.Input("text");
			genderList.value = App.getTranslation(report.gender.name);
			genderList.setDisabled(true);
			content.attach(genderList);	
		}

		if (report.method) {
			const methodList = new Ui.Input("text");
			methodList.value = App.getTranslation(report.method.name);
			methodList.setDisabled(true);
			content.attach(methodList);	
		}

		if (report.activity) {
			const activityList = new Ui.Input("text");
			activityList.value = App.getTranslation(report.activity.name);
			activityList.setDisabled(true);
			content.attach(activityList);	
		}
	
		const quantityLbl = document.createElement("h3");
		quantityLbl.className = "title";
		quantityLbl.innerHTML = App.translations.Quantity;
		content.container.appendChild(quantityLbl);
		const quantity = new Ui.Input("number");
		quantity.setDisabled(true);
		quantity.value = report.quantity.toString();
		content.attach(quantity);

		if (Utils.isValidDate(report.from)) {
			const fromLbl = document.createElement("h3");
			fromLbl.className = "title";
			fromLbl.innerHTML = App.translations.ObservedFrom;
			content.container.appendChild(fromLbl);
			const from = new Ui.Input("text");
			from.value = Utils.toYYYY_MM_DD_hh_mm(report.from);
			from.setDisabled(true);
			content.attach(from);

			if (Utils.isValidDate(report.until)) {
				const untilLbl = document.createElement("h3");
				untilLbl.className = "title";
				untilLbl.innerHTML = App.translations.ObservedUntil;
				content.container.appendChild(untilLbl);
				const until = new Ui.Input("text");
				until.value = Utils.toYYYY_MM_DD_hh_mm(report.until);
				until.setDisabled(true);
				content.attach(until);
			}
		}

		if (report.comment.text !== "") {
			const privateCommentLbl = document.createElement("h3");
			privateCommentLbl.className = "title";
			privateCommentLbl.innerHTML = App.translations.PrivateComment;
			content.container.appendChild(privateCommentLbl);
			const privateComment = document.createElement("textarea");
			privateComment.style.height = "128px";
			privateComment.disabled = true;
			privateComment.innerHTML = report.comment.text;
			content.container.appendChild(privateComment);	
		}
	
		const listsLbl = document.createElement("h3");
		listsLbl.className = "title";
		listsLbl.innerHTML = App.translations.Checklists;
		content.container.appendChild(listsLbl);
		content.attach(self.mChecklistContainer);

		report.lists.forEach(e => {
			const dd = new Ui.Div();
			dd.style.position = "relative";
			dd.style.width = "100%";
			self.mChecklistContainer.attach(dd);
			const c = new Ui.Input("checkbox");
			c.container.setAttribute("disabled", "");
			(c.container as HTMLInputElement).checked = true;

			dd.attach(c);
			const cLbl = document.createElement("label");
			cLbl.setAttribute("for", c.container.getAttribute("name")!);
			cLbl.innerHTML = App.getTranslation(e.name);
			dd.container.appendChild(cLbl);
		});

		if (report.tags.length > 0) {
			const tagsLbl = document.createElement("h3");
			tagsLbl.className = "title";
			tagsLbl.innerHTML = App.translations.Tags;
			content.container.appendChild(tagsLbl);
	
			const tagsContainer = new Ui.Div();
			tagsContainer.style.position = "relative";
			content.attach(tagsContainer);
	
			report.tags.forEach(e => {
				const tag = document.createElement("div");
				tag.className = "tag";
				if (e.color !== "") {
					tag.style.backgroundColor = e.color;
				}
				const c = document.createElement("p");
				c.innerHTML = e.name;
				tag.appendChild(c);
				tagsContainer.container.appendChild(tag);
			});
		}

		this.mReportButton = document.createElement("button");
		this.mReportButton.innerHTML = App.translations.ArchiveSighting;
		this.mReportButton.style.width = "auto";
		this.mReportButton.style.cssFloat = "right";
		this.mReportButton.style.borderRadius = "8px";
		this.mReportButton.style.backgroundColor = "#a11f1f";
		this.mReportButton.addEventListener("click", function() {
			new Ui.ConfirmBox(App.translations.ArchiveSighting, App.translations.ConfirmArchiveSighting, null, "var(--success-color)", function() {
				const r = new HttpRequest("POST", "/v1/reports/" + report.uuid + "/archive");
				r.process(function(status: any, data: any) {
					self.mPopUp.close();
					setTimeout(() => {
						location.reload();
					}, 150);
				});
			}, function() {
			}, App.translations.ArchiveSighting, App.translations.Cancel);
		});
		content.container.appendChild(this.mReportButton);

		this.mPopUp = new Ui.PopUp(content.container, function() {
			closeCallback();
		});
	}
}

function openEditReportPopUp(report: Models.Report, closeCallback: () => void = function() {}): void {
	new EditReportPopUp(report, closeCallback);
}
