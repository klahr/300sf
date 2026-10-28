/// <reference path="ui/Div.ts"/>
/// <reference path="models/Recording.ts"/>
/// <reference path="App.ts"/>

function openRecordingPopUp(species: Models.Species, recording: Models.Recording) {
	const content = new Ui.Div();
	content.style.textAlign = "left";
	content.style.position = "relative";
	content.style.width = "100%";

	const name = document.createElement("h1");
	name.innerHTML = App.getTranslation(species.name);
	name.style.margin = "unset";
	content.container.appendChild(name);

	const subName = document.createElement("p");
	subName.innerHTML = recording.type.charAt(0).toUpperCase() + recording.type.slice(1);
	subName.style.fontStyle = "italic";
	content.container.appendChild(subName);

	const mapDiv = new Ui.Div();
	mapDiv.container.id = "map";
	mapDiv.style.height = "300px";
	mapDiv.style.width = "100%";
	mapDiv.style.position = "relative";
	mapDiv.style.borderRadius = "var(--small-distance)";
	content.attach(mapDiv);

	const place = document.createElement("p");
	place.innerHTML = recording.location + ", " + recording.country;
	place.style.fontStyle = "italic";
	content.container.appendChild(place);

	setTimeout(function () {
		const latLng = [recording.latitude, recording.longitude];
		const map = L.map('map').setView(latLng, 4);
		L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
			maxZoom: 19,
			attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
		}).addTo(map);
		const markerGroup = L.layerGroup().addTo(map);
		L.marker(latLng).addTo(markerGroup);
	}, 10);

	const sono = document.createElement("img");
	sono.style.width = "100%";
	sono.style.height = "64px";
	sono.alt = "TODO";
	sono.style.borderRadius = "var(--small-distance)";
	sono.src = recording.sono;
	sono.style.filter = "invert(0.9)";
	content.container.appendChild(sono);

	const a = document.createElement("audio");
	a.style.width = "100%";
	a.style.height = "32px";
	a.style.marginBottom = "8px";
	a.setAttribute("controls", "");
	const source = document.createElement("source");
	source.src = recording.url;
	source.type = "audio/mpeg";
	a.appendChild(source);
	content.container.appendChild(a);
	a.addEventListener("timeupdate", function() {
		const currentTime = a.currentTime;
		const duration = a.duration;
		if (!isNaN(duration)) {
			const percentage = (currentTime / duration) * 100;
			// TODO: Use this to display progress in the sono.
			console.log(`Percentage played: ${percentage.toFixed(2)}%`);
		}
	});

	const recorderLbl = document.createElement("h3");
	recorderLbl.className = "title";
	recorderLbl.innerHTML = App.translations.Recorder;
	content.container.appendChild(recorderLbl);

	const recorder = new Ui.Input("text");
	recorder.setDisabled(true);
	recorder.value = recording.recorder;
	content.attach(recorder);

	const p = document.createElement("p");
	p.innerHTML = "xeno-canto - <a target='_blank' href='https://xeno-canto.org/'>" + App.translations.GoToWebsite + "</a>";
	p.style.fontSize = "0.8rem";
	p.style.fontWeight = "bold";
	p.style.textAlign = "right";
	content.container.appendChild(p);

	new Ui.PopUp(content.container, function () {
	});
}
