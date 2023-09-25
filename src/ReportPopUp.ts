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

	const asdf = document.createElement("h3");
    asdf.className = "title";
    asdf.innerHTML = "Species ok!";
    content.container.appendChild(asdf);

    const r = new HttpRequest("GET", "/v1/reports/basis");
	r.process(function(status: any, data: any) {
        const d = data as Responses.ReportsBasis;
        console.log(d);
        d.species.forEach(e => {
            speciesList.addOption(App.getTranslation(e.name), e.uuid);
        });
    });

	new Ui.PopUp(content.container, function() {
		console.log("closed");
	});
}
