const express = require('express');
const expressLayouts = require('express-ejs-layouts');
const expressSession = require('express-session');
const bodyParser = require('body-parser');
const cookieParser = require('cookie-parser');
const i18n = require('i18n');
const app = express();
const port = 3000;

i18n.configure({
	locales: ['en', 'sv'],
	cookie: 'lang',
	queryParameter: 'lang',
	updateFiles: false,
	directory: __dirname + '/translations/locale'
});

app.set('view engine', 'ejs');
app.use(i18n.init);
app.use(bodyParser.urlencoded({extended: true}));
app.use(expressLayouts);
app.use(cookieParser());
app.use(expressSession({
	secret: 'keyboard cat',
	cookie: {
		maxAge: 60000 * 60 * 48
	}
}));

app.use('/', require('./routes/app'));
app.use('/lite', require('./routes/lite-app'));
app.use('/v1', require('./routes/v1'));
app.use(express.static(__dirname + '/public'));

app.listen(port, function() {
	console.log(`listening on port ${port}`);
});
