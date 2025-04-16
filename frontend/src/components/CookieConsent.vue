<template>
  <div
    v-if="!hasConsent"
    class="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 shadow-lg z-50"
  >
    <div class="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8">
      <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div class="flex-1">
          <h3 class="text-lg font-medium text-gray-900">{{ t('cookieConsent.title') }}</h3>
          <p class="mt-1 text-sm text-gray-500">
            {{ t('cookieConsent.description') }}
            <RouterLink to="/privacy" class="text-indigo-600 hover:text-indigo-500 underline">
              {{ t('cookieConsent.privacyPolicy') }}
            </RouterLink>
          </p>
          <div class="mt-4 space-y-2">
            <div class="flex items-center">
              <input
                id="necessary"
                type="checkbox"
                checked
                disabled
                class="h-4 w-4 text-indigo-600 border-gray-300 rounded"
              />
              <label for="necessary" class="ml-2 text-sm text-gray-700">
                {{ t('cookieConsent.categories.necessary') }}
              </label>
            </div>
            <div class="flex items-center">
              <input
                id="preferences"
                v-model="preferences"
                type="checkbox"
                class="h-4 w-4 text-indigo-600 border-gray-300 rounded"
              />
              <label for="preferences" class="ml-2 text-sm text-gray-700">
                {{ t('cookieConsent.categories.preferences') }}
              </label>
            </div>
            <div class="flex items-center">
              <input
                id="analytics"
                v-model="analytics"
                type="checkbox"
                class="h-4 w-4 text-indigo-600 border-gray-300 rounded"
              />
              <label for="analytics" class="ml-2 text-sm text-gray-700">
                {{ t('cookieConsent.categories.analytics') }}
              </label>
            </div>
          </div>
        </div>
        <div class="flex flex-col sm:flex-row gap-2">
          <button
            type="button"
            class="inline-flex items-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            @click="acceptAll"
          >
            {{ t('cookieConsent.acceptAll') }}
          </button>
          <button
            type="button"
            class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            @click="savePreferences"
          >
            {{ t('cookieConsent.savePreferences') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'
import { loadUmamiAnalytics } from '@/main'

const { t } = useI18n()

const hasConsent = ref(false)
const preferences = ref(false)
const analytics = ref(false)

const COOKIE_CONSENT_KEY = 'cookie_consent'

onMounted(() => {
  const consent = localStorage.getItem(COOKIE_CONSENT_KEY)
  if (consent) {
    const parsedConsent = JSON.parse(consent)
    hasConsent.value = true
    preferences.value = parsedConsent.preferences
    analytics.value = parsedConsent.analytics
  }
})

const savePreferences = () => {
  const consent = {
    preferences: preferences.value,
    analytics: analytics.value,
    timestamp: new Date().toISOString(),
  }
  localStorage.setItem(COOKIE_CONSENT_KEY, JSON.stringify(consent))
  hasConsent.value = true

  // Only load analytics if user consented
  if (analytics.value) {
    loadUmamiAnalytics()
  }
}

const acceptAll = () => {
  // Ensure analytics is enabled when accepting all
  analytics.value = true
  preferences.value = true

  const consent = {
    preferences: true,
    analytics: true,
    timestamp: new Date().toISOString(),
  }
  localStorage.setItem(COOKIE_CONSENT_KEY, JSON.stringify(consent))
  hasConsent.value = true

  // Load Umami analytics
  loadUmamiAnalytics()
}
</script>
