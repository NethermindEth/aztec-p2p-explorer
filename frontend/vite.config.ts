import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import checker from 'vite-plugin-checker';
import svgr from 'vite-plugin-svgr';

// https://vitejs.dev/config/
export default defineConfig({
  envPrefix: ['VITE_', 'TURNSTILE_'],
  plugins: [react(), checker({ typescript: true }), svgr()],
  define: {
    // Pass environment variables to the client
    'import.meta.env.VITE_AZTEC_NETWORK': JSON.stringify(process.env.VITE_AZTEC_NETWORK || 'aztec-testnet'),
  },
  server: {
    proxy: {
      '/api': {
        target: process.env.VITE_BASE_URL || 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
