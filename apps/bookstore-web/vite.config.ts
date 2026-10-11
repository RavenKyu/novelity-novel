import tailwindcss from '@tailwindcss/vite';
import adapter from '@sveltejs/adapter-node';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit({ adapter: adapter() })],
	server: { port: Number(process.env.BOOKSTORE_PORT) || 8001, strictPort: true }
});
