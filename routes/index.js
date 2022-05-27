var express = require('express');
const fileUpload = require('express-fileupload');
var http    = require('http');
const { stringify } = require('querystring');
var nm = require('./network_manager');

var router = express.Router();
module.exports = router;

router.get('/', function(req, res) {
	const accessToken = req.cookies.sf300_access_token;
	if (accessToken != null) {
		res.render('pages/dashboard');
		// postRequest('GET', '/v1/page/dashboard', {"X-Access-Token":accessToken}, function(response) {
		// 	res.render('pages/dashboard', { data: response });
		// }, function(error) {
		// 	console.log(error.description);
		// 	res.render('pages/home');
		// });
	} else {
		res.render('pages/landing-page');
	}
});

router.get('/login', function(req, res) {
	res.render('pages/login', {message: "", error: ""});
});

router.get('/signup', function(req, res) {
	res.render('pages/signup', {message: "", error: ""});
});

