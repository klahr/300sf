const express = require('express');
const router = express.Router();
const axios = require('axios');
const config = require('./config');
const i18n = require('i18n');
const nm = require('./network-manager');
module.exports = router;

function renderPage(req, res) {
	res.render('pages/app', {
		data: {
			language: i18n.getLocale(req),
			redirect: req.path.substring(1),
			queryParameters: JSON.stringify(req.query)
		}
	});
}

router.get('/', function (req, res) {
	res.render('pages/app', {
		data: {
			language: i18n.getLocale(req),
			redirect: ""
		}
	});
});

router.get('/home', function (req, res) {
	renderPage(req, res);
});

router.get('/sign-up', function (req, res) {
	renderPage(req, res);
});

router.get('/forgot', function (req, res) {
	renderPage(req, res);
});

router.get('/settings', function (req, res) {
	renderPage(req, res);
});

router.get('/checklists', function (req, res) {
	renderPage(req, res);
});

router.get('/contribute', function (req, res) {
	renderPage(req, res);
});

router.get('/donation', function (req, res) {
	renderPage(req, res);
});

router.get('/contact', function (req, res) {
	renderPage(req, res);
});

router.get('/tags', function (req, res) {
	renderPage(req, res);
});

router.get('/reports', function (req, res) {
	renderPage(req, res);
});

router.get('/set-password', function (req, res) {
	renderPage(req, res);
});

router.get('/delete-account', function (req, res) {
	renderPage(req, res);
});

router.get('/admin', function (req, res) {
	renderPage(req, res);
});

router.get('/list/:uuid', function (req, res) {
	res.render('pages/app', {
		data: {
			language: i18n.getLocale(req),
			redirect: "list",
			queryParameters: JSON.stringify({ uuid: req.params.uuid })
		}
	});
});

router.get('/compare', function (req, res) {
	if (!req.query.uuids) {
		res.redirect(200, req.baseUrl);
	}
	const r = {
		uuids: req.query.uuids.split(',')
	}
	nm.request('POST', '/v1/lists/compare', {}, function (data, statusCode) {
		res.render('pages/compare', {
			data: data
		});
	}, function (err, statusCode) {
		console.log(err);
		res.redirect(200, req.baseUrl);
	}, JSON.stringify(r));
});

router.get('/terms', function (req, res) {
	res.render('pages/terms', {
		layout: false
	});
});

router.get('/privacy-policy', function (req, res) {
	res.render('pages/privacy-policy', {
		layout: false
	});
});

router.post('/api/v1/auth', async (req, res) => {
	const apiUrl = config.ApiEndpoint + req.originalUrl.substring(4);

	req.body.client_id = config.ClientId;
	req.body.client_secret = config.ClientSecret;

	try {
		const axiosConfig = {
			method: req.method,
			url: apiUrl,
			headers: {
				...req.headers,
				'Content-Length': Buffer.byteLength(JSON.stringify(req.body))
			},
			data: req.body,
		};

		const response = await axios(axiosConfig);

		res.status(response.status).json(response.data);
	} catch (error) {
		if (error.response) {
			res.status(error.response.status).json(error.response.data ? error.response.data : {});
		} else {
			res.status(503).json({});
		}
	}
});

router.all('/api/*', async (req, res) => {
	const apiUrl = config.ApiEndpoint + req.originalUrl.substring(4);

	req.headers['X-Api-Key'] = config.ApiKey;

	try {
		const axiosConfig = {
			method: req.method,
			url: apiUrl,
			headers: req.headers,
			data: req.body,
		};

		const response = await axios(axiosConfig);

		res.status(response.status).json(response.data);
	} catch (error) {
		if (error.response) {
			res.status(error.response.status).json(error.response.data ? error.response.data : {});
		} else {
			res.status(503).json({});
		}
	}
});
