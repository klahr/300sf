var express = require('express');
const fileUpload = require('express-fileupload');
var http    = require('http');
const { stringify } = require('querystring');
var nm = require('./network_manager');

var router = express.Router();
module.exports = router;

router.get('/', function(req, res) {
	res.render('pages/landing-page');
});

router.get('/login', function(req, res) {
	res.render('pages/login', {message: "", error: ""});
});

router.get('/signup', function(req, res) {
	res.render('pages/signup', {message: "", error: ""});
});

