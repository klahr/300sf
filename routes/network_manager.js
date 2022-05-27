var http = require('http');

var NetworkManager = {
	ApiVersion: {
		V1: "v1"
	},

	Method: {
		POST: "POST",
		GET: "GET",
		PUT: "PUT",
		DELETE: "DELETE"
	},

	hasAccessToken(req) {
		return req.cookies.sf300_access_token;
	},

	isSuccessful(statusCode) {
		return statusCode >= 200 && statusCode < 300;
	},

	getHost() {
		return "127.0.0.1";
	},

	getPort() {
		return "8090";
	},

	isJson(str) {
		try {
			JSON.parse(str);
		} catch (e) {
			return false;
		}
		return true;
	},

	request(method, path, headers, successCallback, errorCallback, body = "", apiVersion = this.ApiVersion.V1) {
		const options = {
			hostname: this.getHost(),
			port: this.getPort(),
			path: encodeURI(path),
			method: method,
			headers: headers
		};

		const req = http.request(options, (res) => {
			res.setEncoding('utf8');
			var data = "";
			res.on('data', (chunk) => {
				data += chunk;
			});
			res.on('end', () => {
				//console.log(data);
				//console.log(res.statusCode);
				if (this.isSuccessful(res.statusCode)) {
					if (res.statusCode === 204) {
						successCallback(null, res.statusCode);
					} else {
						if (this.isJson(data)) {
							successCallback(JSON.parse(data), res.statusCode);
						} else {
							successCallback(data, res.statusCode);
						}
					}
				} else {
					console.log(data);
					errorCallback(JSON.parse(data), res.statusCode);
				}
			});
		});
	
		req.on('error', (e) => {
			console.error(`problem with request: ${e.message}`);
		});
	
		req.write(body);
		req.end();
	},
	
	onNotLoggedIn(req, res) {
		res.redirect('/login?from=' + req.url);
	}
}

module.exports = NetworkManager;
