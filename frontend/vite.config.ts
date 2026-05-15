import tailwindcss from '@tailwindcss/vite';
import vue from '@vitejs/plugin-vue';
import path from 'path';
import {defineConfig, loadEnv} from 'vite';

export default defineConfig(({mode}) => {
  const env = loadEnv(mode, '.', '');
  const buildId = process.env.OHARA_BUILD_ID || Date.now().toString(36);
  return {
    plugins: [vue(), tailwindcss()],
    define: {
      'process.env.GEMINI_API_KEY': JSON.stringify(env.GEMINI_API_KEY),
    },
    resolve: {
      alias: {
        '@': path.resolve(__dirname, '.'),
      },
    },
    build: {
      rollupOptions: {
        output: {
          assetFileNames: `assets/[name]-[hash]-${buildId}[extname]`,
          chunkFileNames: `assets/[name]-[hash]-${buildId}.js`,
          entryFileNames: `assets/[name]-[hash]-${buildId}.js`,
        },
      },
    },
    server: {
      // HMR is disabled in AI Studio via DISABLE_HMR env var.
      // Do not modify—file watching is disabled to prevent flickering during agent edits.
      hmr: process.env.DISABLE_HMR !== 'true',
      proxy: {
        '/api': 'http://localhost:3000',
        '/manga': 'http://localhost:3000',
        '/audio': 'http://localhost:3000',
      },
    },
  };
});
