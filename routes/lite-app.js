const express = require('express');
const router = express.Router();
const nm = require('./network-manager');
const config = require('./config');
const i18n = require('i18n');
module.exports = router;

router.get('/', function(req, res) {
	const accessToken = req.cookies.x300sf_at;
	if (accessToken != null) {
        res.render('lite/pages/dashboard', { layout: 'lite-layout'});
    } else {
        res.render('lite/pages/home', { layout: 'lite-layout'});
    }
});

router.post('/logout', function(req, res) {
	res.clearCookie("x300sf_at");
	res.redirect('/lite');
});

router.get('/account', function(req, res) {
	const accessToken = req.cookies.x300sf_at;
	if (accessToken != null) {
        nm.request('GET', '/v1/users/whoami', {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
            res.render('lite/pages/account', { layout: 'lite-layout', data: response});
        }, function(error) {
            console.log(error);
        });
    } else {
        res.redirect('/lite');
    }
});

router.get('/templates', function(req, res) {
	const accessToken = req.cookies.x300sf_at;
	if (accessToken != null) {
        nm.request('GET', '/v1/list_templates/list', {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
            res.render('lite/pages/templates', { layout: 'lite-layout', data: response, lang: i18n.getLocale(req)});
        }, function(error) {
            console.log(error);
        });
    } else {
        res.redirect('/lite');
    }
});

router.post('/list_template/create', function(req, res) {
    const accessToken = req.cookies.x300sf_at;
    const lang = i18n.getLocale(req)
	const body = {
        'name': {},
        'visibility': parseInt(req.body.visibility),
        'description': {}
    }
    body['name'][lang] = req.body.name;
    body['description'][lang] = req.body.description;
    if (accessToken != null) {
        nm.request('POST', '/v1/list_templates', {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
            res.redirect('/lite/templates');
        }, function(error) {
            console.log(error);
        }, JSON.stringify(body));
    } else {
        res.redirect('/lite');
    }
});

router.get('/api', function(req, res) {
	const accessToken = req.cookies.x300sf_at;
	if (accessToken != null) {
        nm.request('GET', '/v1/api_keys/list', {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
            res.render('lite/pages/api', { layout: 'lite-layout', data: response});
        }, function(error) {
            console.log(error);
        });
    } else {
        res.redirect('/lite');
    }
});

router.post('/api/create', function(req, res) {
    const accessToken = req.cookies.x300sf_at;
	const body = {
        'name': req.body.name
    }
    if (accessToken != null) {
        nm.request('POST', '/v1/api_keys', {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
            const api_key = response.api_key;
            nm.request('GET', '/v1/api_keys/list', {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
                res.render('lite/pages/api', { layout: 'lite-layout', data: response, new_api_key: api_key});
            }, function(error) {
                console.log(error);
            });
        }, function(error) {
            console.log(error);
        }, JSON.stringify(body));
    } else {
        res.redirect('/lite');
    }
});

router.post('/api/revoke', function(req, res) {
    const accessToken = req.cookies.x300sf_at;
    if (accessToken != null) {
        nm.request('POST', '/v1/api_keys/' + req.body.prefix + '/revoke', {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
            res.redirect('/lite/api');
        }, function(error) {
            console.log(error);
        });
    } else {
        res.redirect('/lite');
    }
});
