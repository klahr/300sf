const express = require('express');
const expressLayouts = require('express-ejs-layouts');
const i18n = require('i18n');
const compression = require('compression');
const morgan = require('morgan');
const app = express();
const port = process.env.PORT || 3000;

i18n.configure({
	locales: ['en', 'sv', 'de'],
	defaultLocale: 'sv',
	queryParameter: 'lang',
	updateFiles: false,
	directory: __dirname + '/translations/locale'
});

app.set('view engine', 'ejs');
app.set('views', __dirname + '/views');
app.use(i18n.init);
app.use(compression());
app.use(express.json());
app.use(expressLayouts);
if (process.env.NODE_ENV !== 'test') {
	app.use(morgan('dev'));
}

app.use('/', require('./routes/app'));
app.use(express.static(__dirname + '/public'));

module.exports = app;

if (require.main === module) {
	app.listen(port, function () {
		console.log(`listening on port ${port}`);
	});
}
