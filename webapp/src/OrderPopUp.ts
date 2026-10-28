/// <reference path="ui/Select.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="App.ts"/>
/// <reference path="ImageViewer.ts"/>

function openOrderPopUp(order: Models.Order) {
	const content = new Ui.Div();
	content.style.textAlign = "left";
	content.style.position = "relative";
	content.style.width = "100%";

	const name = document.createElement("h1");
	name.innerHTML = App.getTranslation(order.name);
	name.style.margin = "unset";
	content.container.appendChild(name);

	const subName = document.createElement("p");
	subName.innerHTML = order.name.la;
	subName.style.fontStyle = "italic";
	content.container.appendChild(subName);

	const summary = document.createElement("p");
	content.container.appendChild(summary);

	const images = document.createElement("div");
	content.container.appendChild(images);

	const r = new HttpRequest("GET", "/v1/orders/" + order.uuid);
	r.process(function (status: any, data: any) {
		summary.textContent = data.wikipedia.summary;

		const p = document.createElement("p");
		const articleUrl = Utils.safeUrl(data.wikipedia.url);
		p.innerHTML = "Wikipedia" + (articleUrl ? " - <a target='_blank' rel='noopener' href='" + Utils.escapeHtml(articleUrl) + "'>" + App.translations.GoToArticle + "</a>" : "");
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
	});

	new Ui.PopUp(content.container, function () {
	});
}
