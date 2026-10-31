/// <reference path="models/List.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="ui/Sheet.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="utils/Sort.ts"/>
/// <reference path="ui/DivSelect.ts"/>
/// <reference path="Colors.ts"/>
/// <reference path="SpeciesListPopUp.ts"/>
/// <reference path="ui/Segmented.ts"/>
/// <reference path="ShareChecklist.ts"/>

// A checklist in the management list: what it holds, renaming, recolouring, who can see it and
// archiving it.
class EditListPopUp {
	private mSheet: Ui.Sheet;

	constructor(list: Models.List, closeCallback: () => void) {
		const self = this;

		const sheet = new Ui.Sheet({ eyebrow: App.translations.Checklist, title: App.getTranslation(list.name), subtitle: App.translations.SpeciesCount.replace("{0}", list.species.length.toString()), size: "compact" });
		this.mSheet = sheet;

		const dot = document.createElement("span");
		dot.className = "list-dot";
		dot.style.setProperty("--dot", list.color !== "" ? list.color : "var(--blue)");
		sheet.names.querySelector("p.sheet-eyebrow")?.prepend(dot);

		const form = sheet.section();

		const name = new Ui.Input("text");
		name.value = App.getTranslation(list.name);
		name.setPlaceHolder(App.translations.Name);
		sheet.field(form, App.translations.Name, name.container);

		const colorSelect = new Ui.DivSelect();
		colorSelect.addClass("swatches");
		Colors.palette.forEach(c => colorSelect.addOption("", c).style.backgroundColor = c);
		const color = list.color !== "" ? list.color : Colors.fallback;
		colorSelect.select(color);
		sheet.field(form, App.translations.Color, colorSelect.container);

		// Others see a checklist's share page only when it isn't private. Public checklists are shown as shared
		// with a link, which is what they are.
		const savedVisibility = String(ShareChecklist.visibility(list));
		let visibility = savedVisibility;
		const visibilitySelect = ShareChecklist.visibilitySelect(visibility, function (v: string) {
			visibility = v;
			saveButton.disabled = name.value.trim() === "" || !changed();
			showShare();
		});
		sheet.field(form, App.translations.Visibility, visibilitySelect);
		const visibilityHint = document.createElement("p");
		visibilityHint.className = "details";
		visibilityHint.textContent = App.translations.ChecklistVisibilityHint;
		form.appendChild(visibilityHint);

		// The share page link, for a checklist that isn't private. Copying it saves the visibility
		// first, since the link doesn't open until then (and an old list gets its id then).
		const shareControl = document.createElement("div");
		const shareUrl = new Ui.Input("text");
		(shareUrl.container as HTMLInputElement).readOnly = true;
		shareControl.appendChild(shareUrl.container);
		const copyLink = Ui.actionButton(App.translations.Copy, "content_copy", "inline");
		shareControl.appendChild(copyLink);
		const shareField = sheet.field(form, App.translations.Share, shareControl);
		const showShare = function (): void {
			shareField.style.display = visibility === String(Models.Visibility.Private) ? "none" : "";
			shareUrl.value = ShareChecklist.link(list);
		};
		showShare();
		copyLink.addEventListener("click", function () {
			const copy = function (): void {
				navigator.clipboard.writeText(shareUrl.value).then(function () {
					new Ui.Toast(App.translations.LinkCopied, Utils.escapeHtml(shareUrl.value), null, "var(--success-color)", function () { });
				});
			};
			// Lists shared before ids grew longer get a new one on the server first.
			if (visibility === String(ShareChecklist.visibility(list)) && list.uuid.length >= 10) {
				copy();
				return;
			}
			Ui.setBusy(copyLink, true);
			saveVisibility(function () {
				Ui.setBusy(copyLink, false);
				saveButton.disabled = name.value.trim() === "" || !changed();
				copy();
			}, function () {
				Ui.setBusy(copyLink, false);
				failed();
			});
		});

		// Lists made from the user's own species list can have their species changed.
		if (list.template) {
			new HttpRequest("GET", "/v1/list_templates/list/me").process(function (status: any, data: any) {
				const template = ((data.result || []) as Array<Models.ListTemplate>).find(t => t.uuid === list.template);
				if (!template) {
					return;
				}
				const edit = Ui.actionButton(App.translations.EditSpecies, "playlist_add", "secondary");
				edit.addEventListener("click", function () {
					self.mSheet.close();
					Ui.PopUp.whenHistorySettled(function () {
						openSpeciesListPopUp(template);
					});
				});
				sheet.footer.insertBefore(edit, sheet.footer.firstChild);
			});
		}

		sheet.addAction(Ui.actionButton(App.translations.ArchiveChecklist, "archive", "danger")).addEventListener("click", function () {
			new Ui.ConfirmBox(App.translations.ArchiveChecklist, App.translations.ConfirmArchiveChecklist, "/img/messagebox-256.webp", "var(--error-color)", function () {
				const r = new HttpRequest("POST", "/v1/lists/" + list.uuid + "/archive");
				r.process(function (status: any, data: any) {
					self.mSheet.close();
					setTimeout(() => {
						location.reload();
					}, 150);
				});
			}, function () {
			}, App.translations.ArchiveChecklist, App.translations.Cancel);
		});

		const saveButton = sheet.addAction(Ui.actionButton(App.translations.SaveChanges, "check", "success"));
		const changed = function (): boolean {
			return name.value.trim() !== App.getTranslation(list.name) || (colorSelect.value !== null && colorSelect.value !== color) || visibility !== String(ShareChecklist.visibility(list));
		};
		saveButton.disabled = true;
		name.container.addEventListener("input", function () {
			saveButton.disabled = name.value.trim() === "" || !changed();
		});
		colorSelect.container.addEventListener("click", function () {
			saveButton.disabled = name.value.trim() === "" || !changed();
		});
		const failed = function (): void {
			new Ui.MessageBox(App.translations.Error, App.translations.SomethingWentWrong, "/img/messagebox-256.webp", "var(--error-color)", function () { });
		};
		// Sharing an old list gives it a new id, which the other changes must then use.
		const saveVisibility = function (done: () => void, fail: () => void): void {
			new HttpRequest("POST", "/v1/lists/" + list.uuid + "/visibility", { visibility: parseInt(visibility) }).process(function (status: any, data: any) {
				list.visibility = parseInt(visibility);
				list.uuid = data.uuid;
				showShare();
				done();
			}, fail);
		};
		saveButton.addEventListener("click", function () {
			const finish = function (): void {
				self.mSheet.close();
				setTimeout(() => {
					location.reload();
				}, 150);
			};
			const fail = function (): void {
				Ui.setBusy(saveButton, false);
				failed();
			};
			const saveRest = function (): void {
				const requests: Array<HttpRequest> = [];
				if (name.value.trim() !== App.getTranslation(list.name)) {
					const n = new Models.Translation();
					n.en = n.sv = n.de = name.value.trim();
					requests.push(new HttpRequest("POST", "/v1/lists/" + list.uuid + "/rename", { name: n }));
				}
				if (colorSelect.value !== null && colorSelect.value !== color) {
					requests.push(new HttpRequest("POST", "/v1/lists/" + list.uuid + "/color", { color: colorSelect.value }));
				}
				if (requests.length === 0) {
					finish();
					return;
				}
				let pending = requests.length;
				requests.forEach(r => r.process(function () {
					if (--pending === 0) {
						finish();
					}
				}, fail));
			};
			Ui.setBusy(saveButton, true);
			if (visibility !== String(ShareChecklist.visibility(list))) {
				saveVisibility(saveRest, fail);
			} else {
				saveRest();
			}
		});

		sheet.open(closeCallback);
	}
}

function openEditListPopUp(list: Models.List, closeCallback: () => void = function () { }): void {
	new EditListPopUp(list, closeCallback);
}
