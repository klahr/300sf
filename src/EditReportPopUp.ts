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
	private mSpeciesList: Ui.Select;
	private mChecklistContainer: Ui.Div;
	private mLocation = L.latLng;
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
		name.innerHTML = App.translations.Sightings;
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
				console.log(latLng);
				const map = L.map('map').setView(latLng, 13);
				L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
					maxZoom: 19,
					attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
				}).addTo(map);
				const markerGroup = L.layerGroup().addTo(map);
				L.marker(latLng).addTo(markerGroup);
			}, 10);
		}

		this.mSpeciesList = new Ui.Select(App.getTranslation(report.species.name), "");
		this.mSpeciesList.setDisabled(true);
		content.attach(this.mSpeciesList);
	
		const ageList = new Ui.Select(report.age ? App.getTranslation(report.age.name) : "", "");
		ageList.setDisabled(true);
		content.attach(ageList);

		const genderList = new Ui.Select(report.gender ? App.getTranslation(report.gender.name) : "", "");
		genderList.setDisabled(true);
		content.attach(genderList);

		const methodList = new Ui.Select(report.method ? App.getTranslation(report.method.name) : "", "");
		methodList.setDisabled(true);
		content.attach(methodList);

		const activityList = new Ui.Select(report.activity ? App.getTranslation(report.activity.name) : "", "");
		activityList.setDisabled(true);
		content.attach(activityList);
	
		const quantityLbl = document.createElement("h3");
		quantityLbl.className = "title";
		quantityLbl.innerHTML = App.translations.Quantity;
		content.container.appendChild(quantityLbl);
		const quantity = new Ui.Input("number");
		quantity.setDisabled(true);
		quantity.value = report.quantity.toString();
		content.attach(quantity);
	
		const fromLbl = document.createElement("h3");
		fromLbl.className = "title";
		fromLbl.innerHTML = App.translations.ObservedFrom;
		content.container.appendChild(fromLbl);
		const from = new Ui.Input("datetime-local");
		from.setDisabled(true);
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
		until.setDisabled(true);
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
		privateComment.disabled = true;
		privateComment.innerHTML = report.comment.text;
		content.container.appendChild(privateComment);
	
		const listsLbl = document.createElement("h3");
		listsLbl.className = "title";
		listsLbl.innerHTML = App.translations.AvailableChecklists;
		content.container.appendChild(listsLbl);
		content.attach(self.mChecklistContainer);

		this.mReportButton = document.createElement("button");
		this.mReportButton.innerHTML = App.translations.DeleteSighting;
		this.mReportButton.style.width = "auto";
		this.mReportButton.style.cssFloat = "right";
		this.mReportButton.style.borderRadius = "8px";
		this.mReportButton.style.backgroundColor = "#a11f1f";
		this.mReportButton.addEventListener("click", function() {
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
