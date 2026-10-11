import { env } from '$env/dynamic/private';
import type { User } from '@novelity/api-client';
import { configuredOrigin, coreApiRequest } from '$lib/server/core-api';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ request }) => {
	const bookstoreOrigin = env.BOOKSTORE_URL || 'http://localhost:8001';
	const studioUrl = configuredOrigin(env.STUDIO_URL || 'http://localhost:8000').origin;
	const canonical = configuredOrigin(bookstoreOrigin).origin;
	try {
		const response = await coreApiRequest(
			{ apiOrigin: env.CORE_API_URL || 'http://localhost:8080', bookstoreOrigin },
			'/api/store/session',
			request
		);
		if (!response.ok) throw new Error('Session unavailable');
		const { user } = (await response.json()) as { user: User | null };
		return { user, sessionUnavailable: false, studioUrl, canonical };
	} catch {
		return { user: null, sessionUnavailable: true, studioUrl, canonical };
	}
};
