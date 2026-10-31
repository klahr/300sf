/// <reference path="AdminCommon.ts"/>

// Statistics: how the site grows and who comes back. Each chart is one series, drawn as bars per period
// (or a line for a running total) with the period and value on hover.
namespace AdminStats {
	const ns = "http://www.w3.org/2000/svg";
	const width = 520;
	const height = 120;

	interface Response {
		weeks: Array<string>;
		signups: Array<number>;
		accounts: Array<number>;
		weekly_active: Array<number>;
		sightings: Array<number>;
		lists: Array<number>;
		competitions: Array<number>;
		months: Array<string>;
		monthly_active: Array<number>;
		cohorts: Array<{ month: string, size: number, retained: Array<number> }>;
		shared_lists: number;
		recent_competitions: Array<{ name: Models.Translation, participants: number }>;
	}

	function weekLabel(monday: string): string {
		return new Date(monday + "T00:00:00").toLocaleDateString(App.language, { year: "numeric", month: "short", day: "numeric" });
	}

	function monthLabel(month: string): string {
		return new Date(month + "-01T00:00:00").toLocaleDateString(App.language, { year: "numeric", month: "short" });
	}

	// A chart under a heading, with the latest value and a quieter summary beside it.
	function figure(parent: HTMLElement, title: string, latest: number, summary: string, svg: SVGSVGElement, labels: Array<string>): void {
		const f = document.createElement("figure");
		f.className = "admin-chart";
		const caption = document.createElement("figcaption");
		caption.innerHTML = "<span>" + Utils.escapeHtml(title) + "</span><strong>" + latest.toLocaleString(App.language) + "</strong>";
		if (summary !== "") {
			const s = document.createElement("small");
			s.textContent = summary;
			caption.appendChild(s);
		}
		f.appendChild(caption);
		f.appendChild(svg);
		// The time axis: where it starts and ends.
		const axis = document.createElement("div");
		axis.className = "admin-chart-axis";
		axis.innerHTML = "<span>" + Utils.escapeHtml(labels[0]) + "</span><span>" + Utils.escapeHtml(labels[labels.length - 1]) + "</span>";
		f.appendChild(axis);
		parent.appendChild(f);
	}

	function chart(title: string, values: Array<number>): SVGSVGElement {
		const svg = document.createElementNS(ns, "svg");
		svg.setAttribute("viewBox", "0 0 " + width + " " + height);
		svg.setAttribute("preserveAspectRatio", "none");
		svg.setAttribute("role", "img");
		svg.setAttribute("aria-label", title + ": " + values.join(", "));
		const baseline = document.createElementNS(ns, "line");
		baseline.setAttribute("x1", "0");
		baseline.setAttribute("x2", String(width));
		baseline.setAttribute("y1", String(height));
		baseline.setAttribute("y2", String(height));
		baseline.setAttribute("class", "baseline");
		svg.appendChild(baseline);
		return svg;
	}

	// One hover column per period, wider than its mark, titled with the period and value.
	function hover(svg: SVGSVGElement, i: number, step: number, label: string, value: number): void {
		const g = document.createElementNS(ns, "g");
		const t = document.createElementNS(ns, "title");
		t.textContent = label + ": " + value.toLocaleString(App.language);
		g.appendChild(t);
		const hit = document.createElementNS(ns, "rect");
		hit.setAttribute("x", (i * step).toFixed(1));
		hit.setAttribute("y", "0");
		hit.setAttribute("width", step.toFixed(1));
		hit.setAttribute("height", String(height));
		hit.setAttribute("class", "hit");
		g.appendChild(hit);
		svg.appendChild(g);
	}

	function bars(title: string, values: Array<number>, labels: Array<string>): SVGSVGElement {
		const svg = chart(title, values);
		const max = Math.max(1, ...values);
		const step = width / values.length;
		// A 2px gap between bars, so neighbours read as separate.
		const w = Math.max(1, step - 2);
		values.forEach((v, i) => {
			if (v > 0) {
				const h = Math.max(2, v / max * (height - 4));
				const bar = document.createElementNS(ns, "path");
				const x = i * step + 1;
				const y = height - h;
				const r = Math.min(4, w / 2, h);
				// Rounded at the top only; the bar stands on the baseline.
				bar.setAttribute("d", "M" + x + "," + height + "V" + (y + r) + "Q" + x + "," + y + " " + (x + r) + "," + y + "H" + (x + w - r) + "Q" + (x + w) + "," + y + " " + (x + w) + "," + (y + r) + "V" + height + "Z");
				bar.setAttribute("class", "bar");
				svg.appendChild(bar);
			}
		});
		values.forEach((v, i) => hover(svg, i, step, labels[i], v));
		return svg;
	}

	function line(title: string, values: Array<number>, labels: Array<string>): SVGSVGElement {
		const svg = chart(title, values);
		const max = Math.max(1, ...values);
		const step = width / values.length;
		const x = (i: number) => i * step + step / 2;
		const y = (v: number) => height - 2 - v / max * (height - 6);
		let d = "";
		values.forEach((v, i) => d += (i === 0 ? "M" : "L") + x(i).toFixed(1) + "," + y(v).toFixed(1));
		const path = document.createElementNS(ns, "path");
		path.setAttribute("d", d);
		path.setAttribute("class", "line");
		path.setAttribute("vector-effect", "non-scaling-stroke");
		svg.appendChild(path);
		values.forEach((v, i) => hover(svg, i, step, labels[i], v));
		return svg;
	}

	function sum(values: Array<number>): number {
		return values.reduce((a, b) => a + b, 0);
	}

	// Cohorts as a table: one row per signup month, the share still reporting in each month after, shaded
	// by that share in one hue.
	function retention(parent: HTMLElement, data: Response): void {
		const wrap = document.createElement("div");
		wrap.className = "admin-retention";
		const table = document.createElement("table");
		const head = table.createTHead().insertRow();
		const th = function (text: string): void {
			const c = document.createElement("th");
			c.textContent = text;
			head.appendChild(c);
		};
		th(App.translations.StatsCohort);
		th(App.translations.StatsAccounts);
		data.months.forEach((_, i) => th(String(i)));
		const body = table.createTBody();
		data.cohorts.forEach(cohort => {
			const row = body.insertRow();
			row.insertCell().textContent = monthLabel(cohort.month);
			row.insertCell().textContent = cohort.size.toLocaleString(App.language);
			data.months.forEach((_, i) => {
				const cell = row.insertCell();
				if (i >= cohort.retained.length || cohort.size === 0) {
					return;
				}
				const share = cohort.retained[i] / cohort.size;
				cell.textContent = Math.round(share * 100) + "%";
				cell.title = cohort.retained[i] + " / " + cohort.size;
				cell.style.setProperty("--share", String(Math.round(share * 100)));
				cell.className = share >= 0.5 ? "strong" : "";
			});
		});
		wrap.appendChild(table);
		parent.appendChild(wrap);
	}

	export function open(): void {
		const sheet = new Ui.Sheet({ title: App.translations.Statistics, size: "tall" });
		const section = sheet.section();
		const charts = document.createElement("div");
		section.appendChild(charts);

		new HttpRequest("GET", "/v1/admin/stats", null).process(function (status: any, data: Response) {
			const weeks = data.weeks.map(weekLabel);
			const months = data.months.map(monthLabel);
			const last = (values: Array<number>) => values[values.length - 1];
			const total = (values: Array<number>) => App.translations.StatsTotal.replace("{0}", sum(values).toLocaleString(App.language));

			figure(charts, App.translations.StatsAccounts, last(data.accounts), "", line(App.translations.StatsAccounts, data.accounts, weeks), weeks);
			figure(charts, App.translations.StatsSignups, last(data.signups), total(data.signups), bars(App.translations.StatsSignups, data.signups, weeks), weeks);
			figure(charts, App.translations.StatsWeeklyActive, last(data.weekly_active), "", bars(App.translations.StatsWeeklyActive, data.weekly_active, weeks), weeks);
			figure(charts, App.translations.StatsMonthlyActive, last(data.monthly_active), "", bars(App.translations.StatsMonthlyActive, data.monthly_active, months), months);
			Admin.note(charts).textContent = App.translations.StatsActiveHint;
			figure(charts, App.translations.StatsSightings, last(data.sightings), total(data.sightings), bars(App.translations.StatsSightings, data.sightings, weeks), weeks);
			figure(charts, App.translations.StatsChecklists, last(data.lists), total(data.lists), bars(App.translations.StatsChecklists, data.lists, weeks), weeks);
			figure(charts, App.translations.StatsCompetitions, last(data.competitions), total(data.competitions), bars(App.translations.StatsCompetitions, data.competitions, weeks), weeks);

			const retentionTitle = document.createElement("h2");
			retentionTitle.textContent = App.translations.StatsRetention;
			charts.appendChild(retentionTitle);
			Admin.note(charts).textContent = App.translations.StatsRetentionHint;
			retention(charts, data);

			const rows = Admin.list(charts);
			Admin.row(rows, App.translations.StatsSharedChecklists, data.shared_lists.toLocaleString(App.language), null);
			const recentTitle = document.createElement("h2");
			recentTitle.textContent = App.translations.StatsRecentCompetitions;
			charts.appendChild(recentTitle);
			const recent = Admin.list(charts);
			data.recent_competitions.forEach(c => {
				Admin.row(recent, App.getTranslation(c.name), App.translations.Participants + ": " + c.participants, null);
			});
		}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
			Admin.showError(error);
		});

		sheet.open();
	}
}

function openAdminDonationPopUp(): void {
	const sheet = new Ui.Sheet({ title: App.translations.Donation, size: "compact" });
	const section = sheet.section();
	const raised = Admin.input(sheet, section, "admin-donation-raised", App.translations.DonationRaisedSek, "number");
	const goal = Admin.input(sheet, section, "admin-donation-goal", App.translations.DonationGoalSek, "number");
	Admin.note(section).textContent = App.translations.DonationHint;
	const save = sheet.addAction(Ui.actionButton(App.translations.Save, "check", "success"));
	save.disabled = true;
	const valid = function (): boolean {
		return raised.value !== "" && goal.value !== "" && parseInt(raised.value) >= 0 && parseInt(goal.value) > 0;
	};
	[raised, goal].forEach(i => i.container.addEventListener("input", function () {
		save.disabled = !valid();
	}));

	new HttpRequest("GET", "/v1/donation", null).process(function (status: any, data: { raised: number, goal: number }) {
		raised.value = String(data.raised);
		goal.value = String(data.goal);
		save.disabled = !valid();
	}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
		Admin.showError(error);
	});

	save.addEventListener("click", function () {
		Ui.setBusy(save, true);
		new HttpRequest("PUT", "/v1/donation", { raised: parseInt(raised.value), goal: parseInt(goal.value) }).process(function () {
			new Ui.Toast(App.translations.Saved, "", null, "var(--success-color)", function () { });
			sheet.close();
		}, function (status: any, data: Responses.GeneralErrorResponse, error: any) {
			Admin.showError(error);
		}, function () {
			Ui.setBusy(save, false);
		});
	});
	sheet.open();
}
