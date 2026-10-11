import assert from 'node:assert/strict';
import { test } from 'node:test';
import { coreApiRequest, configuredOrigin, publicCookies } from './core-api.ts';

const config = { apiOrigin: 'http://core-api:8080', bookstoreOrigin: 'https://novelity.kr' };

test('only the configured API gets public cookies, with the configured bookstore Host', async () => {
	let calls = 0;
	const fetcher: typeof fetch = async (input, init) => {
		calls++;
		assert.equal(String(input), 'http://core-api:8080/api/store/session');
		const headers = new Headers(init?.headers);
		assert.equal(headers.get('host'), 'novelity.kr');
		assert.equal(headers.get('cookie'), 'nn_session=public; nn_oauth=flow');
		assert.equal(headers.get('authorization'), null);
		assert.equal(init?.redirect, 'manual');
		return new Response(null, { status: 302, headers: { location: 'https://evil.example' } });
	};
	const request = new Request('https://attacker.example/', {
		headers: {
			cookie: 'nn_admin_session=admin; nn_session=public; unrelated=secret; nn_oauth=flow'
		}
	});
	const response = await coreApiRequest(
		config,
		'/api/store/session',
		request,
		{
			headers: { cookie: 'injected=secret', authorization: 'Bearer injected' }
		},
		fetcher
	);
	assert.equal(response.status, 302);
	assert.equal(calls, 1);
});

test('arbitrary origins, traversal and private surfaces are rejected before fetch', async () => {
	for (const path of [
		'https://evil.example/api/store/session',
		'//evil.example/api/store/session',
		'/api/store/../../admin',
		'/api/store/../admin/healthz',
		'/api/admin/healthz',
		'/api/studio/healthz',
		'/api/store/..\\admin'
	]) {
		await assert.rejects(
			coreApiRequest(config, path, new Request('https://novelity.kr/'), {}, async () => {
				assert.fail('credentials escaped their API boundary');
			})
		);
	}
});

test('CSRF origin evidence survives proxying', async () => {
	const request = new Request('https://novelity.kr/api/auth/logout', {
		method: 'POST',
		headers: { origin: 'https://studio.novelity.kr', 'sec-fetch-site': 'same-site' }
	});
	await coreApiRequest(config, '/api/auth/logout', request, { method: 'POST' }, async (_, init) => {
		const headers = new Headers(init?.headers);
		assert.equal(headers.get('origin'), 'https://studio.novelity.kr');
		assert.equal(headers.get('sec-fetch-site'), 'same-site');
		assert.equal(headers.get('cookie'), null);
		return new Response(null, { status: 403 });
	});
});

test('configuration accepts origins only, and cookie names match exactly', () => {
	for (const value of [
		'file:///tmp/api',
		'http://user:pass@api',
		'https://api/path',
		'https://api/?q=1'
	]) {
		assert.throws(() => configuredOrigin(value));
	}
	assert.equal(publicCookies('nn_session_extra=x; nn_admin_oauth=a; cookie=x'), '');
});
