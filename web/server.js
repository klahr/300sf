const express = require('express');
const expressLayouts = require('express-ejs-layouts');
const i18n = require('i18n');
const app = express();
const port = process.env.PORT || 8080;

i18n.configure({
	locales: ['en', 'sv'],
	queryParameter: 'lang',
	updateFiles: false,
	directory: __dirname + '/translations/locale'
});

app.set('view engine', 'ejs');
app.set('views', __dirname + '/views');
app.use(expressLayouts);
app.use(i18n.init);

app.use('/', require('./routes/index'));
app.use(express.static(__dirname + '/public'));

module.exports = app;

if (require.main === module) {
	app.listen(port, function () {
		console.log('300sf web started');
	});
}
