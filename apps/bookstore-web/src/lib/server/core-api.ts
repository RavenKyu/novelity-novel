// This helper owns the only credential-bearing server transport. Neither the
// request Host nor a caller-supplied URL can select its upstream origin.
export interface ApiConfig {
	apiOrigin: string;
	bookstoreOrigin: string;
}

export function configuredOrigin(value: string): URL {
	const url = new URL(value);
	if (
		!['http:', 'https:'].includes(url.protocol) ||
		url.username ||
		url.password ||
		url.pathname !== '/' ||
		url.search ||
		url.hash
	)
		throw new Error('Expected an HTTP origin');
	return url;
}

export function publicCookies(raw: string): string {
	return raw
		.split(';')
		.map((part) => part.trim())
		.filter((part) => {
			const name = part.slice(0, part.indexOf('='));
			return name === 'nn_session' || name === 'nn_oauth';
		})
		.join('; ');
}

export function isPublicApiPath(path: string): boolean {
	return (
		path === '/api/healthz' ||
		path.startsWith('/api/store/') ||
		[
			'/api/auth/me',
			'/api/auth/logout',
			'/api/auth/google/login',
			'/api/auth/google/callback'
		].includes(path)
	);
}

export async function coreApiRequest(
	config: ApiConfig,
	path: string,
	request: Request,
	init: RequestInit = {},
	fetcher: typeof fetch = fetch
): Promise<Response> {
	const origin = configuredOrigin(config.apiOrigin);
	const bookstore = configuredOrigin(config.bookstoreOrigin);
	const target = new URL(path, origin);
	if (
		!path.startsWith('/api/') ||
		path.includes('\\') ||
		target.origin !== origin.origin ||
		!isPublicApiPath(target.pathname)
	)
		throw new Error('Untrusted API target');
	const headers = new Headers(init.headers);
	// Remove ambient credentials; only the public cookies from this request are sent.
	headers.delete('authorization');
	headers.delete('cookie');
	headers.set('host', bookstore.host);
	const cookie = publicCookies(request.headers.get('cookie') ?? '');
	if (cookie) headers.set('cookie', cookie);
	for (const name of ['accept', 'content-type', 'origin', 'referer', 'sec-fetch-site']) {
		const value = request.headers.get(name);
		if (value) headers.set(name, value);
	}
	return fetcher(target, {
		...init,
		headers,
		redirect: 'manual',
		signal: AbortSignal.timeout(5000)
	});
}
