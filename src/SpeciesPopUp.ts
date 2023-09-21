function openSpeciesPopUp(species: Models.Species) {
	const content = new Ui.Div();
	content.style.textAlign = "left";
	content.style.position = "relative";

	const name = document.createElement("h1");
	name.innerHTML = App.getTranslation(species.name);
	name.style.margin = "unset";
	content.container.appendChild(name);

	const subName = document.createElement("p");
	subName.innerHTML = species.name.la;
	subName.style.fontStyle = "italic";
	content.container.appendChild(subName);

	const summary = document.createElement("p");
	summary.style.margin = "16px";
	content.container.appendChild(summary);

	const images = document.createElement("div");
	content.container.appendChild(images);

	console.log(species);
	const r = new HttpRequest("GET", "/v1/species/" + species.uuid);
	r.process(function(status: any, data: any) {
		console.log(data);
		summary.innerHTML = data.wikipedia.summary;

		if (data.wikipedia.images) {
			data.wikipedia.images.forEach(e => {
				const img = document.createElement("div");
				img.style.background = "url(" + e + ")";
				img.style.width = "123px";
				img.style.height = "69px";
				img.style.backgroundSize = "cover";
				img.style.display = "inline-block";
				img.style.margin = "8px";
				img.style.borderRadius = "8px";
				images.appendChild(img);
			});
		}
	});

	new Ui.PopUp(content.container, function() {
		console.log("closed");
	});
}
