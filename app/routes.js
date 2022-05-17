var express = require('express');
const fileUpload = require('express-fileupload');
var http    = require('http');
const { stringify } = require('querystring');

var router = express.Router();
module.exports = router;

router.get('/', function(req, res) {
	res.render('pages/landing-page');
});
