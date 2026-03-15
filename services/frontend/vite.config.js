import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './tests/setup.ts',
  },
  server: {
    host: '0.0.0.0', // Listen on all available network interfaces
    port: 5173, // Default Vite port
    proxy: {
      '/api': {
        target: 'http://backend:8081', // Proxy to Go backend service
        changeOrigin: true,
        secure: false,
      },
    },
  },
});