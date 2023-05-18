const express = require('express');
const router = express.Router();
const nm = require('./network-manager');
const config = require('./config');
module.exports = router;

router.get('/', function(req, res) {
	var accessToken = req.cookies.x300sf_at;
	if (accessToken != null) {
        res.render('lite/pages/dashboard', { layout: false});
    } else {
        res.render('lite/pages/home', { layout: false});
    }
});
