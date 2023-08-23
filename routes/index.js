var express = require('express');
var http    = require('http');

var router = express.Router();
module.exports = router;

router.get('/', function(req, res) {
	res.render('pages/landing-page');
});

