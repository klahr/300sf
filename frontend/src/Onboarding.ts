/// <reference path="HttpRequest.ts"/>
/// <reference path="requests/Lists.ts"/>
/// <reference path="models/Visibility.ts"/>
/// <reference path="ui/Progress.ts"/>
/// <reference path="Colors.ts"/>

// Helps people without a checklist: every sighting belongs to one, so this is the first thing a new
// account needs.
namespace Onboarding {
	// Creates the "300 Swedish Birds" list (or the first template if it is missing) for the user.
	export function createStarterList(done: (uuid: string | null) => void): void {
		new HttpRequest("GET", "/v1/list_templates/list").process(function (status: any, data: any) {
			const templates = (data.result || []) as Array<any>;
			const template = templates.find(t => (t.name.en || "").indexOf("300") === 0) || templates[0];
			if (!template) {
				done(null);
				return;
			}
			const r = new Requests.Lists();
			r.name = new Models.Translation();
			r.name.en = template.name.en;
			r.name.sv = template.name.sv;
			r.name.de = template.name.de;
			r.template = template.uuid;
			r.visibility = Models.Visibility.Hidden;
			r.color = Colors.fallback;
			new HttpRequest("POST", "/v1/lists", r).process(function (status: any, data: any) {
				window.dispatchEvent(new Event("listschanged"));
				done(data.uuid);
			}, function () {
				done(null);
			});
		}, function () {
			done(null);
		});
	}

	// A card that explains checklists and creates the starter list in one tap.
	export function starterCard(onCreated: (uuid: string) => void): HTMLElement {
		const card = document.createElement("div");
		card.className = "starter-card";

		const icon = document.createElement("span");
		icon.className = "material-symbols-outlined";
		icon.textContent = "checklist";
		card.appendChild(icon);

		const title = document.createElement("h3");
		title.textContent = App.translations.StartFirstList;
		card.appendChild(title);

		const text = document.createElement("p");
		text.textContent = App.translations.StartFirstListDescription;
		card.appendChild(text);

		const actions = document.createElement("div");
		actions.className = "starter-actions";
		card.appendChild(actions);

		const start = document.createElement("button");
		start.className = "inline";
		start.innerHTML = "<span class='material-symbols-outlined'>add</span>" + App.translations.StartStarterList;
		start.addEventListener("click", function () {
			Ui.setBusy(start, true);
			createStarterList(function (uuid: string | null) {
				if (uuid) {
					onCreated(uuid);
				} else {
					Ui.setBusy(start, false);
					new Ui.MessageBox(App.translations.Error, App.translations.SomethingWentWrong, "/img/messagebox-256.webp", "var(--error-color)", function () { });
				}
			});
		});
		actions.appendChild(start);

		const other = document.createElement("a");
		other.href = "/checklists";
		other.textContent = App.translations.ChooseOtherList;
		actions.appendChild(other);

		return card;
	}
}
