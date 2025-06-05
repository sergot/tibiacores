import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  server: {
    proxy: {
      '/api': {
        target: 'http://backend:8080',
      },
    }
  },
  plugins: [vue(), vueDevTools(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  build: {
    // Enable source maps for production debugging
    sourcemap: true,

    // Split chunks for better caching
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['vue', 'vue-router', 'vue-i18n', 'pinia'],
          ui: ['@headlessui/vue'],
          utils: ['axios']
        }
      }
    },

    // Optimize chunk size warnings
    chunkSizeWarningLimit: 1000,

    // Enable CSS code splitting
    cssCodeSplit: true,

    // Minify CSS and JS
    minify: 'esbuild',

    // Target modern browsers for better optimization
    target: 'es2020'
  },

  // Enable CSS preprocessing optimizations
  css: {
    devSourcemap: true
  }
})
