import { test, expect } from '@playwright/test';
import { seedSession, removeSeeded } from './fixtures';

const adminURL = process.env.E2E_ADMIN_URL ?? `http://localhost:${process.env.ADMIN_PORT ?? '8002'}`;
const adminEmail = process.env.E2E_ADMIN_EMAIL ?? 'e2e-admin@example.com';

test('public deployments never expose administrator routes', async ({ request, baseURL }) => {
 const storeURL = process.env.E2E_BOOKSTORE_URL ?? `http://localhost:${process.env.BOOKSTORE_PORT ?? '8001'}`;
 for (const origin of [baseURL!, storeURL]) {
  for (const path of ['/api/admin/users', '/api/admin/healthz']) {
   expect((await request.get(origin + path)).status()).toBe(404);
  }
 }
});

test('admin deployment rejects public tokens and reader sessions', async ({ request }) => {
 const reader = seedSession('E2E 독자', { audience: 'admin' });
 const publicAdmin = seedSession('E2E 공개 관리자', { email: adminEmail });
 try {
  for (const cookie of [`nn_session=${publicAdmin.token}`, `nn_admin_session=${publicAdmin.token}`]) {
   expect((await request.get(adminURL + '/api/admin/users', { headers: { Cookie: cookie } })).status()).toBe(401);
  }
  expect((await request.get(adminURL + '/api/admin/users', { headers: { Cookie: `nn_admin_session=${reader.token}` } })).status()).toBe(403);
  for (const path of ['/api/studio/healthz', '/api/store/healthz']) {
   expect((await request.get(adminURL + path)).status()).toBe(404);
  }
 } finally {
  removeSeeded(reader.sub);
  removeSeeded(publicAdmin.sub);
 }
});

test('admin SPA uses its own session and logout keeps the public session', async ({ page, context, request }) => {
 const admin = seedSession('E2E 관리자', { audience: 'admin', email: adminEmail });
 const publicSession = seedSession('E2E 공개 관리자', { email: adminEmail });
 try {
  await context.addCookies([
   { name: 'nn_admin_session', value: admin.token, url: adminURL, httpOnly: true, sameSite: 'Lax' },
   { name: 'nn_session', value: publicSession.token, url: adminURL, httpOnly: true, sameSite: 'Lax' }
  ]);
  const users = await request.get(adminURL + '/api/admin/users', { headers: { Cookie: `nn_admin_session=${admin.token}` } });
  expect(users.status()).toBe(200);
  expect(await users.json()).toEqual({ users: [] });
  await page.goto(adminURL);
  await expect(page.getByRole('heading', { name: '관리자 홈' })).toBeVisible();
  await expect(page.getByText('관리자 연결: 정상')).toBeVisible();
  await page.getByRole('button', { name: /E2E 관리자/ }).click();
  await page.getByRole('menuitem', { name: '로그아웃' }).click();
  await expect(page).toHaveURL(/\/login$/);
  expect((await request.get(adminURL + '/api/auth/me', { headers: { Cookie: `nn_admin_session=${admin.token}` } })).status()).toBe(401);
  const studioURL = process.env.E2E_BASE_URL ?? `http://localhost:${process.env.WEB_PORT ?? '8000'}`;
  expect((await request.get(studioURL + '/api/auth/me', { headers: { Cookie: `nn_session=${publicSession.token}` } })).status()).toBe(200);
 } finally {
  removeSeeded(admin.sub);
  removeSeeded(publicSession.sub);
 }
});
