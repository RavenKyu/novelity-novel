import tailwindcss from '@tailwindcss/vite';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import pkg from './package.json' with { type: 'json' };

export default defineConfig({
	define: {
		__APP_VERSION__: JSON.stringify(pkg.version)
	},
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			adapter: adapter({ fallback: 'index.html' })
		})
	],
	server: {
		// Same origin as the web container, so ADMIN_URL and the Google redirect URI stay valid.
		port: Number(process.env.ADMIN_PORT) || 8002,
		strictPort: true,
		// Local dev: forward API calls to core-api (nginx does this in containers).
		host: '127.0.0.1',
		proxy: { '/api': process.env.ADMIN_CORE_API_URL || 'http://localhost:8081' }
	}
});
