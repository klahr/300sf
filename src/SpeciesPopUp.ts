/// <reference path="ui/Select.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="App.ts"/>

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
	summary.style.margin = "16px 16px 0 16px";
	content.container.appendChild(summary);

	const images = document.createElement("div");
	content.container.appendChild(images);

	const recordings = document.createElement("div");
	content.container.appendChild(recordings);

	const r = new HttpRequest("GET", "/v1/species/" + species.uuid);
	r.process(function(status: any, data: any) {
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
				img.onload = function() {
					const i = document.createElement("div");
					i.style.background = "url(" + url + ")";
					i.style.width = "123px";
					i.style.height = "69px";
					i.style.backgroundSize = "cover";
					i.style.backgroundPosition = "center";
					i.style.display = "inline-block";
					i.style.margin = "8px";
					i.style.borderRadius = "8px";
					images.appendChild(i);
				}
				img.onerror = function() {
					console.warn("could not load: " + img.src);
				}

				img.src = url
			});
		}

		if (data.recordings) {
			data.recordings.forEach(e => {
				const p = document.createElement("p");
				p.style.fontStyle = "italic";
				p.innerHTML = e.type.charAt(0).toUpperCase() + e.type.slice(1);
				recordings.appendChild(p);
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

	new Ui.PopUp(content.container, function() {
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
