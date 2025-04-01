<template>
  <div class="bg-gradient-to-r from-indigo-50 to-blue-50 p-6 rounded-lg border border-indigo-100 shadow-sm">
    <h3 class="text-lg font-semibold text-indigo-900 mb-3">{{ t('registerSuggestion.title') }}</h3>
    <p class="text-gray-600 mb-4">
      {{ t('registerSuggestion.subtitle') }}
    </p>
    <ul class="space-y-2 mb-6">
      <li class="flex items-center text-gray-700">
        <svg class="h-5 w-5 text-indigo-500 mr-2" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
        </svg>
        {{ t('registerSuggestion.features.accessAnywhere') }}
      </li>
      <li class="flex items-center text-gray-700">
        <svg class="h-5 w-5 text-indigo-500 mr-2" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
        </svg>
        {{ t('registerSuggestion.features.secureData') }}
      </li>
      <li class="flex items-center text-gray-700">
        <svg class="h-5 w-5 text-indigo-500 mr-2" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
        </svg>
        {{ t('registerSuggestion.features.collaborate') }}
      </li>
    </ul>
    <button
      @click="navigateToSignup"
      class="w-full bg-indigo-600 text-white py-2 px-4 rounded-md hover:bg-indigo-700 transition-colors duration-200"
    >
      {{ t('registerSuggestion.completeButton') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'

const router = useRouter()
const { t } = useI18n()

const navigateToSignup = () => {
  router.push('/signup')
}
</script>
