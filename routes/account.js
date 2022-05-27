var express = require('express');
var nm = require('./network_manager');

var router = express.Router();
module.exports = router;

router.post('/signup', function(req, res) {
	const request = {
		"email": req.body.email,
		"display_name": req.body.display_name,
		"password": req.body.password,
		"accepted_terms": true
	};

	nm.request(nm.Method.POST, '/v1/users', {}, function(response) {
		console.log(response);
	}, function(error) {
		console.log(error);
	}, JSON.stringify(request));
});

