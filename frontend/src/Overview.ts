/// <reference path="ui/Progress.ts"/>
/// <reference path="models/Report.ts"/>
/// <reference path="models/List.ts"/>
/// <reference path="utils/Date.ts"/>
/// <reference path="utils/Html.ts"/>

// The dashboard's numbers: a row of stat tiles and one tile per checklist with its progress and the new
// species it gained per month.
namespace Overview {
	const POINTS = 24;

	// The date each species was first reported, keyed by species uuid.
	export function firstSeen(reports: Array<Models.Report>): Map<string, Date> {
		const first = new Map<string, Date>();
		reports.forEach(r => {
			if (!Utils.isValidDate(r.from)) {
				return;
			}
			const d = new Date(r.from);
			const known = first.get(r.species.uuid);
			if (!known || d < known) {
				first.set(r.species.uuid, d);
			}
		});
		return first;
	}

	export function statTile(icon: string, label: string, value: number): HTMLDivElement {
		const tile = document.createElement("div");
		tile.className = "stat-tile";

		const i = document.createElement("span");
		i.className = "material-symbols-outlined";
		i.textContent = icon;
		tile.appendChild(i);

		const v = document.createElement("p");
		v.className = "stat-value";
		Ui.countUp(v, value);
		tile.appendChild(v);

		const l = document.createElement("p");
		l.className = "stat-label";
		l.textContent = label;
		tile.appendChild(l);

		return tile;
	}

	export function statRow(listReports: Array<Responses.ListReport>, lastReports: Array<Models.Report>): HTMLDivElement {
		const all = Array<Models.Report>();
		listReports.forEach(e => {
			e.reports.forEach(r => all.push(r));
		});
		const first = firstSeen(all);
		const species = new Set<string>();
		all.forEach(r => species.add(r.species.uuid));
		const year = new Date().getFullYear();
		let newThisYear = 0;
		first.forEach(d => {
			if (d.getFullYear() === year) {
				++newThisYear;
			}
		});

		const row = document.createElement("div");
		row.className = "stat-row";
		row.appendChild(statTile("raven", App.translations.SpeciesSeen, species.size));
		row.appendChild(statTile("auto_awesome", App.translations.NewSpeciesThisYear, newThisYear));
		row.appendChild(statTile("visibility", App.translations.SightingsLast14Days, lastReports.length));
		return row;
	}

	// The number of species seen at evenly spaced moments from the list's start to now, or to its end
	// if it has ended.
	function growth(list: Models.List, reports: Array<Models.Report>): Array<{ date: Date, count: number }> {
		const dates = Array<Date>();
		firstSeen(reports).forEach(d => dates.push(d));
		dates.sort((a, b) => a.getTime() - b.getTime());

		let end = new Date();
		if (Utils.isValidDate(list.until) && new Date(list.until) < end) {
			end = new Date(list.until);
		}
		let start = dates.length > 0 ? dates[0] : end;
		if (Utils.isValidDate(list.from) && new Date(list.from) < start) {
			start = new Date(list.from);
		}
		if (end < start) {
			end = start;
		}

		const points = Array<{ date: Date, count: number }>();
		let i = 0;
		for (let p = 0; p < POINTS; ++p) {
			const t = new Date(start.getTime() + (end.getTime() - start.getTime()) * p / (POINTS - 1));
			while (i < dates.length && dates[i] <= t) {
				++i;
			}
			points.push({ date: t, count: i });
		}
		// Every line starts from nothing; sightings dated after the end still count at the last point.
		points[0].count = 0;
		points[POINTS - 1].count = dates.length;
		return points;
	}

	function sparkline(points: Array<{ date: Date, count: number }>): SVGSVGElement {
		const ns = "http://www.w3.org/2000/svg";
		const width = 200;
		const height = 40;
		const pad = 4;
		const max = Math.max(1, points[points.length - 1].count);
		const x = (i: number) => i / (points.length - 1) * width;
		const y = (c: number) => pad + (height - pad) - c / max * (height - pad * 2);

		const svg = document.createElementNS(ns, "svg");
		svg.setAttribute("viewBox", "0 0 " + width + " " + height);
		svg.setAttribute("preserveAspectRatio", "none");
		svg.setAttribute("class", "sparkline");
		svg.setAttribute("role", "img");
		svg.setAttribute("aria-label", App.translations.SpeciesSeen + ": " + points[0].count + " → " + points[points.length - 1].count);

		let line = "";
		points.forEach((p, i) => {
			line += (i === 0 ? "M" : "L") + x(i).toFixed(1) + "," + y(p.count).toFixed(1);
		});

		const area = document.createElementNS(ns, "path");
		area.setAttribute("d", line + "L" + width + "," + height + "L0," + height + "Z");
		area.setAttribute("class", "area");
		svg.appendChild(area);

		const path = document.createElementNS(ns, "path");
		path.setAttribute("d", line);
		path.setAttribute("class", "line");
		path.setAttribute("vector-effect", "non-scaling-stroke");
		svg.appendChild(path);

		// One hover column per point, with the date and the count so far.
		points.forEach((p, i) => {
			const g = document.createElementNS(ns, "g");
			const title = document.createElementNS(ns, "title");
			title.textContent = p.date.toLocaleDateString(App.language, { year: "numeric", month: "short", day: "numeric" }) + ": " + p.count;
			g.appendChild(title);
			const hit = document.createElementNS(ns, "rect");
			hit.setAttribute("x", (x(i) - width / (points.length - 1) / 2).toFixed(1));
			hit.setAttribute("y", "0");
			hit.setAttribute("width", (width / (points.length - 1)).toFixed(1));
			hit.setAttribute("height", height.toString());
			hit.setAttribute("class", "hit");
			g.appendChild(hit);
			svg.appendChild(g);
		});

		return svg;
	}

	export function listTile(listReport: Responses.ListReport, onOpen: () => void): HTMLDivElement {
		const list = listReport.list;
		const seen = Utils.numberOfUniqueSpeciesInReports(listReport.reports);
		const total = list.species.length;
		const percent = total > 0 ? Math.floor(seen / total * 100) : 0;

		const tile = document.createElement("div");
		tile.className = "list-tile";
		tile.setAttribute("role", "link");
		tile.tabIndex = 0;
		if (list.color !== "") {
			tile.style.setProperty("--list-color", list.color);
		}
		tile.addEventListener("click", onOpen);
		tile.addEventListener("keydown", function (event) {
			if (event.key === "Enter") {
				onOpen();
			}
		});

		const name = document.createElement("p");
		name.className = "list-tile-name";
		name.textContent = App.getTranslation(list.name);
		tile.appendChild(name);

		const figures = document.createElement("div");
		figures.className = "list-tile-figures";
		tile.appendChild(figures);

		const count = document.createElement("p");
		count.className = "list-tile-count";
		const strong = document.createElement("strong");
		Ui.countUp(strong, seen, 1000);
		count.appendChild(strong);
		count.appendChild(document.createTextNode(" / " + total));
		figures.appendChild(count);

		const pct = document.createElement("p");
		pct.className = "list-tile-percent";
		pct.textContent = percent + "%";
		figures.appendChild(pct);

		const meter = document.createElement("div");
		meter.className = "list-tile-meter";
		meter.setAttribute("role", "meter");
		meter.setAttribute("aria-valuemin", "0");
		meter.setAttribute("aria-valuemax", total.toString());
		meter.setAttribute("aria-valuenow", seen.toString());
		const fill = document.createElement("div");
		meter.appendChild(fill);
		tile.appendChild(meter);
		requestAnimationFrame(function () {
			requestAnimationFrame(function () {
				fill.style.width = (total > 0 ? seen / total * 100 : 0) + "%";
			});
		});

		const trend = document.createElement("div");
		trend.className = "list-tile-trend";
		const points = growth(list, listReport.reports);
		const chart = document.createElement("div");
		chart.className = "list-tile-chart";
		chart.appendChild(sparkline(points));
		const dot = document.createElement("span");
		dot.className = "end-dot";
		dot.style.bottom = (points[points.length - 1].count > 0 ? "calc(100% - 4px)" : "4px");
		chart.appendChild(dot);
		trend.appendChild(chart);
		// An empty list has no start worth naming.
		if (seen > 0) {
			const caption = document.createElement("p");
			caption.textContent = App.translations.ProgressSince.replace("{0}", points[0].date.getFullYear().toString());
			trend.appendChild(caption);
		}
		tile.appendChild(trend);

		return tile;
	}
}
