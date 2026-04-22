import { defineConfig } from 'vite';
import preact from '@preact/preset-vite';

const buildSha = process.env.VITE_BUILD_SHA || 'local';

export default defineConfig({
  plugins: [preact()],
  build: {
    rollupOptions: {
      output: {
        entryFileNames: `assets/[name]-[hash]-${buildSha}.js`,
        chunkFileNames: `assets/[name]-[hash]-${buildSha}.js`,
        assetFileNames: `assets/[name]-[hash]-${buildSha}[extname]`,
      },
    },
  },
  resolve: {
    alias: {
      'react': 'preact/compat',
      'react-dom/test-utils': 'preact/test-utils',
      'react-dom': 'preact/compat',
      'react/jsx-runtime': 'preact/jsx-runtime',
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './tests/setup.ts',
    exclude: ['**/node_modules/**', '**/tests/e2e/**', '**/*.spec.ts'],
    server: {
      deps: {
        inline: ['react-router-dom', 'react-router', '@remix-run/router'],
      },
    },
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