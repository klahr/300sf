const express = require('express');
const router = express.Router();
const config = require('./config');
const i18n = require('i18n');
module.exports = router;

function renderPage(req, res) {
	res.render('pages/app', {
		data: {
			language: i18n.getLocale(req),
			api_key: config.ApiKey,
			redirect: req.path.substring(1),
			queryParameters: JSON.stringify(req.query)
		}
	});
}

router.get('/', function(req, res) {
	res.render('pages/app', {
		data: {
			language: i18n.getLocale(req),
			api_key: config.ApiKey,
			redirect: ""
		}
	});
});

router.get('/home', function(req, res) {
	renderPage(req, res);
});

router.get('/sign-up', function(req, res) {
	renderPage(req, res);
});

router.get('/settings', function(req, res) {
	renderPage(req, res);
});

router.get('/list/:uuid', function(req, res) {
	res.render('pages/app', {
		data: {
			language: i18n.getLocale(req),
			api_key: config.ApiKey,
			redirect: "list",
			queryParameters: JSON.stringify({uuid: req.params.uuid})
		}
	});
});
