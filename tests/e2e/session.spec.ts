import { test, expect } from './fixtures';

test('signed-in user sees the admin shell and API health', async ({ page, user }) => {
	await page.goto('/');
	await expect(page.locator('aside').getByText(user.name)).toBeVisible();
	await expect(page.locator('aside').getByText(user.email)).toBeVisible();
	await expect(page.getByText('정상')).toBeVisible();
});

test('signed-in user visiting /login is sent home', async ({ page, user }) => {
	await page.goto('/login');
	await expect(page).toHaveURL(/\/$/);
	await expect(page.locator('aside').getByText(user.name)).toBeVisible();
});

test('logout ends the session', async ({ page, user }) => {
	await page.goto('/');
	await page.locator('aside').getByText(user.name).click();
	await page.getByRole('menuitem', { name: '로그아웃' }).click();

	await expect(page).toHaveURL(/\/login$/);
	expect((await page.request.get('/api/auth/me')).status()).toBe(401);
});
