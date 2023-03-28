const express = require('express');
const http = require('http');
const router = express.Router();

module.exports = router;

router.get('/', function(req, res) {
	res.render('pages/home');
});

