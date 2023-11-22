const express = require('express');
const expressLayouts = require('express-ejs-layouts');
const bodyParser = require('body-parser');
const i18n = require('i18n');
const compression = require('compression');
const morgan = require('morgan');
const app = express();
const port = 3000;

i18n.configure({
	locales: ['en', 'sv', 'de'],
	defaultLocale: 'sv',
	queryParameter: 'lang',
	updateFiles: false,
	directory: __dirname + '/translations/locale'
});

app.set('view engine', 'ejs');
app.use(i18n.init);
app.use(compression());
app.use(bodyParser.json());
app.use(expressLayouts);
app.use(morgan('dev'));

app.use('/', require('./routes/app'));
app.use('/lite', require('./routes/lite-app'));
app.use('/v1', require('./routes/v1'));
app.use(express.static(__dirname + '/public'));

app.listen(port, function () {
	console.log(`listening on port ${port}`);
});
