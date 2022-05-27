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

