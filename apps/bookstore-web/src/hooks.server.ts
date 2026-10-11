import { env } from '$env/dynamic/private';
import type { Handle } from '@sveltejs/kit';
import { configuredOrigin, coreApiRequest, isPublicApiPath } from '$lib/server/core-api';

// The same proxy runs in Vite and adapter-node, including OAuth Set-Cookie and
// Location. Redirects are returned to the browser, never followed with cookies.
export const handle: Handle = async ({ event, resolve }) => {
	if (event.url.pathname === '/api' || event.url.pathname.startsWith('/api/')) {
		if (!isPublicApiPath(event.url.pathname)) return new Response(null, { status: 404 });
		const bookstoreOrigin = env.BOOKSTORE_URL || 'http://localhost:8001';
		if (event.url.origin !== configuredOrigin(bookstoreOrigin).origin) {
			return new Response(null, { status: 400 });
		}
		try {
			const method = event.request.method;
			const body = ['GET', 'HEAD'].includes(method) ? undefined : await event.request.arrayBuffer();
			if (body && body.byteLength > 512 * 1024) return new Response(null, { status: 413 });
			const upstream = await coreApiRequest(
				{ apiOrigin: env.CORE_API_URL || 'http://localhost:8080', bookstoreOrigin },
				event.url.pathname + event.url.search,
				event.request,
				{ method, body }
			);
			const headers = new Headers();
			for (const name of ['content-type', 'location']) {
				const value = upstream.headers.get(name);
				if (value) headers.set(name, value);
			}
			for (const cookie of upstream.headers.getSetCookie()) headers.append('set-cookie', cookie);
			headers.set('cache-control', 'private, no-store');
			return new Response(upstream.body, { status: upstream.status, headers });
		} catch {
			return new Response('API 연결을 확인할 수 없습니다.', { status: 502 });
		}
	}
	const response = await resolve(event);
	// All pages carry per-request session state in their server-rendered header.
	response.headers.set('cache-control', 'private, no-store');
	return response;
};
