const express = require('express');
const router = express.Router();
const axios = require('axios');
const config = require('./settings');
const i18n = require('i18n');
module.exports = router;

// Client-side routes; the single-page app picks the page from `redirect`.
const pages = [
	'home', 'sign-up', 'forgot', 'settings', 'checklists', 'contribute', 'donation',
	'contact', 'tags', 'reports', 'set-password', 'delete-account', 'competitions',
	'admin', 'experimental'
];

// JSON that is safe to place inside a <script> element.
function scriptJson(value) {
	return JSON.stringify(value)
		.replace(/</g, '\\u003c')
		.replace(/>/g, '\\u003e')
		.replace(/&/g, '\\u0026')
		.replace(/\u2028/g, '\\u2028')
		.replace(/\u2029/g, '\\u2029');
}

function renderApp(req, res, redirect, queryParameters) {
	res.render('pages/app', {
		data: {
			boot: scriptJson({
				language: i18n.getLocale(req),
				redirect: redirect,
				queryParameters: queryParameters
			})
		}
	});
}

router.get('/', function (req, res) {
	renderApp(req, res, '', {});
});

for (const page of pages) {
	router.get('/' + page, function (req, res) {
		renderApp(req, res, page, req.query);
	});
}

router.get('/list/:uuid', function (req, res) {
	renderApp(req, res, 'list', { uuid: req.params.uuid });
});

router.get('/compare', async (req, res) => {
	if (!req.query.uuids) {
		return res.redirect(req.baseUrl || '/');
	}

	try {
		const response = await axios.post(config.ApiEndpoint + '/v1/lists/compare', {
			uuids: String(req.query.uuids).split(',')
		});
		res.render('pages/compare', { data: response.data });
	} catch (error) {
		console.log(error.message);
		res.redirect(req.baseUrl || '/');
	}
});

router.get('/terms', function (req, res) {
	res.render('pages/terms', { layout: false });
});

router.get('/privacy-policy', function (req, res) {
	res.render('pages/privacy-policy', { layout: false });
});

async function forward(req, res, url, headers, data) {
	try {
		const response = await axios({
			method: req.method,
			url: url,
			headers: headers,
			data: data
		});
		res.status(response.status).json(response.data);
	} catch (error) {
		if (error.response) {
			res.status(error.response.status).json(error.response.data ? error.response.data : {});
		} else {
			res.status(503).json({});
		}
	}
}

// The API key never reaches the browser: it is added here on login.
router.post('/api/v1/auth', (req, res) => {
	const body = { ...req.body, client_id: config.ClientId, client_secret: config.ClientSecret };
	const headers = { ...req.headers, 'content-length': Buffer.byteLength(JSON.stringify(body)) };
	forward(req, res, config.ApiEndpoint + req.originalUrl.substring(4), headers, body);
});

// Audio uploads (multipart) are passed through untouched.
router.post('/analyze', express.raw({ type: () => true, limit: '20mb' }), (req, res) => {
	forward(req, res, config.AnalyzeEndpoint, req.headers, req.body);
});

router.all('/api/*', (req, res) => {
	forward(req, res, config.ApiEndpoint + req.originalUrl.substring(4), req.headers, req.method == 'GET' ? null : req.body);
});
