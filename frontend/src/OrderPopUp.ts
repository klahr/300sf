/// <reference path="ui/Div.ts"/>
/// <reference path="ui/Sheet.ts"/>
/// <reference path="App.ts"/>
/// <reference path="ImageViewer.ts"/>
/// <reference path="SpeciesPopUp.ts"/>

// A bird order, laid out like a species: photo header, the Wikipedia summary and photos.
function openOrderPopUp(order: Models.Order) {
	const sheet = new Ui.Sheet({ title: App.getTranslation(order.name), subtitle: order.name.la, latin: true, hero: true });

	const about = sheet.section(App.translations.AboutOrder);
	const summary = document.createElement("p");
	summary.className = "species-summary";
	summary.innerHTML = "<span class='skeleton line'></span><span class='skeleton line'></span><span class='skeleton line'></span><span class='skeleton line short'></span>";
	about.appendChild(summary);

	const readMore = document.createElement("a");
	readMore.className = "sheet-link hidden";
	readMore.target = "_blank";
	readMore.rel = "noopener";
	readMore.innerHTML = App.translations.ReadOnWikipedia + " <span class='material-symbols-outlined'>open_in_new</span>";
	about.appendChild(readMore);

	const photos = sheet.section(App.translations.Photos);
	const strip = photoStrip(photos);

	const r = new HttpRequest("GET", "/v1/orders/" + order.uuid);
	r.process(function (status: any, data: any) {
		const wikipedia = data.wikipedia || {};
		sheet.setHeroImage(wikipediaHeroUrl(wikipedia));
		summary.textContent = wikipedia.summary ? wikipedia.summary : App.translations.NoProvidedInfo;
		const articleUrl = Utils.safeUrl(wikipedia.url);
		if (articleUrl) {
			readMore.href = articleUrl;
			readMore.classList.remove("hidden");
		}
		fillPhotoStrip(photos, strip, (wikipedia.images || []).filter(isPhotoUrl));
	}, function () {
		sheet.setHeroImage("");
		summary.textContent = App.translations.NoProvidedInfo;
		photos.classList.add("hidden");
	});

	sheet.open();
}
