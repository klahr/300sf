/// <reference path="ui/Sheet.ts"/>
/// <reference path="App.ts"/>
/// <reference path="models/Announcement.ts"/>
/// <reference path="Colors.ts"/>

class AdminAnnouncementsPopUp {
	private mAddButton: HTMLButtonElement;
	private mEnglishContent: HTMLTextAreaElement;
	private mSwedishContent: HTMLTextAreaElement;
	private mGermanContent: HTMLTextAreaElement;
	private mName: Ui.Input;
	private mIcon: Ui.Input;
	private mAnnouncementsDiv: Ui.Div;

	constructor(closeCallback: () => void) {
		const self = this;

		const content = new Ui.Div();
		content.style.textAlign = "left";
		content.style.width = "100%";
		content.style.position = "relative";


		Ui.Sheet.wrap(content.container, App.translations.Announcements).open(function () {
			closeCallback();
		});

		this.mAnnouncementsDiv = new Ui.Div();
		this.mAnnouncementsDiv.style.position = "relative";
		this.mAnnouncementsDiv.style.width = "100%";
		content.attach(this.mAnnouncementsDiv);

		const addAnnouncementLbl = document.createElement("h2");
		addAnnouncementLbl.innerHTML = App.translations.AddAnnouncement;
		content.container.appendChild(addAnnouncementLbl);

		this.mName = new Ui.Input("text");
		this.mName.setPlaceHolder(App.translations.Name);
		content.attach(this.mName);
		this.mName.container.addEventListener("input", () => {
			self.verify();
		});

		this.mIcon = new Ui.Input("text");
		this.mIcon.setPlaceHolder(App.translations.IconName);
		content.attach(this.mIcon);
		this.mIcon.container.addEventListener("input", () => {
			self.verify();
		});

		this.mEnglishContent = document.createElement("textarea");
		this.mEnglishContent.placeholder = App.translations.Message + " (english)";
		this.mEnglishContent.style.height = "128px";
		content.container.appendChild(this.mEnglishContent);
		this.mEnglishContent.addEventListener("input", () => {
			self.verify();
		});

		this.mSwedishContent = document.createElement("textarea");
		this.mSwedishContent.placeholder = App.translations.Message + " (swedish)";
		this.mSwedishContent.style.height = "128px";
		content.container.appendChild(this.mSwedishContent);

		this.mGermanContent = document.createElement("textarea");
		this.mGermanContent.placeholder = App.translations.Message + " (german)";
		this.mGermanContent.style.height = "128px";
		content.container.appendChild(this.mGermanContent);

		const fromLbl = document.createElement("h3");
		fromLbl.className = "title";
		fromLbl.innerHTML = App.translations.From;
		content.container.appendChild(fromLbl);
		const from = new Ui.Input("date");
		content.attach(from);

		const untilLbl = document.createElement("h3");
		untilLbl.className = "title";
		untilLbl.innerHTML = App.translations.Until;
		content.container.appendChild(untilLbl);
		const until = new Ui.Input("date");
		content.attach(until);

		const colorSelect = new Ui.DivSelect();
		colorSelect.addClass("swatches");
		Colors.palette.forEach(c => colorSelect.addOption("", c).style.backgroundColor = c);
		content.attach(colorSelect);
		colorSelect.select(Colors.fallback);

		const onHold = new Ui.Input("checkbox");
		onHold.style.position = "relative";
		onHold.container.setAttribute("name", "on_hold");
		onHold.container.id = "on_hold";
		(onHold.container as HTMLInputElement).checked = true;
		content.attach(onHold);

		const onHoldLbl = document.createElement("label");
		onHoldLbl.setAttribute("for", onHold.container.id);
		onHoldLbl.innerHTML = App.translations.Draft;
		content.container.appendChild(onHoldLbl);

		this.mAddButton = document.createElement("button");
		this.mAddButton.innerHTML = App.translations.AddAnnouncement;
		this.mAddButton.disabled = true;
		content.container.appendChild(this.mAddButton);
		this.mAddButton.addEventListener("click", function () {
			self.mAddButton.disabled = true;
			self.mAddButton.innerHTML = App.translations.PleaseWait + "...";
			self.mName.setDisabled(true);
			self.mEnglishContent.disabled = true;
			from.setDisabled(true);
			until.setDisabled(true);
			onHold.setDisabled(true);
			const d = {
				name: self.mName.value,
				content: {
					en: self.mEnglishContent.value,
					sv: self.mSwedishContent.value,
					de: self.mGermanContent.value
				},
				on_hold: (onHold.container as HTMLInputElement).checked,
				icon: self.mIcon.value,
				color: colorSelect.value,
				from: "",
				until: ""
			};
			if (from.value !== "") {
				d.from = Utils.toISODate(new Date(from.value));
			}
			if (until.value !== "") {
				d.until = Utils.toISODate(new Date(until.value));
			}
			const r = new HttpRequest("POST", "/v1/announcements", d);
			r.process((status: AnalyserNode, data: any) => {
				new Ui.MessageBox(App.translations.Success, "", "/img/messagebox-256.webp", "var(--success-color)", function () { });
				self.refresh();
			}, (status: any, data: Responses.GeneralErrorResponse, error: any) => {
				new Ui.MessageBox(App.translations.Error, error, "/img/messagebox-256.webp", "var(--error-color)", function () { });
			}, () => {
				self.mAddButton.disabled = false;
				self.mAddButton.innerHTML = App.translations.AddAnnouncement;
				self.mName.setDisabled(false);
				self.mEnglishContent.disabled = false;
				from.setDisabled(false);
				until.setDisabled(false);
				onHold.setDisabled(false);
			});
		});

		this.refresh();
	}

	private refresh(): void {
		const r = new HttpRequest("GET", "/v1/announcements/all", null);
		const self = this;
		r.process((status: any, data: any) => {
			self.mAnnouncementsDiv.clear();
			const announcements = data.data as Array<Models.Announcement>;
			announcements.forEach(e => {
				const n = document.createElement("h2");
				n.textContent = e.name;
				self.mAnnouncementsDiv.container.appendChild(n);

				if (Utils.isValidDate(e.from) || Utils.isValidDate(e.until)) {
					const timespan = document.createElement("p");
					timespan.innerHTML = "";
					self.mAnnouncementsDiv.container.appendChild(timespan);
					if (Utils.isValidDate(e.from)) {
						timespan.innerHTML = Utils.toYYYY_MM_DD(e.from);
					}
					if (Utils.isValidDate(e.until)) {
						timespan.innerHTML += " ⮕ " + Utils.toYYYY_MM_DD(e.until);
					}
				}

				const p = document.createElement("span");
				p.className = "material-symbols-outlined";
				p.style.cursor = "pointer";
				p.style.fontSize = "32px";

				if (Utils.isValidDate(e.on_hold)) {
					p.style.color = "var(--error-color)";
					p.innerHTML = "play_arrow";
					p.addEventListener("click", function () {
						const r = new HttpRequest("PUT", "/v1/announcements/" + e.uuid + "/resume");
						r.process((status: any, data: any) => {
							self.refresh();
						});
					});
				} else {
					p.style.color = "var(--success-color)";
					p.innerHTML = "pause";
					p.addEventListener("click", function () {
						const r = new HttpRequest("PUT", "/v1/announcements/" + e.uuid + "/pause");
						r.process((status: any, data: any) => {
							self.refresh();
						});
					});
				}
				self.mAnnouncementsDiv.container.appendChild(p);

				const t = document.createElement("span");
				t.className = "material-symbols-outlined";
				t.style.cursor = "pointer";
				t.style.fontSize = "32px";
				t.style.cssFloat = "right";
				t.style.color = "var(--secondary-color)";
				t.innerHTML = "delete_forever";
				t.addEventListener("click", function () {
					new Ui.ConfirmBox(App.translations.AreYouSure, App.translations.ConfirmDeleteAnnouncement, null, "var(--error-color)", function() {
						const r = new HttpRequest("DELETE", "/v1/announcements/" + e.uuid);
						r.process((status: any, data: any) => {
							self.refresh();
						});	
					}, function() {}, App.translations.Delete, App.translations.Cancel);
				});
				self.mAnnouncementsDiv.container.appendChild(t);

				const a = document.createElement("div");
				a.className = "announcement";
				self.mAnnouncementsDiv.container.appendChild(a);
				const img = document.createElement("p");
				const icon = document.createElement("span");
				icon.className = "material-symbols-outlined";
				icon.style.fontSize = "2rem";
				icon.style.color = e.color;
				icon.textContent = e.icon;
				img.appendChild(icon);
				a.style.boxShadow = "inset 0 0 0px 1px " + e.color;

				a.appendChild(img);

				const close = document.createElement("span");
				close.className = "material-symbols-outlined close";
				close.style.position = "absolute";
				close.style.cursor = "pointer";
				close.style.fontSize = "24px";
				close.style.right = "var(--small-distance)";
				close.style.top = "var(--small-distance)";
				close.style.color = "var(--secondary-color)";
				close.innerHTML = "close";
				a.appendChild(close);

				const c = document.createElement("p");
				c.innerHTML = App.getTranslation(e.content);
				a.appendChild(c);
			});
		}, (status: any, data: Responses.GeneralErrorResponse, error: any) => {

		}, undefined);
	}

	private verify(): void {
		this.mAddButton.disabled = this.mEnglishContent.value.length < 4 || this.mName.value.length < 4;
	}
}

function openAdminAnnouncementsPopUp(closeCallback: () => void = function () { }): void {
	new AdminAnnouncementsPopUp(closeCallback);
}