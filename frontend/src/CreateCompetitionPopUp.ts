/// <reference path="MapSelect.ts"/>
/// <reference path="ui/Sheet.ts"/>
/// <reference path="ui/Select.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="models/Competition.ts"/>
/// <reference path="models/Visibility.ts"/>
/// <reference path="responses/Competitions.ts"/>
/// <reference path="requests/Competitions.ts"/>
/// <reference path="Colors.ts"/>

// The form for a new competition. created is called with it once it exists.
function openCreateCompetition(created: (competition: Models.Competition) => void): void {
	const sheet = new Ui.Sheet({ title: App.translations.NewCompetition, size: "tall" });

	const segmented = function (options: Array<[string, string, string]>, selected: string, changed: (value: string) => void): HTMLDivElement {
		const group = document.createElement("div");
		group.className = "segmented";
		group.setAttribute("role", "radiogroup");
		const select = function (value: string, notify: boolean) {
			Array.from(group.children).forEach(b => {
				const on = (b as HTMLElement).dataset.value === value;
				b.classList.toggle("selected", on);
				b.setAttribute("aria-checked", String(on));
			});
			if (notify) {
				changed(value);
			}
		};
		options.forEach(o => {
			const b = document.createElement("button");
			b.type = "button";
			b.setAttribute("role", "radio");
			b.dataset.value = o[0];
			b.innerHTML = "<span class='material-symbols-outlined'>" + o[1] + "</span>" + Utils.escapeHtml(o[2]);
			b.addEventListener("click", function () {
				select(o[0], true);
			});
			group.appendChild(b);
		});
		select(selected, false);
		return group;
	};

	const dateInput = function (id: string, value: Date): HTMLInputElement {
		const input = document.createElement("input");
		input.type = "date";
		input.id = id;
		input.value = Utils.toYYYY_MM_DD(value);
		return input;
	};

	// Name and description.
	const about = sheet.section();
	const name = document.createElement("input");
	name.type = "text";
	name.id = "competition-name";
	name.maxLength = 80;
	name.placeholder = App.translations.CompetitionNameExample;
	sheet.field(about, App.translations.Name, name);
	const description = document.createElement("textarea");
	description.id = "competition-description";
	description.rows = 2;
	description.maxLength = 500;
	sheet.field(about, App.translations.DescriptionOptional, description);

	// When.
	const when = sheet.section(App.translations.Period);
	const grid = document.createElement("div");
	grid.className = "sheet-grid";
	when.appendChild(grid);
	const today = new Date();
	const from = dateInput("competition-from", today);
	sheet.field(grid, App.translations.StartsAt, from);
	const until = dateInput("competition-until", new Date(today.getFullYear(), 11, 31));
	sheet.field(grid, App.translations.EndsAt, until);
	const dateError = document.createElement("p");
	dateError.className = "details error";
	when.appendChild(dateError);

	// What counts.
	const counts = sheet.section(App.translations.WhatCounts);
	const list = new Ui.Select(App.translations.AllSpecies, App.translations.Search, function () { });
	list.addOption(App.translations.AllSpecies, "");
	list.setSelected("");
	sheet.field(counts, App.translations.SpeciesList, list.container);

	// Where: anywhere, or in counties chosen on the map or found by name.
	const zonesField = document.createElement("div");
	const picker = document.createElement("div");
	picker.className = "zone-picker hidden";
	const selected = new Map<string, string>();
	const chips = document.createElement("div");
	chips.className = "zone-chips";
	const map = new MapSelect(function (zone: Models.RegionZone) {
		toggleZone(zone, !selected.has(zone.uuid!));
	});

	const toggleZone = function (zone: Models.RegionZone, on: boolean) {
		const uuid = zone.uuid!;
		if (on) {
			selected.set(uuid, App.getTranslation(zone.name));
		} else {
			selected.delete(uuid);
		}
		map.setSelected(uuid, on);
		showChips();
		validate();
	};
	const showChips = function () {
		chips.innerHTML = "";
		Array.from(selected.entries()).sort((a, b) => a[1].localeCompare(b[1], App.language)).forEach(([uuid, label]) => {
			const b = document.createElement("button");
			b.type = "button";
			b.className = "zone-chip";
			b.setAttribute("aria-pressed", "true");
			b.setAttribute("aria-label", App.translations.RemoveZone.replace("{0}", label));
			b.innerHTML = Utils.escapeHtml(label) + "<span class='material-symbols-outlined'>close</span>";
			b.addEventListener("click", function () {
				toggleZone({ uuid: uuid, name: { sv: label, en: label, de: label } as Models.Translation }, false);
			});
			chips.appendChild(b);
		});
	};

	const hint = document.createElement("p");
	hint.className = "details";
	hint.textContent = App.translations.ZonePickerHint;
	picker.appendChild(hint);

	// Searches run when asked, never as the user types: the place service forbids that.
	const searchForm = document.createElement("form");
	searchForm.className = "zone-search";
	const searchInput = document.createElement("input");
	searchInput.type = "search";
	searchInput.id = "competition-zone-search";
	searchInput.placeholder = App.translations.ZoneSearchPlaceholder;
	searchInput.setAttribute("aria-label", App.translations.ZoneSearchPlaceholder);
	searchInput.autocomplete = "off";
	const searchButton = document.createElement("button");
	searchButton.type = "submit";
	searchButton.className = "secondary";
	searchButton.innerHTML = "<span class='material-symbols-outlined'>search</span>" + App.translations.Search;
	searchForm.appendChild(searchInput);
	searchForm.appendChild(searchButton);
	picker.appendChild(searchForm);
	const results = document.createElement("div");
	results.className = "zone-results";
	results.setAttribute("aria-live", "polite");
	picker.appendChild(results);

	const say = function (text: string) {
		results.innerHTML = "";
		const p = document.createElement("p");
		p.className = "details";
		p.textContent = text;
		results.appendChild(p);
	};
	const failed = function (status: number) {
		say(status === 503 ? App.translations.ZoneSearchOff : status === 429 ? App.translations.ZoneSearchTooMany : App.translations.SomethingWentWrong);
	};
	searchForm.addEventListener("submit", function (e) {
		e.preventDefault();
		const text = searchInput.value.trim();
		if (text.length < 2) {
			return;
		}
		Ui.setBusy(searchButton, true);
		const r = new HttpRequest("GET", "/v1/zones/search?q=" + encodeURIComponent(text));
		r.setQuiet();
		r.process(function (status: any, data: any) {
			const found = (data.result || []) as Array<Models.RegionZone>;
			if (found.length === 0) {
				say(App.translations.ZoneSearchNone);
				return;
			}
			results.innerHTML = "";
			found.forEach(f => {
				const b = document.createElement("button");
				b.type = "button";
				b.className = "zone-result";
				b.innerHTML = "<span class='material-symbols-outlined'>add_location_alt</span><span>" + Utils.escapeHtml(App.getTranslation(f.name)) +
					(f.country ? "<span class='details'> · " + Utils.escapeHtml(f.country) + "</span>" : "") + "</span>";
				b.addEventListener("click", function () {
					Ui.setBusy(b, true);
					// Makes the zone if no sighting has yet, and gives its outline for the map.
					const z = new HttpRequest("POST", "/v1/zones/osm", { osm_id: f.osm_id });
					z.setQuiet();
					z.process(function (status: any, data: any) {
						const zone = data as Models.RegionZone;
						map.showZone(zone);
						toggleZone(zone, true);
						results.innerHTML = "";
						searchInput.value = "";
					}, failed, function () {
						Ui.setBusy(b, false);
					});
				});
				results.appendChild(b);
			});
		}, failed, function () {
			Ui.setBusy(searchButton, false);
		});
	});

	picker.appendChild(map.container);
	picker.appendChild(chips);

	let anywhere = true;
	zonesField.appendChild(segmented([["anywhere", "public", App.translations.AllowedAnywhereShort], ["counties", "location_on", App.translations.SelectedCounties]], "anywhere", function (v: string) {
		anywhere = v === "anywhere";
		picker.classList.toggle("hidden", anywhere);
		if (!anywhere) {
			map.show();
		}
		validate();
	}));
	zonesField.appendChild(picker);
	sheet.field(counts, App.translations.WhereSightingsCount, zonesField);

	// How it ends.
	const ends = sheet.section(App.translations.Winner);
	const target = document.createElement("input");
	target.type = "number";
	target.id = "competition-target";
	target.min = "1";
	target.step = "1";
	target.value = "100";
	const targetField = document.createElement("div");
	targetField.className = "hidden";
	let firstTo = false;
	ends.appendChild(segmented([["most", "leaderboard", App.translations.MostSpeciesWinsShort], ["first", "flag", App.translations.FirstToReach]], "most", function (v: string) {
		firstTo = v === "first";
		targetField.classList.toggle("hidden", !firstTo);
		validate();
	}));
	ends.appendChild(targetField);
	sheet.field(targetField, App.translations.NumberOfSpeciesToWin, target);

	// Who takes part.
	const who = sheet.section(App.translations.Participants);
	let visibility = Models.Visibility.Private;
	const visibilityHint = document.createElement("p");
	visibilityHint.className = "details";
	visibilityHint.textContent = App.translations.CompetitionPrivateHint;
	who.appendChild(segmented([[String(Models.Visibility.Private), "lock", App.translations.CompetitionPrivate], [String(Models.Visibility.Public), "public", App.translations.CompetitionPublic]], String(Models.Visibility.Private), function (v: string) {
		visibility = Number(v);
		visibilityHint.textContent = visibility === Models.Visibility.Public ? App.translations.CompetitionPublicHint : App.translations.CompetitionPrivateHint;
	}));
	who.appendChild(visibilityHint);
	const whoGrid = document.createElement("div");
	whoGrid.className = "sheet-grid";
	who.appendChild(whoGrid);
	const max = document.createElement("input");
	max.type = "number";
	max.id = "competition-max";
	max.min = "2";
	max.step = "1";
	max.placeholder = App.translations.NoLimit;
	sheet.field(whoGrid, App.translations.MaxCompetitors, max);
	const password = document.createElement("input");
	password.type = "text";
	password.id = "competition-password-new";
	password.autocomplete = "off";
	password.maxLength = 64;
	password.placeholder = App.translations.None;
	sheet.field(whoGrid, App.translations.PasswordOptional, password);
	const joinRow = document.createElement("div");
	joinRow.className = "check-row";
	const join = document.createElement("input");
	join.type = "checkbox";
	join.id = "competition-join";
	join.checked = true;
	const joinLabel = document.createElement("label");
	joinLabel.htmlFor = join.id;
	joinLabel.textContent = App.translations.IParticipate;
	joinRow.appendChild(join);
	joinRow.appendChild(joinLabel);
	who.appendChild(joinRow);

	// Look.
	const look = sheet.section(App.translations.Color);
	const swatches = document.createElement("div");
	swatches.className = "color-swatches";
	swatches.setAttribute("role", "radiogroup");
	let color = Colors.fallback;
	Colors.palette.forEach(c => {
		const b = document.createElement("button");
		b.type = "button";
		b.setAttribute("role", "radio");
		b.setAttribute("aria-label", c);
		b.style.backgroundColor = c;
		b.addEventListener("click", function () {
			color = c;
			Array.from(swatches.children).forEach(e => e.setAttribute("aria-checked", String(e === b)));
		});
		b.setAttribute("aria-checked", String(c === color));
		swatches.appendChild(b);
	});
	look.appendChild(swatches);

	const create = sheet.addAction(Ui.actionButton(App.translations.CreateCompetition, "emoji_events", "success"));
	const error = document.createElement("p");
	error.className = "details error";
	sheet.footer.insertBefore(error, create);

	const selectedZones = function (): Array<string> {
		return Array.from(selected.keys());
	};

	const validate = function () {
		const start = new Date(from.value);
		const end = new Date(until.value);
		const datesOk = from.value !== "" && until.value !== "" && end >= start;
		dateError.textContent = datesOk || from.value === "" || until.value === "" ? "" : App.translations.EndBeforeStart;
		create.disabled = name.value.trim() === "" || !datesOk ||
			(!anywhere && selectedZones().length === 0) ||
			(firstTo && !(Number(target.value) >= 1));
	};
	[name, from, until, target].forEach(e => e.addEventListener("input", validate));

	new HttpRequest("GET", "/v1/competitions/basis").process(function (status: any, data: any) {
		const basis = data as Responses.CompetitionsBasis;
		basis.list_templates.forEach(t => {
			list.addOption(App.getTranslation(t.name) + " · " + App.translations.NumberOfSpecies.replace("{0}", String(t.size)), t.uuid);
		});
	});

	create.addEventListener("click", function () {
		const text = name.value.trim();
		const about = description.value.trim();
		const r = new Requests.Competitions();
		r.name = { sv: text, en: text, de: text } as Models.Translation;
		r.description = { sv: about, en: about, de: about } as Models.Translation;
		// Whole days in the user's time zone: from the first day's midnight to the one after the last.
		const start = new Date(from.value + "T00:00:00");
		const end = new Date(until.value + "T00:00:00");
		end.setDate(end.getDate() + 1);
		r.from = start.toISOString();
		r.until = end.toISOString();
		r.list_template_uuid = list.value || "";
		r.end_on_reports = firstTo ? Math.floor(Number(target.value)) : 0;
		r.max_competitors = max.value !== "" ? Math.max(2, Math.floor(Number(max.value))) : 0;
		r.visibility = visibility;
		r.color = color;
		r.competition_zones_uuids = anywhere ? [] : selectedZones();
		r.password = password.value;
		r.creator_join = join.checked;

		error.textContent = "";
		Ui.setBusy(create, true);
		new HttpRequest("POST", "/v1/competitions", r).process(function (status: any, data: any) {
			sheet.close();
			Ui.PopUp.whenHistorySettled(function () {
				created(data as Models.Competition);
			});
		}, function (status: any, data: any) {
			error.textContent = status === 400 && data && data.message && data.message.indexOf("future") >= 0 ? App.translations.EndInPast : App.translations.SomethingWentWrong;
		}, function () {
			Ui.setBusy(create, false);
		});
	});

	validate();
	sheet.open(function () {
		map.remove();
	});
	name.focus();
}
