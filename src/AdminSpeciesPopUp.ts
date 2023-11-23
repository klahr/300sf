/// <reference path="App.ts"/>

class AdminSpeciesPopUp {
	private mSpeciesList: Ui.Select;
	private mOrdersList: Ui.Select;
	private mAddButton: HTMLButtonElement;
	private mEnglish: Ui.Input;
	private mSwedish: Ui.Input;
	private mGerman: Ui.Input;
	private mLatin: Ui.Input;
	private mSpecies: Array<Models.Species> = new Array();

	constructor(closeCallback: () => void) {
		const self = this;

		const content = new Ui.Div();
		content.style.textAlign = "left";
		content.style.width = "100%";
		content.style.position = "relative";

		const name = document.createElement("h1");
		name.innerHTML = App.translations.SpeciesPlural;
		name.style.margin = "unset";
		content.container.appendChild(name);

		this.mSpeciesList = new Ui.Select(App.translations.AllSpecies, "", (value: string | null) => {
			self.mSpecies.forEach(e => {
				if (e.uuid === value) {
					openSpeciesPopUp(e, null);
					return;
				}
			});
		});

		content.attach(this.mSpeciesList);

		new Ui.PopUp(content.container, function () {
			closeCallback();
		});

		const addSpeciesLbl = document.createElement("h2");
		addSpeciesLbl.innerHTML = App.translations.AddSpecies;
		content.container.appendChild(addSpeciesLbl);

		this.mEnglish = new Ui.Input("text");
		this.mEnglish.setPlaceHolder(App.translations.English);
		content.attach(this.mEnglish);
		this.mEnglish.container.addEventListener("input", () => {
			self.verify();
		});

		this.mLatin = new Ui.Input("text");
		this.mLatin.setPlaceHolder(App.translations.Latin);
		content.attach(this.mLatin);
		this.mLatin.container.addEventListener("input", () => {
			self.verify();
		});

		this.mSwedish = new Ui.Input("text");
		this.mSwedish.setPlaceHolder(App.translations.Swedish);
		content.attach(this.mSwedish);
		this.mSwedish.container.addEventListener("input", () => {
			self.verify();
		});

		this.mGerman = new Ui.Input("text");
		this.mGerman.setPlaceHolder(App.translations.German);
		content.attach(this.mGerman);
		this.mGerman.container.addEventListener("input", () => {
			self.verify();
		});

		this.mOrdersList = new Ui.Select(App.translations.Order, "", (value: string | null) => {
			self.verify();
		});
		content.attach(this.mOrdersList);

		this.mAddButton = document.createElement("button");
		this.mAddButton.innerHTML = App.translations.AddSpecies;
		this.mAddButton.disabled = true;
		content.container.appendChild(this.mAddButton);
		this.mAddButton.addEventListener("click", function () {
			self.mAddButton.disabled = true;
			self.mAddButton.innerHTML = App.translations.PleaseWait + "...";
			self.mEnglish.setDisabled(true);
			self.mSwedish.setDisabled(true);
			self.mGerman.setDisabled(true);
			self.mLatin.setDisabled(true);
			const d = {
				name: {
					en: self.mEnglish.value,
					sv: self.mSwedish.value,
					de: self.mGerman.value,
					la: self.mLatin.value
				},
				order_uuid: self.mOrdersList.value
			};
			new Ui.ConfirmBox(App.translations.ConfirmSpeciesAddition, App.translations.ConfirmSpeciesAdditionBody, null, "var(--success-color)", function () {
				const r = new HttpRequest("POST", "/v1/species", d);
				r.process((status: any, data: any) => {
					new Ui.MessageBox(App.translations.Success, "", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--success-color)", function () { });
					self.refresh();
				}, (status: any, data: Responses.GeneralErrorResponse, error: any) => {
					new Ui.MessageBox(App.translations.Error, error, "", "var(--error-color)", function () { });
				}, () => {
					self.mAddButton.disabled = false;
					self.mEnglish.setDisabled(false);
					self.mSwedish.setDisabled(false);
					self.mGerman.setDisabled(false);
					self.mLatin.setDisabled(false);
					self.mAddButton.innerHTML = App.translations.AddSpecies;
				});
			}, function () {
				self.mAddButton.disabled = false;
				self.mEnglish.setDisabled(false);
				self.mSwedish.setDisabled(false);
				self.mGerman.setDisabled(false);
				self.mLatin.setDisabled(false);
				self.mAddButton.innerHTML = App.translations.AddSpecies;
			}, App.translations.AddSpecies, App.translations.Cancel);
		});

		this.refresh();
	}

	private refresh(): void {
		const r = new HttpRequest("GET", "/v1/species/list", null);
		const self = this;
		r.process((status: any, data: any) => {
			self.mSpecies = data.result;
			self.mSpecies.sort(Utils.sortByNameOnCurrentLanguage);
			self.mSpeciesList.clear();
			self.mSpecies.forEach(e => {
				const s = e as Models.Species;
				self.mSpeciesList.addOption(App.getTranslation(s.name), s.uuid);
			});
		}, (status: any, data: Responses.GeneralErrorResponse, error: any) => {
		}, undefined);

		const orders = new HttpRequest("GET", "/v1/orders/list", null);
		orders.process((status: any, data: any) => {
			self.mOrdersList.clear();
			data.result.sort(Utils.sortByNameOnCurrentLanguage);
			data.result.forEach(e => {
				const o = e as Models.Order;
				self.mOrdersList.addOption(App.getTranslation(o.name), o.uuid);
			});
		});
	}

	private verify(): void {
		this.mAddButton.disabled = this.mEnglish.value === "" || this.mLatin.value === "" || this.mOrdersList.value === null;
	}
}

function openAdminSpeciesPopUp(closeCallback: () => void = function () { }): void {
	new AdminSpeciesPopUp(closeCallback);
}