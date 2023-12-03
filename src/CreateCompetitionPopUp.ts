/// <reference path="ui/Div.ts"/>
/// <reference path="ui/MultiInput.ts"/>
/// <reference path="ui/MultiTextArea.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="MapSelect.ts"/>
/// <reference path="models/Competition.ts"/>
/// <reference path="utils/Base64.ts"/>
/// <reference path="responses/Competitions.ts"/>

class CreateCompetitionPopUp {
	private mAddButton: HTMLButtonElement;
	private mName: Ui.MultiInput;
	private mDescription: Ui.MultiTextArea;
	private mFrom: Ui.Input;
	private mUntil: Ui.Input;
	private mTemplateSelect: Ui.Select;

	constructor(closeCallback: (competition: Models.Competition | null) => void) {
		const self = this;
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

		this.mName = new Ui.MultiInput("text");
		content.attach(this.mName);
		this.mName.addOption(App.translations.English, App.translations.Name, "en");
		this.mName.addOption(App.translations.Swedish, App.translations.Name, "sv");
		this.mName.addOption(App.translations.German, App.translations.Name, "de");
		this.mName.setSelected(App.language);
		this.mName.addEventListener("input", function() {
			self.validate();
		});

		const descriptionLbl = document.createElement("h3");
		descriptionLbl.innerHTML = "Description";
		content.container.appendChild(descriptionLbl);

		this.mDescription = new Ui.MultiTextArea();
		content.attach(this.mDescription );
		this.mDescription .addOption(App.translations.English, "Description", "en");
		this.mDescription .addOption(App.translations.Swedish, "Description", "sv");
		this.mDescription .addOption(App.translations.German, "Description", "de");
		this.mDescription .setSelected(App.language);
		this.mDescription.addEventListener("input", function() {
			self.validate();
		});

		const templateLbl = document.createElement("h3");
		templateLbl.className = "title";
		templateLbl.innerHTML = "Synlighet";
		content.container.appendChild(templateLbl);
		const visibility = new Ui.Select("Visibility" + "...", App.translations.Search);
		visibility.addOption("Public", Models.Visibility.Public.toString());
		visibility.addOption("Private", Models.Visibility.Private.toString());
		visibility.setSelected(Models.Visibility.Private.toString());
		content.attach(visibility);

		const visibilityLbl = document.createElement("h3");
		visibilityLbl.className = "title";
		visibilityLbl.innerHTML = App.translations.Templates;
		content.container.appendChild(visibilityLbl);
		this.mTemplateSelect = new Ui.Select(App.translations.SelectTemplate + "...", App.translations.Search, function (value: string | null) {
			self.validate();
		});
		content.attach(this.mTemplateSelect);

		const fromLbl = document.createElement("h3");
		fromLbl.className = "title";
		fromLbl.innerHTML = "Starts at";
		content.container.appendChild(fromLbl);
		this.mFrom = new Ui.Input("datetime-local");
		content.attach(this.mFrom);

		const now = new Date();
		now.setMinutes(now.getMinutes() - now.getTimezoneOffset());
		now.setSeconds(now.getMinutes());
		this.mFrom.value = now.toISOString().slice(0, -8);
		this.mFrom.container.addEventListener("input", function() {
			self.validate();
		});

		const untilLbl = document.createElement("h3");
		untilLbl.className = "title";
		untilLbl.innerHTML = "Ends at";
		content.container.appendChild(untilLbl);
		this.mUntil = new Ui.Input("datetime-local");
		now.setFullYear(now.getFullYear() + 1);
		this.mUntil.value = now.toISOString().slice(0, -8);
		content.attach(this.mUntil);
		this.mUntil.container.addEventListener("input", function() {
			self.validate();
		});

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

		this.mAddButton = document.createElement("button");
		this.mAddButton.innerHTML = "Create competition";
		this.mAddButton.style.width = "auto";
		this.mAddButton.style.cssFloat = "right";
		this.mAddButton.style.borderRadius = "8px";
		this.mAddButton.style.backgroundColor = "var(--success-color)";
		this.mAddButton.disabled = true;
		content.container.appendChild(this.mAddButton);
		this.mAddButton.addEventListener("click", function() {
			const r = new Requests.Competitions();
			r.name = new Models.Translation();
			r.name.en = self.mName.getValue("en");
			r.name.sv = self.mName.getValue("sv");
			r.name.de = self.mName.getValue("de");
			r.list_template_uuid = self.mTemplateSelect.value!;
			r.description = new Models.Translation();
			r.description.en = self.mDescription.getValue("en");
			r.description.sv = self.mDescription.getValue("sv");
			r.description.de = self.mDescription.getValue("de");
			r.end_on_reports = Utils.toInt(endOnReports.value);
			r.max_competitors = Utils.toInt(maxCompetitors.value);
			r.visibility = Utils.toInt(visibility.value!);
			r.color = colorSelect.value!;
			r.competition_zones_uuids = map.getSelectedZoneUuids();
			r.password = password.value;
			r.creator_join = (participateSelf.container as HTMLInputElement).checked;

			if (self.mFrom.value !== "") {
				r.from = Utils.toISODate(new Date(self.mFrom.value));
			}
			if (self.mUntil.value !== "") {
				r.until = Utils.toISODate(new Date(self.mUntil.value));
			}

			if (joinFrom.value !== "") {
				r.join_from = Utils.toISODate(new Date(joinFrom.value));
			}
			if (joinUntil.value !== "") {
				r.join_until = Utils.toISODate(new Date(joinUntil.value));
			}

			console.log(r);
		});

		new Ui.PopUp(content.container, function () {
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
			});
			d.list_templates.forEach(e => {
				self.mTemplateSelect.addOption(App.getTranslation(e.name), e.uuid);
			});
		});
	}

	private validate(): void {
		this.mAddButton.disabled = false;

		if (this.mName.getValue("sv") === "" && this.mName.getValue("en") === "" && this.mName.getValue("de") === "") {
			this.mAddButton.disabled = true;
			return;
		}

		if (this.mDescription.getValue("sv") === "" && this.mDescription.getValue("en") === "" && this.mDescription.getValue("de") === "") {
			this.mAddButton.disabled = true;
			return;
		}

		if (this.mFrom.value === "" || this.mUntil.value === "") {
			this.mAddButton.disabled = true;
			return;
		}

		if (this.mTemplateSelect.value === null) {
			this.mAddButton.disabled = true;
			return;
		}
	}
}