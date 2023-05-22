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
            res.redirect('/lite');
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
            res.redirect('/lite');
        });
    } else {
        res.redirect('/lite');
    }
});

router.get('/templates/:uuid', function(req, res) {
	const accessToken = req.cookies.x300sf_at;
	if (accessToken != null) {
        const data = {
            list: null,
            species: null
        };
        nm.request('GET', '/v1/list_templates/' + req.params.uuid, {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
            data.list = response;
            nm.request('GET', '/v1/species/list', {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
                data.species = response;
                res.render('lite/pages/template-edit', { layout: 'lite-layout', data: data, lang: i18n.getLocale(req)});
            }, function(error) {
                console.log(error);
                res.redirect('/lite');
            });
        }, function(error) {
            console.log(error);
            res.redirect('/lite');
        });
    } else {
        res.redirect('/lite');
    }
});

router.get('/species/:uuid', function(req, res) {
	const accessToken = req.cookies.x300sf_at;
	if (accessToken != null) {
        const data = {
            list: null,
            species: null
        };
        nm.request('GET', '/v1/species/' + req.params.uuid, {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
            res.render('lite/pages/species', { layout: 'lite-layout', data: response, lang: i18n.getLocale(req)});
        }, function(error) {
            console.log(error);
            res.redirect('/lite');
        });
    } else {
        res.redirect('/lite');
    }
});

router.post('/list_template/:uuid/name', function(req, res) {
	const accessToken = req.cookies.x300sf_at;
	if (accessToken != null) {
        const body = {
            'name': {
                'en': req.body.name_en,
                'sv': req.body.name_sv
            }
        };
        nm.request('POST', '/v1/list_templates/' + req.params.uuid + '/name', {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
            res.redirect('/lite/templates/' + req.params.uuid);
        }, function(error) {
            console.log(error);
            res.redirect('/lite/templates/' + req.params.uuid);
        }, JSON.stringify(body));
    } else {
        res.redirect('/lite');
    }
});

router.post('/list_template/:uuid/description', function(req, res) {
	const accessToken = req.cookies.x300sf_at;
	if (accessToken != null) {
        const body = {
            'description': {
                'en': req.body.description_en,
                'sv': req.body.description_sv
            }
        };
        nm.request('POST', '/v1/list_templates/' + req.params.uuid + '/description', {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
            res.redirect('/lite/templates/' + req.params.uuid);
        }, function(error) {
            console.log(error);
            res.redirect('/lite/templates/' + req.params.uuid);
        }, JSON.stringify(body));
    } else {
        res.redirect('/lite');
    }
});

router.post('/list_template/:uuid/archive', function(req, res) {
	const accessToken = req.cookies.x300sf_at;
	if (accessToken != null) {
        nm.request('POST', '/v1/list_templates/' + req.params.uuid + '/archive', {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
            res.redirect('/lite/templates');
        }, function(error) {
            console.log(error);
            res.redirect('/lite/templates');
        });
    } else {
        res.redirect('/lite');
    }
});

router.post('/list_template/:uuid/add_species', function(req, res) {
	const accessToken = req.cookies.x300sf_at;
	if (accessToken != null) {
        const body = {
            'species_uuid': req.body.species_id
        };
        nm.request('POST', '/v1/list_templates/' + req.params.uuid + '/add_species', {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
            res.redirect('/lite/templates/' + req.params.uuid);
        }, function(error) {
            console.log(error);
            res.redirect('/lite/templates/' + req.params.uuid);
        }, JSON.stringify(body));
    } else {
        res.redirect('/lite');
    }
});

router.post('/list_template/:uuid/remove_species', function(req, res) {
	const accessToken = req.cookies.x300sf_at;
	if (accessToken != null) {
        const body = {
            'species_uuid': req.body.species_id
        };
        nm.request('POST', '/v1/list_templates/' + req.params.uuid + '/remove_species', {'X-Api-Key': config.ApiKey, 'X-Access-Token': req.cookies.x300sf_at}, function(response) {
            res.redirect('/lite/templates/' + req.params.uuid);
        }, function(error) {
            console.log(error);
            res.redirect('/lite/templates/' + req.params.uuid);
        }, JSON.stringify(body));
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
            res.redirect('/lite/templates');
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
            res.redirect('/api');
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
                res.redirect('/api');
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
            res.redirect('/api');
        });
    } else {
        res.redirect('/lite');
    }
});
