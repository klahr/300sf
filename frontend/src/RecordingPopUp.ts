/// <reference path="Maps.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="ui/Sheet.ts"/>
/// <reference path="models/Recording.ts"/>
/// <reference path="App.ts"/>

// One xeno-canto recording: the sound with its sonogram, where it was made and by whom.
function openRecordingPopUp(species: Models.Species, recording: Models.Recording) {
	const type = recording.type.charAt(0).toUpperCase() + recording.type.slice(1);
	const sheet = new Ui.Sheet({ eyebrow: type, title: App.getTranslation(species.name), subtitle: species.name.la, latin: true, size: "auto" });

	const sound = sheet.section(App.translations.Sounds);
	const sono = document.createElement("img");
	sono.className = "sonogram";
	sono.alt = "";
	sono.src = recording.sono;
	sound.appendChild(sono);

	const a = document.createElement("audio");
	a.setAttribute("controls", "");
	const source = document.createElement("source");
	source.src = recording.url;
	source.type = "audio/mpeg";
	a.appendChild(source);
	sound.appendChild(a);

	const recorder = document.createElement("p");
	recorder.className = "details";
	recorder.textContent = App.translations.Recorder + ": " + recording.recorder;
	sound.appendChild(recorder);

	const where = sheet.section(App.translations.ObservationLocation);
	const mapDiv = new Ui.Div();
	mapDiv.container.classList.add("map");
	mapDiv.style.height = "220px";
	mapDiv.style.position = "relative";
	where.appendChild(mapDiv.container);

	const place = document.createElement("p");
	place.className = "details";
	place.textContent = [recording.location, recording.country].filter(v => v).join(", ");
	where.appendChild(place);

	setTimeout(function () {
		const latLng = [recording.latitude, recording.longitude];
		const map = L.map(mapDiv.container).setView(latLng, 4);
		Maps.tileLayer().addTo(map);
		const markerGroup = L.layerGroup().addTo(map);
		L.marker(latLng).addTo(markerGroup);
	}, 10);

	const credit = document.createElement("p");
	credit.className = "credit";
	credit.innerHTML = "xeno-canto · <a target='_blank' rel='noopener' href='https://xeno-canto.org/'>" + App.translations.GoToWebsite + "</a>";
	where.appendChild(credit);

	sheet.open();
}
