const express = require('express');
const expressLayouts = require('express-ejs-layouts');
const bodyParser = require('body-parser');
const i18n = require('i18n');
const app = express();
const port = 3000;

i18n.configure({
	locales: ['en', 'sv'],
	cookie: 'lang',
	queryParameter: 'lang',
	updateFiles: false,
	directory: __dirname + '/locales'
});

app.set('view engine', 'ejs');
app.use(i18n.init);
app.use(bodyParser.urlencoded({extended: true}));
app.use(expressLayouts);

const router = require('./routes/app'); 
app.use('/', router);
app.use(express.static(__dirname + '/public'));

app.listen(port, function() {
	console.log(`listening on port ${port}`);
});

