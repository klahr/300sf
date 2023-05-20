const express = require('express');
const router = express.Router();
const nm = require('./network-manager');
const config = require('./config');
module.exports = router;

router.get('/', function(req, res) {
	const accessToken = req.cookies.x300sf_at;
	if (accessToken != null) {
        res.render('lite/pages/dashboard', { layout: false});
    } else {
        res.render('lite/pages/home', { layout: false});
    }
});

router.get('/api', function(req, res) {
	const accessToken = req.cookies.x300sf_at;
	if (accessToken != null) {
        nm.request('GET', '/v1/api_keys/list', {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
            res.render('lite/pages/api', { layout: false, data: response});
        }, function(error) {
            console.log(error);
        });
    } else {
        res.redirect('/');
    }
});

router.post('/api/create', function(req, res) {
    const accessToken = req.cookies.x300sf_at;
	const body = {
        "name": req.body.name
    }
    if (accessToken != null) {
        nm.request('POST', '/v1/api_keys', {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
            const api_key = response.api_key;
            nm.request('GET', '/v1/api_keys/list', {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
                res.render('lite/pages/api', { layout: false, data: response, new_api_key: api_key});
            }, function(error) {
                console.log(error);
            });
        }, function(error) {
            console.log(error);
        }, JSON.stringify(body));
    } else {
        res.redirect('/');
    }
});
