// Settings come from routes/config.js (see config.js.template) when present,
// and can be overridden with environment variables.
let file = {};
try {
	file = require('./config');
} catch (e) {
	if (e.code !== 'MODULE_NOT_FOUND') {
		throw e;
	}
}

module.exports = {
	ApiEndpoint: process.env.API_ENDPOINT || file.ApiEndpoint || 'http://localhost:8090',
	AnalyzeEndpoint: process.env.ANALYZE_ENDPOINT || file.AnalyzeEndpoint || 'https://api.300sf.se/analyze',
	ClientId: process.env.CLIENT_ID || file.ClientId || '',
	ClientSecret: process.env.CLIENT_SECRET || file.ClientSecret || ''
};
