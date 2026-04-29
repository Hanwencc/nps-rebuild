import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue(), tailwindcss()],
  // The SPA is mounted under /ui/ by the Go server (see
  // web/webui/embed.go). All built asset URLs must therefore be
  // prefixed with /ui/ so <script src="/ui/assets/...">. The dev
  // server still works because Vite rewrites it transparently.
  base: '/ui/',
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  // The compiled SPA is embedded by the Go binary at web/webui/dist
  // (see web/webui/embed.go). Output relative to the repo root so
  // `yarn build` populates the embed FS in one step.
  build: {
    outDir: '../web/webui/dist',
    emptyOutDir: true,
    sourcemap: false,
    target: 'es2020',
  },
  server: {
    // Listen on all interfaces (incl. IPv6 ::1) so the browser can
    // reach the dev server via either `localhost` or `127.0.0.1`.
    // Windows often resolves `localhost` to ::1, which a default
    // 127.0.0.1-only bind would refuse with ERR_CONNECTION_REFUSED.
    host: true,
    port: 5173,
    strictPort: false,
    // Dev mode: forward /api to the running Go backend (default 8081).
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8081',
        changeOrigin: true,
      },
    },
  },
})
