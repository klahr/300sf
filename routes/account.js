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
		res.render('pages/login', {message: "Du kan nu logga in med ditt nya konto!", error: ""});
	}, function(error) {
		var message;
		switch (error.code) {
			case -20002:
				message = "Ogiltig epost-adress!";
				break;
			case -20003:
				message = "Ogiltigt namn!";
				break;
			case -2004:
				message = "Ogiltigt lösenord!";
				break;
			default:
				message = "Ett oväntat fel har uppstått!";
				break;
		}
		res.render('pages/signup', {error: message});
	}, JSON.stringify(request));
});

router.post('/login', function(req, res) {
	const email = req.body.email;
	const password = req.body.password;

	const basicAuth = "Basic " + Buffer.from(email + ':' + password).toString('base64');
	const keepLoggedIn = req.body.keep_logged_in == "on";

	const requestData = {
		"permissions": "basic,edit,admin"
	};

	nm.request('POST', '/v1/auth', {'Authorization': basicAuth}, function(response, statusCode) {
		if (keepLoggedIn) {
			const date = new Date();
			res.cookie("sf300_access_token", response.access_token, {
				expires: new Date(date.getFullYear(), date.getMonth() + 1, date.getDate())
			});
		} else {
			res.cookie("sf300_access_token", response.access_token);
		}
		res.redirect('/');
	}, function(error, statusCode) {
		res.render('pages/login', {message: "", error: error.error});
	}, JSON.stringify(requestData));
});

router.post('/logout', function(req, res) {
	res.clearCookie("sf300_access_token");
	res.redirect('/');
});

