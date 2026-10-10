import { defineConfig, devices } from '@playwright/test';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));

/**
 * Playwright E2E harness against the running app.
 *
 * - Targets E2E_BASE_URL (default http://localhost:${WEB_PORT:-8000}), which is
 *   either the dev servers (`npm run dev`) or the containerized app
 *   (`docker compose --profile app up -d --build`, as in CI). global-setup fails
 *   fast if /api/healthz is not 200, instead of every spec timing out.
 * - Workers fixed to 1: specs share one MongoDB and seed/clean their own users.
 */
const baseURL = process.env.E2E_BASE_URL ?? `http://localhost:${process.env.WEB_PORT ?? '8000'}`;

export default defineConfig({
	testDir: resolve(__dirname),
	testMatch: /.*\.spec\.ts/,
	outputDir: resolve(__dirname, '../../artifacts/e2e'),
	globalSetup: resolve(__dirname, 'global-setup.ts'),
	timeout: 30_000,
	expect: { timeout: 10_000 },
	workers: 1,
	retries: 1,
	fullyParallel: false,
	reporter: process.env.CI ? [['list'], ['html', { open: 'never', outputFolder: resolve(__dirname, '../../artifacts/e2e-report') }]] : [['list']],
	use: {
		baseURL,
		trace: 'retain-on-failure',
		screenshot: 'only-on-failure',
		locale: 'ko-KR'
	},
	projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }]
});
