function getCookie(cname: string): string {
	let name = cname + "=";
	const decodedCookie = decodeURIComponent(document.cookie);
	const ca = decodedCookie.split(';');
	for (let i = 0; i <ca.length; i++) {
		let c = ca[i];
		while (c.charAt(0) == ' ') {
			c = c.substring(1);
		}
		if (c.indexOf(name) == 0) {
			return c.substring(name.length, c.length);
		}
	}
	return "";
}

function setCookie(cname: string, value: string, expiresAt: number | null = null) {
	let expires = "";
	if (expiresAt) {
	  let date = new Date();
	  date.setTime(expiresAt * 1000);
	  expires = "; expires=" + date.toUTCString();
	}
	document.cookie = cname + "=" + (value || "") + expires + "; path=/; SameSite=Lax";
}
