/// <reference path="ui/Div.ts"/>
/// <reference path="ui/MultiInput.ts"/>
/// <reference path="ui/MultiTextArea.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="models/Competition.ts"/>
/// <reference path="utils/Base64.ts"/>
/// <reference path="responses/Competitions.ts"/>

class CreateCompetitionPopUp {
	constructor(closeCallback: (competition: Models.Competition | null) => void) {
		const content = new Ui.Div();
		content.style.textAlign = "left";
		content.style.width = "100%";
		content.style.position = "relative";

		const title = document.createElement("h1");
		title.innerHTML = "New competition";
		title.style.margin = "unset";
		content.container.appendChild(title);

		const nameLbl = document.createElement("h3");
		nameLbl.innerHTML = App.translations.Name;
		content.container.appendChild(nameLbl);

		const name = new Ui.MultiInput("text");
		content.attach(name);
		name.addOption(App.translations.English, App.translations.Name, "en");
		name.addOption(App.translations.Swedish, App.translations.Name, "sv");
		name.addOption(App.translations.German, App.translations.Name, "de");
		name.setSelected(App.language);

		const descriptionLbl = document.createElement("h3");
		descriptionLbl.innerHTML = "Description";
		content.container.appendChild(descriptionLbl);

		const description = new Ui.MultiTextArea();
		content.attach(description);
		description.addOption(App.translations.English, "Description", "en");
		description.addOption(App.translations.Swedish, "Description", "sv");
		description.addOption(App.translations.German, "Description", "de");
		description.setSelected(App.language);

		const fromLbl = document.createElement("h3");
		fromLbl.className = "title";
		fromLbl.innerHTML = "Starts at";
		content.container.appendChild(fromLbl);
		const from = new Ui.Input("datetime-local");
		content.attach(from);

		const now = new Date();
		now.setMinutes(now.getMinutes() - now.getTimezoneOffset());
		now.setSeconds(now.getMinutes());
		from.value = now.toISOString().slice(0, -8);

		const untilLbl = document.createElement("h3");
		untilLbl.className = "title";
		untilLbl.innerHTML = "Ends at";
		content.container.appendChild(untilLbl);
		const until = new Ui.Input("datetime-local");
		content.attach(until);

		const endOnReportsLbl = document.createElement("h3");
		endOnReportsLbl.innerHTML = "End on reports";
		content.container.appendChild(endOnReportsLbl);

		const endOnReports = new Ui.Input("number");
		endOnReports.setPlaceHolder("No");
		endOnReports.container.setAttribute("min", "0");
		content.attach(endOnReports);

		const maxCompetitorsLbl = document.createElement("h3");
		maxCompetitorsLbl.innerHTML = "Max number of competitors";
		content.container.appendChild(maxCompetitorsLbl);

		const maxCompetitors = new Ui.Input("number");
		maxCompetitors.setPlaceHolder("No limit");
		maxCompetitors.container.setAttribute("min", "1");
		content.attach(maxCompetitors);

		const joinFromLbl = document.createElement("h3");
		joinFromLbl.className = "title";
		joinFromLbl.innerHTML = "Can join from";
		content.container.appendChild(joinFromLbl);
		const joinFrom = new Ui.Input("datetime-local");
		content.attach(joinFrom);

		const joinUntilLbl = document.createElement("h3");
		joinUntilLbl.className = "title";
		joinUntilLbl.innerHTML = "Can join until";
		content.container.appendChild(joinUntilLbl);
		const joinUntil = new Ui.Input("datetime-local");
		content.attach(joinUntil);

		const colorLbl = document.createElement("h3");
		colorLbl.className = "title";
		colorLbl.innerHTML = App.translations.Color;
		content.container.appendChild(colorLbl);

		const colorSelect = new Ui.DivSelect();
		colorSelect.style.width = "100%";
		colorSelect.addOption("", "#8227ac").style.backgroundColor = "#8227ac";
		colorSelect.addOption("", "#0d81b4").style.backgroundColor = "#0d81b4";
		colorSelect.addOption("", "#75b98e").style.backgroundColor = "#75b98e";
		colorSelect.addOption("", "#68e060").style.backgroundColor = "#68e060";
		colorSelect.addOption("", "#e06868").style.backgroundColor = "#e06868";
		colorSelect.addOption("", "#ac7827").style.backgroundColor = "#ac7827";
		content.attach(colorSelect);
		colorSelect.select("#68e060");

		const passwordLbl = document.createElement("h3");
		passwordLbl.innerHTML = App.translations.Password;
		content.container.appendChild(passwordLbl);

		const password = new Ui.Input("text");
		password.setPlaceHolder("None");
		content.attach(password);

		const mapLbl = document.createElement("h3");
		mapLbl.innerHTML = "Locations";
		content.container.appendChild(mapLbl);

		const map = new MapSelect((zoneUuids: Array<string>) => {
			if (zoneUuids.length > 0) {
				place.innerHTML = "Observations are only permitted within the selected countries.";
			} else {
				place.innerHTML = "Observations are allowed anywhere.";
			}
		});
		content.attach(map);

		const place = document.createElement("p");
		place.style.fontStyle = "italic";
		content.container.appendChild(place);

		const participateSelf = new Ui.Input("checkbox");
		participateSelf.style.position = "relative";
		participateSelf.container.setAttribute("name", "on_hold");
		participateSelf.container.id = "on_hold";
		(participateSelf.container as HTMLInputElement).checked = true;
		content.attach(participateSelf);

		const participateSelfLbl = document.createElement("label");
		participateSelfLbl.setAttribute("for", participateSelf.container.id);
		participateSelfLbl.innerHTML = "Jag deltar i tävlingen";
		content.container.appendChild(participateSelfLbl);

		const addButton = document.createElement("button");
		addButton.innerHTML = "Create competition";
		addButton.style.width = "auto";
		addButton.style.cssFloat = "right";
		addButton.style.borderRadius = "8px";
		addButton.style.backgroundColor = "var(--success-color)";
		addButton.disabled = true;
		content.container.appendChild(addButton);

		const popUp = new Ui.PopUp(content.container, function () {
			closeCallback(null);
		});

		const r = new HttpRequest("GET", "/v1/competitions/basis");
		r.process(function (status: any, data: any) {
			const d = data as Responses.CompetitionsBasis;
			d.zones.forEach(e => {
				map.addZone(e);
				if (e.name.en === "Sweden") {
					map.setZoneSelected(e.uuid);
				}
			})
		});
	}
}