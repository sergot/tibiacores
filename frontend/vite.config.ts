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
        manualChunks: (id) => {
          // Group all node_modules dependencies
          if (id.includes('node_modules')) {
            // Group Vue ecosystem packages
            if (id.includes('vue') || id.includes('pinia') || id.includes('vue-router') || id.includes('vue-i18n')) {
              return 'vendor'
            }
            // Group UI packages
            if (id.includes('@headlessui') || id.includes('@heroicons')) {
              return 'ui'
            }
            // Group axios and other utilities
            if (id.includes('axios')) {
              return 'utils'
            }
            // Other node_modules packages
            return 'vendor'
          }
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
