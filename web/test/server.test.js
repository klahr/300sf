const { test, before, after } = require('node:test');
const assert = require('node:assert');

let server;
let base;

before(async () => {
	server = require('../server').listen(0, '127.0.0.1');
	await new Promise((resolve) => server.once('listening', resolve));
	base = `http://127.0.0.1:${server.address().port}`;
});

after(() => server.close());

test('landing page renders in English and Swedish', async () => {
	const en = await fetch(base + '/?lang=en');
	assert.strictEqual(en.status, 200);
	const enHtml = await en.text();
	assert.match(enHtml, /<html lang="en">/);
	assert.match(enHtml, /<title>300 Swedish Birds<\/title>/);
	assert.match(enHtml, /Parus major/);

	const sv = await (await fetch(base + '/?lang=sv')).text();
	assert.match(sv, /<html lang="sv">/);
	assert.match(sv, /Talgoxe/);
	assert.doesNotMatch(sv, /Create a free account/);
	assert.match(sv, /app\.300sf\.se\/sign-up/);
});

test('static files are served', async () => {
	const res = await fetch(base + '/robots.txt');
	assert.strictEqual(res.status, 200);
});

test('unknown routes are 404', async () => {
	assert.strictEqual((await fetch(base + '/nope')).status, 404);
});
