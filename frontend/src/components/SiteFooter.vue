<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <footer class="w-full bg-gray-800">
    <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
      <!-- Sponsors and Links -->
      <div class="py-4 flex flex-col sm:flex-row justify-between items-center">
        <div class="mb-4 sm:mb-0 text-center sm:text-left">
          <p class="text-gray-400 text-sm">
            <a
              href="https://github.com/sergot/tibiacores"
              target="_blank"
              rel="noopener noreferrer"
              class="hover:text-gray-300"
              >TibiaCores</a
            >
            © 2025. All rights reserved. Tibia and TibiaME are trademarks of
            <a
              href="https://www.cipsoft.com"
              target="_blank"
              rel="noopener noreferrer"
              class="hover:text-gray-300"
              >CipSoft GmbH</a
            >.
          </p>
        </div>

        <div class="flex items-center">
          <a
            href="https://github.com/sponsors/sergot"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-pink-600 hover:bg-pink-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-pink-500"
          >
            <svg class="h-5 w-5 mr-2" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
              <path
                fill-rule="evenodd"
                d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1.41 15.06V16h2.67v1.06c0 .55-.45 1-1 1h-.67c-.55 0-1-.45-1-1zm0-9.12v4.18c0 .55.45 1 1 1h.67c.55 0 1-.45 1-1V7.94c0-.55-.45-1-1-1h-.67c-.55 0-1 .45-1 1zm6.88 5.06c0 2.76-2.24 5-5 5s-5-2.24-5-5V10h1.94v2.94c0 1.71 1.35 3.06 3.06 3.06s3.06-1.35 3.06-3.06V10h1.94v3z"
                clip-rule="evenodd"
              />
            </svg>
            {{ t('sponsor.support') || 'Support TibiaCores' }}
          </a>
        </div>
      </div>
    </div>
  </footer>
</template>
