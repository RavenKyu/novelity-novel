import { test, expect } from '@playwright/test';

test('anonymous visitor is sent to /login', async ({ page }) => {
	await page.goto('/');
	await expect(page).toHaveURL(/\/login$/);
	await expect(page.getByRole('heading', { name: 'Novelity Admin' })).toBeVisible();
	await expect(page.getByRole('link', { name: 'Google로 로그인' })).toHaveAttribute(
		'href',
		'/api/auth/google/login'
	);
});

test('login error code from the OAuth callback is shown in Korean', async ({ page }) => {
	await page.goto('/login?error=unverified_email');
	await expect(page.getByText('이메일이 확인된 Google 계정만 사용할 수 있습니다.')).toBeVisible();
});

test('unknown error code falls back to the generic message', async ({ page }) => {
	await page.goto('/login?error=whatever');
	await expect(page.getByText('서버 오류로 로그인하지 못했습니다.', { exact: false })).toBeVisible();
});

test('API rejects requests without a session', async ({ request }) => {
	const res = await request.get('/api/auth/me');
	expect(res.status()).toBe(401);
});
