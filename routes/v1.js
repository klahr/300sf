const express = require('express');
const router = express.Router();
const nm = require('./network-manager');
const config = require('./config');
module.exports = router;

router.post('/auth', function(req, res) {
	const post = req.body;
	const email = post.email;
	const password = post.password;
	const authorization = 'Basic ' + Buffer.from(email + ':' + password).toString('base64');
	const stayLoggedIn = post.stay_logged_in === 'on';
	const body = {
		'refresh-token': ''
	};

	nm.request('POST', '/v1/auth', {'Authorization': authorization}, function(response) {
		console.log(response);
		if (stayLoggedIn) {
			res.cookie('x300sf_at', response.access_token, {
				maxAge: response.max_age
			});
		} else {
			res.cookie('x300sf_at', response.access_token);
		}
		res.redirect(post.success);
	}, function(error) {
		console.log(error);
	}, JSON.stringify(body));
});
