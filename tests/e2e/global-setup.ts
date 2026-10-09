import type { FullConfig } from '@playwright/test';

// Fail once with the fix, rather than letting every spec time out.
export default async function globalSetup(config: FullConfig): Promise<void> {
	const baseURL = config.projects[0].use.baseURL!;
	let status = 0;
	try {
		status = (await fetch(`${baseURL}/api/healthz`)).status;
	} catch {
		// unreachable: status stays 0
	}
	if (status !== 200) {
		throw new Error(
			`e2e: ${baseURL}/api/healthz returned ${status || 'no response'}. ` +
				'Start the stack first: docker compose up -d --build'
		);
	}
}
