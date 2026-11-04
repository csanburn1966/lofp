import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  // Removed the base property completely so it runs from the root of the subdomain
  plugins: [react(), tailwindcss()],
  server: {
    port: 4992,
    host: true, // Listen to local network
    allowedHosts: ['lofp.sanburnlab.com'], // Trusted subdomain
    // Production serves players from this dev server, and any request Vite fails
    // to load (scanners probing /@fs/...?raw for files that don't exist) is pushed
    // to every connected browser as a full-screen error overlay. Errors still
    // show in the terminal running Vite.
    hmr: { overlay: false },
    proxy: {
      // Trailing slash matters: a bare '/api' prefix also swallows the
      // frontend's own /api-docs page and sends it to the backend (404).
      '/api/': 'http://localhost:4993',
      '/healthz': 'http://localhost:4993',
      '/ws': {
        target: 'http://localhost:4993',
        ws: true,
      },
    },
  },
})