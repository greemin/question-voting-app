import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './tests/setup.ts',
    exclude: ['**/node_modules/**', '**/tests/e2e/**', '**/*.spec.ts'],
  },
  server: {
    host: '0.0.0.0', // Listen on all available network interfaces
    port: 5173, // Default Vite port
    allowedHosts: typeof process.env.VITE_ALLOWED_HOSTS === 'string' ? [...process.env.VITE_ALLOWED_HOSTS.split(',')] :
      [],
    proxy: {
      '/api': {
        target: process.env.VITE_API_URL, // Proxy to Go backend service
        changeOrigin: true,
        ws: true,
        secure: process.env.ENV === 'production' || process.env.NODE_ENV === 'production',
      },
    },
  },
});