import { test, expect, type Page } from '@playwright/test';

const admin = {
	id: 'admin',
	name: '운영자',
	email: 'admin@example.com',
	picture: '',
	roles: ['reader', 'admin']
};
async function session(page: Page, status = 401, body: unknown = { error: 'unauthorized' }) {
	await page.route('**/api/auth/me', (route) => route.fulfill({ status, json: body }));
}

test('anonymous visitor keeps returnTo and cannot see the shell', async ({ page }) => {
	await session(page);
	await page.goto('/?tab=users#list');
	await expect(page).toHaveURL(/\/login\?returnTo=/);
	await expect(page.getByRole('heading', { name: 'Novelity Admin' })).toBeVisible();
	await expect(page.getByRole('link', { name: 'Google로 로그인' })).toHaveAttribute(
		'href',
		'/api/auth/google/login?returnTo=%2F%3Ftab%3Dusers%23list'
	);
	await expect(page.getByRole('heading', { name: '관리자 홈' })).toHaveCount(0);
});

for (const status of [200, 403]) {
	test(`reader or revoked admin cannot render the shell (${status})`, async ({ page }) => {
		await session(
			page,
			status,
			status === 200 ? { ...admin, roles: ['reader'] } : { error: 'forbidden' }
		);
		await page.goto('/');
		await expect(page).toHaveURL(/error=forbidden/);
		await expect(page.getByText('관리자 권한이 없습니다.', { exact: false })).toBeVisible();
		await expect(page.getByRole('heading', { name: '관리자 홈' })).toHaveCount(0);
	});
}

test('admin sees shared shell and can log out of admin session', async ({ page }) => {
	let signedIn = true;
	await page.route('**/api/auth/me', (route) =>
		route.fulfill({
			status: signedIn ? 200 : 401,
			json: signedIn ? admin : { error: 'unauthorized' }
		})
	);
	await page.route('**/api/admin/users', (route) => route.fulfill({ json: { users: [] } }));
	await page.route('**/api/auth/logout', (route) => {
		expect(route.request().method()).toBe('POST');
		signedIn = false;
		return route.fulfill({ status: 204 });
	});
	await page.goto('/login');
	await expect(page.getByRole('heading', { name: '관리자 홈' })).toBeVisible();
	await expect(page.getByText('관리자 연결: 정상')).toBeVisible();
	await page.getByRole('button', { name: /운영자/ }).click();
	await page.getByRole('menuitem', { name: '로그아웃' }).click();
	await expect(page).toHaveURL(/\/login$/);
	await expect(page.getByRole('heading', { name: '관리자 홈' })).toHaveCount(0);
});

test('expired admin endpoint rechecks session and returns to login', async ({ page }) => {
	let checks = 0;
	await page.route('**/api/auth/me', (route) =>
		route.fulfill({
			status: ++checks === 1 ? 200 : 401,
			json: checks === 1 ? admin : { error: 'unauthorized' }
		})
	);
	await page.route('**/api/admin/users', (route) =>
		route.fulfill({ status: 401, json: { error: 'unauthorized' } })
	);
	await page.goto('/');
	await expect(page).toHaveURL(/\/login$/);
	await expect(page.getByRole('heading', { name: '관리자 홈' })).toHaveCount(0);
});

test('login returnTo cannot navigate to a public or foreign origin', async ({ page }) => {
	await session(page, 200, admin);
	await page.route('**/api/admin/users', (route) => route.fulfill({ json: { users: [] } }));
	await page.goto('/login?returnTo=https%3A%2F%2Fnovelity.kr%2F');
	await expect(page).toHaveURL(/127\.0\.0\.1:18002\/$/);
});

test('session lookup failure leaves login available', async ({ page }) => {
	await session(page, 503, { error: 'unavailable' });
	await page.goto('/');
	await expect(page.getByRole('link', { name: 'Google로 로그인' })).toBeVisible();
	await expect(page.getByRole('heading', { name: '관리자 홈' })).toHaveCount(0);
});
