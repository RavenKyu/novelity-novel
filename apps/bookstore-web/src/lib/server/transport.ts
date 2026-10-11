import { request as httpRequest } from 'node:http';
import { request as httpsRequest } from 'node:https';
import { Readable } from 'node:stream';

// Node 22 fetch replaces Host with the transport host. core-api deliberately
// authenticates against the browser-facing Host, so preserve it with node:http.
// The connection (and TLS SNI) still goes to the configured API origin.
export function upstreamRequest(target: URL, init: RequestInit): Promise<Response> {
	return new Promise((resolve, reject) => {
		const send = target.protocol === 'https:' ? httpsRequest : httpRequest;
		const outgoing = send(
			target,
			{
				method: init.method ?? 'GET',
				headers: Object.fromEntries(new Headers(init.headers)),
				signal: init.signal ?? undefined,
				...(target.protocol === 'https:' ? { servername: target.hostname } : {})
			},
			(incoming) => {
				const headers = new Headers();
				for (let i = 0; i < incoming.rawHeaders.length; i += 2) {
					headers.append(incoming.rawHeaders[i], incoming.rawHeaders[i + 1]);
				}
				const status = incoming.statusCode ?? 502;
				const noBody = init.method === 'HEAD' || [204, 205, 304].includes(status);
				if (noBody) incoming.resume();
				resolve(
					new Response(noBody ? null : (Readable.toWeb(incoming) as ReadableStream<Uint8Array>), {
						status,
						headers
					})
				);
			}
		);
		outgoing.on('error', reject);
		if (init.body instanceof ArrayBuffer) outgoing.end(Buffer.from(init.body));
		else if (typeof init.body === 'string' || init.body == null) outgoing.end(init.body);
		else {
			outgoing.destroy(new Error('Unsupported upstream body'));
		}
	});
}
