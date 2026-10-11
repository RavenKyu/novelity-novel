import { test, expect, seedSession, removeSeeded } from './fixtures';

const bookstoreURL = process.env.E2E_BOOKSTORE_URL ??
	`http://localhost:${process.env.BOOKSTORE_PORT ?? '8001'}`;

test('public bookstore and metadata render without JavaScript', async ({ browser }) => {
	const context = await browser.newContext({ javaScriptEnabled: false });
	try {
		const page = await context.newPage();
		await page.goto(bookstoreURL);
		await expect(page.getByRole('heading', { name: '완성된 이야기의 다음 공간, 서점' })).toBeVisible();
		await expect(page.getByRole('heading', { name: '연재', exact: true })).toBeVisible();
		await expect(page.getByRole('heading', { name: '책', exact: true })).toBeVisible();
		await expect(page.getByRole('link', { name: '로그인', exact: true })).toBeVisible();
		await expect(page.locator('meta[property="og:title"]')).toHaveAttribute('content', /Novelity 서점/);
		await expect(page.locator('link[rel="canonical"]')).toHaveAttribute('href', bookstoreURL + '/');
		const session = await context.request.get(bookstoreURL + '/api/store/session');
		expect(await session.json()).toEqual({ user: null });
		expect((await context.request.get(bookstoreURL + '/api/admin/healthz')).status()).toBe(404);
	} finally { await context.close(); }
});

test('shared public session renders in HTML and logout clears it', async ({ context, page, user }) => {
	// The studio fixture sets a host-only localhost cookie: ports share it in local dev.
	const html = await context.request.get(bookstoreURL);
	expect(html.headers()['cache-control']).toContain('no-store');
	expect(await html.text()).toContain(user.name);
	await page.goto(bookstoreURL);
	await expect(page.getByTestId('session-name')).toHaveText(`${user.name} 님`);
	const forbidden = await context.request.post(bookstoreURL + '/api/auth/logout', {
		headers: { Origin: 'https://untrusted.example' }
	});
	expect(forbidden.status()).toBe(403);
	await page.getByRole('button', { name: '로그아웃', exact: true }).click();
	await expect(page.getByRole('link', { name: '로그인', exact: true })).toBeVisible();
	const session = await context.request.get(bookstoreURL + '/api/store/session');
	expect(await session.json()).toEqual({ user: null });
});

test('stale cookies keep the bookstore public and login returns to the store', async ({ context, page }) => {
	const expired = seedSession('만료 세션');
	removeSeeded(expired.sub);
	await context.addCookies([{ name: 'nn_session', value: expired.token, url: bookstoreURL }]);
	await page.goto(bookstoreURL);
	await expect(page.getByRole('link', { name: '로그인', exact: true })).toBeVisible();
	await page.getByRole('link', { name: '로그인', exact: true }).click();
	await expect(page.getByRole('link', { name: 'Google로 로그인' })).toHaveAttribute('href', /\/api\/auth\/google\/login/);
	// Empty Google credentials in CI exercise the callback/error proxy without real sign-in.
	const login = await context.request.get(bookstoreURL + '/api/auth/google/login', { maxRedirects: 0 });
	expect(login.status()).toBe(302);
	expect(login.headers()['location']).toBe('/login?error=not_configured');
});
