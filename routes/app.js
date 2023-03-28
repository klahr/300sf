const express = require('express');
const http = require('http');
const router = express.Router();

module.exports = router;

router.get('/ping', function(req, res) {
	console.log('ping');
});

