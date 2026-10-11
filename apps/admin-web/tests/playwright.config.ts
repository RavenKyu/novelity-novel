import { defineConfig, devices } from '@playwright/test';

const port = 18002;
const baseURL = `http://127.0.0.1:${port}`;
export default defineConfig({
	testDir: '.',
	testMatch: '*.spec.ts',
	outputDir: '../../../artifacts/admin-web',
	workers: 1,
	use: { baseURL, ...devices['Desktop Chrome'], trace: 'retain-on-failure' },
	webServer: {
		command: `npm run dev -- --port ${port}`,
		url: `${baseURL}/login`,
		reuseExistingServer: false,
		env: { ADMIN_PORT: String(port) }
	}
});
