const http = require('http');

const NetworkManager = {
	request(method, path, headers, success, failed, body = "") {
		const options = {
			hostname: '127.0.0.1',
			port: '8080',
			path: encodeURI(path),
			method: method,
			headers: headers
		};

		const req = http.request(options, (res) => {
			res.setEncoding('utf8');
			var data = '';
			res.on('data', (chunk) => {
				data += chunk;
			});
			res.on('end', () => {
				switch (res.statusCode) {
					case 200:
						success(JSON.parse(data), res.statusCode);
						break;
					case 204:
						success(null, res.statusCode);
						break;
					default:
						failed(JSON.parse(data), res.statusCode);
						break;
				}
			});
		});

		req.on('error', (e) => {
			console.error(`problem with request: ${e.message}`);
		});

		req.write(JSON.stringify(body));
		req.end();
	}
}

module.exports = NetworkManager;
