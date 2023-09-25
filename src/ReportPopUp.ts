/// <reference path="responses/ReportsBasis.ts"/>
/// <reference path="models/Species.ts"/>
/// <reference path="models/List.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="HttpRequest.ts"/>

declare const L;

function openReportPopUp(species: Models.Species | null, list: Models.List | null): void {
	const content = new Ui.Div();
	content.style.textAlign = "left";
    content.style.width = "100%";
	content.style.position = "relative";

	const name = document.createElement("h1");
	name.innerHTML = "Submit";
	name.style.margin = "unset";
	content.container.appendChild(name);

	const locationLbl = document.createElement("h3");
	locationLbl.className = "title";
	locationLbl.innerHTML = "Observation location";
	content.container.appendChild(locationLbl);
	const mapDiv = new Ui.Div();
	mapDiv.container.id = "map";
	mapDiv.style.height = "300px";
	mapDiv.style.width = "100%";
	mapDiv.style.position = "relative";
	mapDiv.style.borderRadius = "var(--small-distance)";
	content.attach(mapDiv);

	const locateButton = document.createElement("button");
	locateButton.innerHTML = "Locate me";
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
    speciesListLbl.innerHTML = "Species";
    content.container.appendChild(speciesListLbl);
	const speciesList = new Ui.Select("Select species...");
	content.attach(speciesList);

	const ageListLbl = document.createElement("h3");
    ageListLbl.className = "title";
    ageListLbl.innerHTML = "Age";
    content.container.appendChild(ageListLbl);
	const ageList = new Ui.Select("Select age...");
	content.attach(ageList);

	const genderListLbl = document.createElement("h3");
    genderListLbl.className = "title";
    genderListLbl.innerHTML = "Gender";
    content.container.appendChild(genderListLbl);
	const genderList = new Ui.Select("Select gender...");
	content.attach(genderList);

	const methodListLbl = document.createElement("h3");
    methodListLbl.className = "title";
    methodListLbl.innerHTML = "Method";
    content.container.appendChild(methodListLbl);
	const methodList = new Ui.Select("Select method...");
	content.attach(methodList);

	const activityListLbl = document.createElement("h3");
    activityListLbl.className = "title";
    activityListLbl.innerHTML = "Activity";
    content.container.appendChild(activityListLbl);
	const activityList = new Ui.Select("Select activity...");
	content.attach(activityList);

	const quantityLbl = document.createElement("h3");
	quantityLbl.className = "title";
	quantityLbl.innerHTML = "Quantity";
	content.container.appendChild(quantityLbl);
	const quantity = new Ui.Input("number");
	quantity.container.setAttribute("min", "1");
	quantity.container.setAttribute("step", "1");
	quantity.value = "1";
	content.attach(quantity);

	const fromLbl = document.createElement("h3");
	fromLbl.className = "title";
	fromLbl.innerHTML = "Observed from";
	content.container.appendChild(fromLbl);
	const from = new Ui.Input("date");
	content.attach(from);

	const untilLbl = document.createElement("h3");
	untilLbl.className = "title";
	untilLbl.innerHTML = "Observed until";
	content.container.appendChild(untilLbl);
	const until = new Ui.Input("date");
	content.attach(until);

	const privateCommentLbl = document.createElement("h3");
	privateCommentLbl.className = "title";
	privateCommentLbl.innerHTML = "Private comment";
	content.container.appendChild(privateCommentLbl);
	const privateComment = document.createElement("textarea");
	privateComment.style.width = "100%";
	privateComment.style.height = "128px";
	privateComment.style.resize = "none";
	privateComment.style.padding = "var(--small-distance)";
	privateComment.style.boxSizing = "border-box";
	content.container.appendChild(privateComment);

    const r = new HttpRequest("GET", "/v1/reports/basis");
	r.process(function(status: any, data: any) {
        const d = data as Responses.ReportsBasis;
        d.species.forEach(e => {
            speciesList.addOption(App.getTranslation(e.name), e.uuid);
        });
		d.ages.forEach(e => {
			ageList.addOption(App.getTranslation(e.name), e.uuid);
		});
		d.genders.forEach(e => {
			genderList.addOption(App.getTranslation(e.name), e.uuid);
		});
		d.methods.forEach(e => {
			methodList.addOption(App.getTranslation(e.name), e.uuid);
		});
		d.activities.forEach(e => {
			activityList.addOption(App.getTranslation(e.name), e.uuid);
		});
    });

	const reportButton = document.createElement("button");
	reportButton.innerHTML = "Submit";
	reportButton.style.width = "auto";
	reportButton.style.cssFloat = "right";
	reportButton.style.borderRadius = "8px";
	reportButton.style.backgroundColor = "var(--success-color)";
	reportButton.addEventListener("click", function() {
		console.log("Submit report");
	});
	content.container.appendChild(reportButton);

	new Ui.PopUp(content.container, function() {
		console.log("closed");
	});
}
