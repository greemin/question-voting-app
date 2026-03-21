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
        target: process.env.VITE_API_URL, // Proxy to Go backend service
        changeOrigin: true,
        ws: true,
        secure: process.env.ENV === 'production' || process.env.NODE_ENV === 'production',
      },
    },
  },
});