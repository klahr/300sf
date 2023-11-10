/// <reference path="App.ts"/>

class AdminAnnouncementsPopUp {
	private mAddButton: HTMLButtonElement;
	private mEnglishContent: HTMLTextAreaElement;
	private mSwedishContent: HTMLTextAreaElement;
	private mGermanContent: HTMLTextAreaElement;
	private mName: Ui.Input;
	private mPriority: Ui.Select;

	constructor(closeCallback: () => void) {
		const self = this;

		const content = new Ui.Div();
		content.style.textAlign = "left";
		content.style.width = "100%";
		content.style.position = "relative";

		const name = document.createElement("h1");
		name.innerHTML = "Announcements";
		name.style.margin = "unset";
		content.container.appendChild(name);

		new Ui.PopUp(content.container, function () {
			closeCallback();
		});

		const addAnnouncementLbl = document.createElement("h2");
		addAnnouncementLbl.innerHTML = "Add announcement";
		content.container.appendChild(addAnnouncementLbl);

		this.mName = new Ui.Input("text");
		this.mName.setPlaceHolder(App.translations.Name);
		content.attach(this.mName);
		this.mName.container.addEventListener("input", () => {
			self.verify();
		});

		this.mEnglishContent = document.createElement("textarea");
		this.mEnglishContent.placeholder = App.translations.Message + " (english)";
		this.mEnglishContent.style.height = "128px";
		content.container.appendChild(this.mEnglishContent);
		this.mEnglishContent.addEventListener("input", () => {
			self.verify();
		});

		this.mPriority = new Ui.Select("Priority", "", (value: string | null) => {
			self.verify();
		});
		this.mPriority.addOption("Low", "0");
		this.mPriority.addOption("Normal", "1");
		this.mPriority.addOption("High", "2");
		content.attach(this.mPriority);

		const fromLbl = document.createElement("h3");
		fromLbl.className = "title";
		fromLbl.innerHTML = "From";
		content.container.appendChild(fromLbl);
		const from = new Ui.Input("date");
		content.attach(from);

		const untilLbl = document.createElement("h3");
		untilLbl.className = "title";
		untilLbl.innerHTML = "Until";
		content.container.appendChild(untilLbl);
		const until = new Ui.Input("date");
		content.attach(until);

		const onHold = new Ui.Input("checkbox");
		onHold.style.position = "relative";
		onHold.container.setAttribute("name", "on_hold");
		onHold.container.id = "on_hold";
		content.attach(onHold);

		const onHoldLbl = document.createElement("label");
		onHoldLbl.setAttribute("for", onHold.container.id);
		onHoldLbl.innerHTML = "On hold";
		content.container.appendChild(onHoldLbl);

		this.mAddButton = document.createElement("button");
		this.mAddButton.innerHTML = "Add announcement";
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
				},
				from: from.value,
				until: until.value,
				on_hold: onHold.value === "on",
				priority: parseInt(self.mPriority.value!)
			};
			new Ui.ConfirmBox("Confirm announcement addition", "Are you sure?", null, "var(--success-color)", function () {
				const r = new HttpRequest("POST", "/v1/announcement", d);
				r.process((status: AnalyserNode, data: any) => {
					new Ui.MessageBox(App.translations.Success, "", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--success-color)", function () { });
					self.refresh();
				}, (status: any, data: Responses.GeneralErrorResponse, error: any) => {
					new Ui.MessageBox(App.translations.Error, error, "", "var(--error-color)", function () { });
				}, () => {
					self.mAddButton.disabled = false;
					self.mAddButton.innerHTML = "Add announcement";
					self.mName.setDisabled(false);
					self.mEnglishContent.disabled = false;
					from.setDisabled(false);
					until.setDisabled(false);
					onHold.setDisabled(false);
				});
			}, function () {
				self.mAddButton.disabled = false;
				self.mAddButton.innerHTML = "Add announcement";
				self.mName.setDisabled(false);
				self.mEnglishContent.disabled = false;
				from.setDisabled(false);
				until.setDisabled(false);
				onHold.setDisabled(false);
			}, "Add announcement", App.translations.Cancel);
		});

		this.refresh();
	}

	private refresh(): void {
		const r = new HttpRequest("GET", "/v1/announcements/all", null);
		const self = this;
		r.process((status: any, data: any) => {
			console.log(data);
		}, (status: any, data: Responses.GeneralErrorResponse, error: any) => {

		}, undefined);
	}

	private verify(): void {
		this.mAddButton.disabled = this.mPriority.value === null || this.mEnglishContent.value.length < 4 || this.mName.value.length < 4;
	}
}

function openAdminAnnouncementsPopUp(closeCallback: () => void = function () { }): void {
	new AdminAnnouncementsPopUp(closeCallback);
}