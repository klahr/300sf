/// <reference path="responses/ReportsBasis.ts"/>
/// <reference path="models/Species.ts"/>
/// <reference path="models/List.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="HttpRequest.ts"/>

function openReportPopUp(species: Models.Species | null, list: Models.List | null): void {
	const content = new Ui.Div();
	content.style.textAlign = "left";
    content.style.width = "100%";
	content.style.position = "relative";

	const name = document.createElement("h1");
	name.innerHTML = "Submit";
	name.style.margin = "unset";
	content.container.appendChild(name);

    const speciesListLbl = document.createElement("h3");
    speciesListLbl.className = "title";
    speciesListLbl.innerHTML = "Species";
    content.container.appendChild(speciesListLbl);
	const speciesList = new Ui.Select("Select species...");
	content.attach(speciesList);

	const ageListLbl = document.createElement("h3");
    ageListLbl.className = "title";
    ageListLbl.innerHTML = "Age";
    content.container.appendChild(ageListLbl);
	const ageList = new Ui.Select("Select age...");
	content.attach(ageList);

	const genderListLbl = document.createElement("h3");
    genderListLbl.className = "title";
    genderListLbl.innerHTML = "Gender";
    content.container.appendChild(genderListLbl);
	const genderList = new Ui.Select("Select gender...");
	content.attach(genderList);

	const methodListLbl = document.createElement("h3");
    methodListLbl.className = "title";
    methodListLbl.innerHTML = "Method";
    content.container.appendChild(methodListLbl);
	const methodList = new Ui.Select("Select method...");
	content.attach(methodList);

	const activityListLbl = document.createElement("h3");
    activityListLbl.className = "title";
    activityListLbl.innerHTML = "Activity";
    content.container.appendChild(activityListLbl);
	const activityList = new Ui.Select("Select activity...");
	content.attach(activityList);

	const quantityLbl = document.createElement("h3");
	quantityLbl.className = "title";
	quantityLbl.innerHTML = "Quantity";
	content.container.appendChild(quantityLbl);
	const quantity = new Ui.Input("number");
	quantity.container.setAttribute("min", "1");
	quantity.container.setAttribute("step", "1");
	content.attach(quantity);

	const fromLbl = document.createElement("h3");
	fromLbl.className = "title";
	fromLbl.innerHTML = "Observed from";
	content.container.appendChild(fromLbl);
	const from = new Ui.Input("date");
	content.attach(from);

	const untilLbl = document.createElement("h3");
	untilLbl.className = "title";
	untilLbl.innerHTML = "Observed until";
	content.container.appendChild(untilLbl);
	const until = new Ui.Input("date");
	content.attach(until);

	const privateCommentLbl = document.createElement("h3");
	privateCommentLbl.className = "title";
	privateCommentLbl.innerHTML = "Private comment";
	content.container.appendChild(privateCommentLbl);
	const privateComment = document.createElement("textarea");
	privateComment.style.width = "100%";
	privateComment.style.height = "128px";
	privateComment.style.resize = "none";
	privateComment.style.padding = "var(--small-distance)";
	privateComment.style.boxSizing = "border-box";
	content.container.appendChild(privateComment);

    const r = new HttpRequest("GET", "/v1/reports/basis");
	r.process(function(status: any, data: any) {
        const d = data as Responses.ReportsBasis;
        console.log(d);
        d.species.forEach(e => {
            speciesList.addOption(App.getTranslation(e.name), e.uuid);
        });
		d.ages.forEach(e => {
			ageList.addOption(App.getTranslation(e.name), e.uuid);
		});
		d.genders.forEach(e => {
			genderList.addOption(App.getTranslation(e.name), e.uuid);
		});
		d.methods.forEach(e => {
			methodList.addOption(App.getTranslation(e.name), e.uuid);
		});
		d.activities.forEach(e => {
			activityList.addOption(App.getTranslation(e.name), e.uuid);
		}
    });

	new Ui.PopUp(content.container, function() {
		console.log("closed");
	});
}
