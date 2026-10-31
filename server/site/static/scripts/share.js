// The share pages: on a screen nobody touches, the species scroll slowly past, pausing at the ends.
// Anyone scrolling themselves takes over for a while.
(function () {
	var panel = document.getElementById("species");
	if (!panel || window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
		return;
	}
	var pausedUntil = performance.now() + 5000;
	var position = 0;
	var last = 0;
	function hold(ms) {
		pausedUntil = performance.now() + ms;
	}
	["wheel", "touchstart", "keydown", "mousedown"].forEach(function (e) {
		panel.addEventListener(e, function () {
			hold(30000);
		}, { passive: true });
	});
	function step(now) {
		var dt = Math.min(now - last, 100);
		last = now;
		var room = panel.scrollHeight - panel.clientHeight;
		if (room > 4 && now > pausedUntil) {
			if (Math.abs(panel.scrollTop - position) > 2) {
				position = panel.scrollTop;
			}
			position += dt * 0.03;
			if (position >= room) {
				position = room;
				hold(6000);
				setTimeout(function () {
					panel.scrollTo({ top: 0, behavior: "smooth" });
					position = 0;
					hold(6000);
				}, 6000);
			}
			panel.scrollTop = position;
		}
		requestAnimationFrame(step);
	}
	requestAnimationFrame(function (now) {
		last = now;
		step(now);
	});
})();
