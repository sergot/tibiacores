<template>
  <div class="bg-indigo-50 border border-indigo-200 rounded-lg p-6">
    <h3 class="text-lg font-semibold text-indigo-900 mb-3">{{ title || t('newsletter.title') }}</h3>
    <p class="text-indigo-700 mb-4 text-sm">
      {{ subtitle || t('newsletter.subtitle') }}
    </p>
    
    <div v-if="!subscribed" class="space-y-3">
      <div class="mb-4">
        <input
          v-model="email"
          type="email"
          :placeholder="t('newsletter.emailPlaceholder')"
          class="w-full px-3 py-2 border border-indigo-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
          :disabled="loading"
          @keyup.enter="subscribe"
        />
      </div>
      <button
        @click="subscribe"
        :disabled="loading || !email.trim()"
        class="w-full bg-indigo-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-indigo-700 transition-colors disabled:bg-indigo-400 disabled:cursor-not-allowed"
      >
        <span v-if="loading" class="flex items-center justify-center">
          <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          {{ t('newsletter.subscribing') }}
        </span>
        <span v-else>{{ t('newsletter.subscribe') }}</span>
      </button>
      
      <div v-if="error" class="text-red-600 text-xs mt-2">
        {{ error }}
      </div>
    </div>
    
    <div v-else class="text-center">
      <div class="flex items-center justify-center mb-2">
        <svg class="h-6 w-6 text-green-600 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
        <span class="text-green-700 font-medium">{{ t('newsletter.success') }}</span>
      </div>
      <p class="text-indigo-700 text-sm">{{ t('newsletter.successMessage') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { newsletterService } from '@/services/newsletter'

interface Props {
  title?: string
  subtitle?: string
}

defineProps<Props>()

const { t } = useI18n()

const email = ref('')
const loading = ref(false)
const error = ref('')
const subscribed = ref(false)

const subscribe = async () => {
  if (loading.value || !email.value.trim()) return

  loading.value = true
  error.value = ''

  try {
    await newsletterService.subscribe(email.value.trim())
    subscribed.value = true
    email.value = ''
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('newsletter.error')
  } finally {
    loading.value = false
  }
}
</script>
