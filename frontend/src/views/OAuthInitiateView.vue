<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import axios from 'axios'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const error = ref('')

onMounted(async () => {
  const provider = route.params.provider
  if (!provider) {
    router.push('/signin')
    return
  }

  try {
    // Get the OAuth URL from our backend which will include the proper state
    const response = await axios.get(`/auth/oauth/${provider}`)
    // Use window with proper type checking
    ;(window as Window).location.href = response.data
  } catch (err) {
    console.error('Failed to initiate OAuth:', err)
    error.value = t('oauth.initiate.error')
  }
})
</script>

<template>
  <div
    class="min-h-[calc(100vh-8rem)] flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8 bg-gray-100"
  >
    <div class="max-w-md w-full space-y-8">
      <div v-if="error" class="text-center">
        <svg
          class="mx-auto h-12 w-12 text-red-500"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
          />
        </svg>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
          {{ error }}
        </h2>
        <button
          @click="router.go(0)"
          class="mt-4 inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
        >
          {{ t('oauth.initiate.tryAgain') }}
        </button>
      </div>
      <div v-else class="text-center">
        <svg
          class="mx-auto h-12 w-12 text-indigo-600 animate-spin"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          ></circle>
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path>
        </svg>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
          {{ t('oauth.initiate.loading', { provider: route.params.provider }) }}
        </h2>
      </div>
    </div>
  </div>
</template>
