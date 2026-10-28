const { test, before, after } = require('node:test');
const assert = require('node:assert');
const http = require('node:http');

// A fake API that records what the webapp forwards to it.
let lastRequest = null;
let fakeApiHandler = null;
const fakeApi = http.createServer((req, res) => {
	let body = '';
	req.on('data', (chunk) => body += chunk);
	req.on('end', () => {
		lastRequest = { method: req.method, url: req.url, headers: req.headers, body: body ? JSON.parse(body) : null };
		fakeApiHandler(req, res, lastRequest);
	});
});

function json(res, status, data) {
	res.writeHead(status, { 'Content-Type': 'application/json' });
	res.end(JSON.stringify(data));
}

let server;
let base;

before(async () => {
	await new Promise((resolve) => fakeApi.listen(0, '127.0.0.1', resolve));
	process.env.NODE_ENV = 'test';
	process.env.API_ENDPOINT = `http://127.0.0.1:${fakeApi.address().port}`;
	process.env.CLIENT_ID = 'test-client';
	process.env.CLIENT_SECRET = 'abc.secret';
	const app = require('../server');
	server = app.listen(0, '127.0.0.1');
	await new Promise((resolve) => server.once('listening', resolve));
	base = `http://127.0.0.1:${server.address().port}`;
});

after(() => {
	server.close();
	fakeApi.close();
});

test('root renders the single-page app', async () => {
	const res = await fetch(base + '/');
	assert.strictEqual(res.status, 200);
	const html = await res.text();
	assert.match(html, /src="\/js\/app.js"/);
	assert.match(html, /new App\("sv", ""/);
});

test('client-side routes render the app with a redirect', async () => {
	const res = await fetch(base + '/checklists?lang=en&x=1');
	assert.strictEqual(res.status, 200);
	const html = await res.text();
	assert.match(html, /new App\("en", "checklists"/);
	assert.match(html, /"x":"1"/);
});

test('list route passes the uuid to the app', async () => {
	const html = await (await fetch(base + '/list/abc123')).text();
	assert.match(html, /"list"/);
	assert.match(html, /"uuid":"abc123"/);
});

test('terms and privacy policy render', async () => {
	assert.strictEqual((await fetch(base + '/terms')).status, 200);
	assert.strictEqual((await fetch(base + '/privacy-policy')).status, 200);
});

test('unknown routes are 404', async () => {
	assert.strictEqual((await fetch(base + '/nope')).status, 404);
});

test('compare without uuids redirects home', async () => {
	const res = await fetch(base + '/compare', { redirect: 'manual' });
	assert.strictEqual(res.status, 302);
});

test('compare fetches from the API and escapes user content', async () => {
	fakeApiHandler = (req, res) => json(res, 200, {
		reference_list: { name: { sv: 'Lista' } },
		species: [{ uuid: 's1', name: { sv: 'Talgoxe', la: 'Parus major' } }],
		result: [{ user: { display_name: '<script>alert(1)</script>' }, reports: [{ species: { uuid: 's1' } }] }]
	});
	const res = await fetch(base + '/compare?uuids=a,b');
	assert.strictEqual(res.status, 200);
	assert.deepStrictEqual(lastRequest.body, { uuids: ['a', 'b'] });
	assert.strictEqual(lastRequest.url, '/v1/lists/compare');
	const html = await res.text();
	assert.match(html, /Talgoxe/);
	assert.doesNotMatch(html, /<script>alert/);
	assert.match(html, /&lt;script&gt;/);
});

test('compare redirects when the API fails', async () => {
	fakeApiHandler = (req, res) => json(res, 400, { message: 'bad' });
	const res = await fetch(base + '/compare?uuids=a', { redirect: 'manual' });
	assert.strictEqual(res.status, 302);
});

test('auth adds the client credentials server-side', async () => {
	fakeApiHandler = (req, res) => json(res, 200, { access_token: 't' });
	const res = await fetch(base + '/api/v1/auth', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ grant_type: 'password', email: 'a@b.se', password: 'pw', client_id: 'evil' })
	});
	assert.strictEqual(res.status, 200);
	assert.deepStrictEqual(await res.json(), { access_token: 't' });
	assert.strictEqual(lastRequest.url, '/v1/auth');
	assert.strictEqual(lastRequest.body.client_id, 'test-client');
	assert.strictEqual(lastRequest.body.client_secret, 'abc.secret');
	assert.strictEqual(lastRequest.body.email, 'a@b.se');
});

test('api proxy forwards method, path, auth header and status', async () => {
	fakeApiHandler = (req, res) => json(res, 401, { code: 7001 });
	const res = await fetch(base + '/api/v1/lists/list?x=1', { headers: { Authorization: 'Bearer abc' } });
	assert.strictEqual(res.status, 401);
	assert.deepStrictEqual(await res.json(), { code: 7001 });
	assert.strictEqual(lastRequest.method, 'GET');
	assert.strictEqual(lastRequest.url, '/v1/lists/list?x=1');
	assert.strictEqual(lastRequest.headers.authorization, 'Bearer abc');
});

test('api proxy forwards bodies on writes', async () => {
	fakeApiHandler = (req, res) => json(res, 200, { ok: true });
	await fetch(base + '/api/v1/tags', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ name: 'garden' })
	});
	assert.strictEqual(lastRequest.method, 'POST');
	assert.deepStrictEqual(lastRequest.body, { name: 'garden' });
});

test('api proxy returns 503 when the API is unreachable', async () => {
	const saved = fakeApiHandler;
	fakeApiHandler = (req, res) => req.socket.destroy();
	const res = await fetch(base + '/api/v1/lists/list');
	assert.strictEqual(res.status, 503);
	fakeApiHandler = saved;
});
