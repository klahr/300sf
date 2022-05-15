var express        = require('express');
var cookieParser   = require('cookie-parser')
var session        = require('express-session');
var expressLayouts = require('express-ejs-layouts');
var bodyParser     = require('body-parser');
var i18n           = require("i18n");
var app            = express();
var port           = 8080;

i18n.configure({
	locales: ['en', 'sv'],
	cookie: "lang",
	queryParameter: "lang",
	updateFiles: false,
	directory: __dirname + '/locales'
});

app.set('view engine', 'ejs');
app.use(expressLayouts);
app.use(cookieParser());
app.use(i18n.init);
app.use(bodyParser.urlencoded({ extended: true }));
app.use(session({ secret: 'keyboard cat', cookie: { maxAge: 60000 * 60 * 48 }}))

var router = require('./app/routes');
app.use('/', router);

app.use(express.static(__dirname + '/public'));

app.listen(port, function() {
	console.log('300sf web started');
});
