const express = require('express');
const router = express.Router();
const nm = require('./network-manager');
const config = require('./config');
const i18n = require('i18n');
module.exports = router;

router.get('/', function(req, res) {
	res.render('pages/app', { data: { language: i18n.getLocale(req), api_key: config.ApiKey }});
});

router.get('/old', function(req, res) {
	nm.request('GET', '/v1/list_templates/kjhkf4', {'X-Api-Key': config.ApiKey}, function(response) {
		console.log(response);
	}, function(error) {
		console.log(error);
	});
	res.render('pages/home');
});

router.get('/terms', function(req, res) {
	res.render('pages/terms', { data: {} });
});

router.get('/login', function(req, res) {
	res.render('pages/login', { data: {} });
});

router.get('/signup', function(req, res) {
	res.render('pages/signup');
});

router.get('/dashboard', function(req, res) {
	res.render('pages/dashboard');
});

router.get('/list', function(req, res) {
	res.render('pages/list');
});

router.post('/signup', function(req, res) {
	const body = {
		email: req.body.email,
		display_name: req.body.display_name,
		password: req.body.password,
		icon_uuid: req.body.icon_uuid,
		accepted_terms: true,
	};

	nm.request('POST', '/v1/users', {}, function(response) {
		res.render('pages/login', { data: {'email': body.email} });
	}, function(error) {
		console.log(error);
		res.redirect('/');
	}, JSON.stringify(body));
});
