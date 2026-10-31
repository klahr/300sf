/// <reference path="HttpRequest.ts"/>
/// <reference path="models/List.ts"/>
/// <reference path="models/Visibility.ts"/>
/// <reference path="ui/Sheet.ts"/>
/// <reference path="ui/Segmented.ts"/>

// Sharing a checklist means sharing its page, /share/<id>, which only opens for a list that isn't private
// and has a long id. The share sheet sets who can see the list and hands out the link; a list from before
// ids grew longer gets a new one when it is shared.
namespace ShareChecklist {
	export function link(list: Models.List): string {
		return window.location.origin + "/share/" + list.uuid;
	}

	// A checklist is private or shared with a link; one made public before is shared with a link.
	export function visibility(list: Models.List): number {
		return list.visibility === Models.Visibility.Public ? Models.Visibility.Hidden : list.visibility;
	}

	export function visibilitySelect(selected: string, changed: (value: string) => void): HTMLDivElement {
		return Ui.segmented([
			[String(Models.Visibility.Private), "lock", App.translations.Private],
			[String(Models.Visibility.Hidden), "link", App.translations.WithLink],
		], selected, changed);
	}

	function copy(url: string): void {
		navigator.clipboard.writeText(url).then(function () {
			new Ui.Toast(App.translations.LinkCopied, Utils.escapeHtml(url), null, "var(--success-color)", function () { });
		});
	}

	// The phone's own share sheet where there is one, otherwise the clipboard.
	function send(list: Models.List): void {
		const url = link(list);
		const nav = navigator as any;
		if (nav.share) {
			nav.share({ title: App.getTranslation(list.name), url: url }).catch(function (e: any) {
				if (!e || e.name !== "AbortError") {
					copy(url);
				}
			});
			return;
		}
		copy(url);
	}

	// The share sheet. changed is called when the list's visibility or id changed.
	export function open(list: Models.List, changed: () => void): void {
		const sheet = new Ui.Sheet({ eyebrow: App.translations.ShareChecklist, title: App.getTranslation(list.name), size: "compact" });
		const form = sheet.section();

		const select = visibilitySelect(String(visibility(list)), function (v: string) {
			save(parseInt(v));
		});
		sheet.field(form, App.translations.Visibility, select);
		const hint = document.createElement("p");
		hint.className = "details";
		hint.textContent = App.translations.ChecklistVisibilityHint;
		form.appendChild(hint);

		const linkControl = document.createElement("div");
		const url = new Ui.Input("text");
		(url.container as HTMLInputElement).readOnly = true;
		linkControl.appendChild(url.container);
		const copyLink = Ui.actionButton(App.translations.Copy, "content_copy", "inline");
		linkControl.appendChild(copyLink);
		copyLink.addEventListener("click", function () {
			copy(url.value);
		});
		const linkField = sheet.field(form, App.translations.Share, linkControl);

		const shareButton = sheet.addAction(Ui.actionButton(App.translations.Share, "share", "success"));
		shareButton.addEventListener("click", function () {
			send(list);
		});

		const show = function (): void {
			const shared = list.visibility !== Models.Visibility.Private;
			linkField.style.display = shared ? "" : "none";
			shareButton.disabled = !shared;
			url.value = link(list);
			select.querySelectorAll("button").forEach(b => {
				const on = b.dataset.value === String(visibility(list));
				b.classList.toggle("selected", on);
				b.setAttribute("aria-checked", String(on));
			});
		};

		// Saved at once, so the link works as soon as it shows.
		const save = function (value: number): void {
			Ui.setBusy(shareButton, true);
			new HttpRequest("POST", "/v1/lists/" + list.uuid + "/visibility", { visibility: value }).process(function (status: any, data: any) {
				list.visibility = value;
				list.uuid = data.uuid;
				changed();
				window.dispatchEvent(new Event("listschanged"));
			}, function () {
				new Ui.MessageBox(App.translations.Error, App.translations.SomethingWentWrong, "/img/messagebox-256.webp", "var(--error-color)", function () { });
			}, function () {
				Ui.setBusy(shareButton, false);
				show();
			});
		};

		show();
		if (list.visibility !== Models.Visibility.Private && list.uuid.length < 10) {
			save(list.visibility);
		}
		sheet.open(function () { });
	}
}
