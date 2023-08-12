const express = require('express');
const router = express.Router();
const nm = require('./network-manager');
const config = require('./config');
module.exports = router;

router.get('/', function(req, res) {
	nm.request('GET', '/v1/list_templates/kjhkf4', {'X-Api-Key': config.ApiKey}, function(response) {
		console.log(response);
	}, function(error) {
		console.log(error);
	});
	res.render('pages/home');
});

router.get('/login', function(req, res) {
	res.render('pages/login');
});
