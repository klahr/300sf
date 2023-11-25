/// <reference path="ui/Select.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="App.ts"/>
/// <reference path="ImageViewer.ts"/>
/// <reference path="models/Recording.ts"/>

function openSpeciesPopUp(species: Models.Species, list: Models.List | null) {
	const content = new Ui.Div();
	content.style.textAlign = "left";
	content.style.position = "relative";
	content.style.width = "100%";

	const name = document.createElement("h1");
	name.innerHTML = App.getTranslation(species.name);
	name.style.margin = "unset";
	content.container.appendChild(name);

	const subName = document.createElement("p");
	subName.innerHTML = species.name.la;
	subName.style.fontStyle = "italic";
	content.container.appendChild(subName);

	const buttonContainer = document.createElement("div");
	buttonContainer.style.width = "100%";
	buttonContainer.style.textAlign = "right";
	content.container.appendChild(buttonContainer);

	const reportButton = document.createElement("button");
	reportButton.innerHTML = App.translations.SubmitSighting;
	reportButton.style.width = "auto";
	reportButton.style.borderRadius = "8px";
	reportButton.style.backgroundColor = "var(--primary-color)";
	reportButton.addEventListener("click", function () {
		openReportPopUp(species, list, function () {
		});
	});
	buttonContainer.appendChild(reportButton);

	const summary = document.createElement("p");
	content.container.appendChild(summary);

	const images = document.createElement("div");
	content.container.appendChild(images);

	const recordings = document.createElement("div");
	content.container.appendChild(recordings);

	const r = new HttpRequest("GET", "/v1/species/" + species.uuid);
	r.process(function (status: any, data: any) {
		summary.innerHTML = data.wikipedia.summary;

		const p = document.createElement("p");
		p.innerHTML = "Wikipedia - <a target='_blank' href='" + data.wikipedia.url + "'>" + App.translations.GoToArticle + "</a>";
		p.style.fontSize = "0.8rem";
		p.style.fontWeight = "bold";
		p.style.textAlign = "right";
		p.style.marginBottom = "16px";
		p.style.marginRight = "16px";
		images.appendChild(p);

		if (data.wikipedia.images) {
			data.wikipedia.images.forEach(e => {
				const url = getThumbnailFromUrl(e, "190px");
				const img = new Image();
				img.onload = function () {
					const i = document.createElement("div");
					i.style.background = "url(" + url + ")";
					i.style.width = "123px";
					i.style.height = "69px";
					i.style.backgroundSize = "cover";
					i.style.backgroundPosition = "center";
					i.style.display = "inline-block";
					i.style.margin = "8px";
					i.style.cursor = "pointer";
					i.style.borderRadius = "8px";
					images.appendChild(i);
					i.addEventListener("click", function () {
						new ImageViewer(e, () => {
						});
					});
				}
				img.onerror = function () {
					console.warn("could not load: " + img.src);
				}

				img.src = url
			});
		}

		if (data.recordings) {
			(data.recordings as Array<Models.Recording>).forEach(e => {
				const p = document.createElement("p");
				p.innerHTML = e.type.charAt(0).toUpperCase() + e.type.slice(1);
				recordings.appendChild(p);
				const info = document.createElement("p");
				info.style.fontSize = "0.8rem";
				info.style.fontStyle = "italic";
				info.innerHTML = e.recorder + ", " + e.location + ", " + e.country;
				recordings.appendChild(info);
				const a = document.createElement("audio");
				a.style.width = "100%";
				a.style.height = "32px";
				a.style.marginBottom = "8px";
				a.setAttribute("controls", "");
				const source = document.createElement("source");
				source.src = e.url;
				source.type = "audio/mpeg";
				a.appendChild(source);
				recordings.appendChild(a);
			});

			const p = document.createElement("p");
			p.innerHTML = "xeno-canto - <a target='_blank' href='https://xeno-canto.org/'>" + App.translations.GoToWebsite + "</a>";
			p.style.fontSize = "0.8rem";
			p.style.fontWeight = "bold";
			p.style.textAlign = "right";
			recordings.appendChild(p);
		}
	});

	new Ui.PopUp(content.container, function () {
	});
}

function getThumbnailFromUrl(url: string, width: string): string {
	const parts = url.split("/");
	const filename = parts.pop();
	const lastHash = parts.pop();
	const secondLastHash = parts.pop();
	const newUrl = `${parts.join('/')}/thumb/${secondLastHash}/${lastHash}/${filename}/${width}-${filename}`
	return newUrl;
}
