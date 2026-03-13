import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

const backendTarget = process.env.VITE_BACKEND_URL ?? 'http://127.0.0.1:8080';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		proxy: {
			// Keep browser same-origin in dev by proxying API + WS calls to the Go backend.
			'/api': {
				target: backendTarget,
				changeOrigin: false,
				ws: true
			}
		}
	}
});
