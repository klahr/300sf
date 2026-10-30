/// <reference path="ui/Toast.ts"/>

// Every release is a new service worker (sw.js carries the version), which takes over at once. The
// page keeps running the code it was loaded with, so offer a reload. Checked at start, every hour and
// whenever the app is shown again, so a tab left open for days still notices.
namespace Updates {
	export function watch(): void {
		if (!("serviceWorker" in navigator)) {
			return;
		}
		// The first install takes control too; that is no update.
		const hadController = navigator.serviceWorker.controller !== null;
		let offered = false;
		navigator.serviceWorker.addEventListener("controllerchange", function () {
			if (!hadController || offered) {
				return;
			}
			offered = true;
			new Ui.Toast(App.translations.UpdateAvailable, App.translations.TapToUpdate, null, "var(--primary)", function () {
				window.location.reload();
			}, 0);
		});

		navigator.serviceWorker.register("/sw.js").then(function (reg) {
			// Offline it can't check; the next time will.
			const check = function () {
				reg.update().catch(function () { });
			};
			setInterval(check, 60 * 60 * 1000);
			document.addEventListener("visibilitychange", function () {
				if (document.visibilityState === "visible") {
					check();
				}
			});
		}).catch(function () { });
	}
}
