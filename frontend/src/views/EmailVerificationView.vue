<template>
  <div class="min-h-screen bg-gray-100 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md mx-auto bg-white rounded-lg shadow-md p-8">
      <div v-if="isVerifying" class="text-center">
        <svg
          class="animate-spin h-10 w-10 text-indigo-500 mx-auto mb-4"
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
        <h2 class="text-xl font-semibold mb-2">{{ t('emailVerification.verifying.message') }}</h2>
        <p class="text-gray-600">{{ t('emailVerification.subtitle') }}</p>
      </div>

      <div v-else-if="error" class="text-center">
        <svg
          class="h-12 w-12 text-red-500 mx-auto mb-4"
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
        <h2 class="text-xl font-semibold text-red-600 mb-2">
          {{ t('emailVerification.verifying.error') }}
        </h2>
        <p class="text-gray-600 mb-4">{{ error }}</p>
        <div class="space-x-4">
          <router-link
            to="/profile"
            class="inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-indigo-700 bg-indigo-100 hover:bg-indigo-200"
          >
            {{ t('emailVerification.actions.backToProfile') }}
          </router-link>
          <button
            @click="router.go(0)"
            class="inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
          >
            {{ t('emailVerification.actions.tryAgain') }}
          </button>
        </div>
      </div>

      <div v-else class="text-center">
        <svg
          class="h-12 w-12 text-green-500 mx-auto mb-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M5 13l4 4L19 7"
          />
        </svg>
        <h2 class="text-xl font-semibold text-green-600 mb-2">
          {{ t('emailVerification.verifying.success') }}
        </h2>
        <p class="text-gray-600 mb-4">{{ t('emailVerification.verifying.redirect') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import axios from 'axios'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const isVerifying = ref(true)
const error = ref('')

onMounted(async () => {
  const token = route.query.token
  const userId = route.query.user_id

  if (!token || !userId) {
    error.value = t('emailVerification.verifying.invalidToken')
    isVerifying.value = false
    return
  }

  try {
    await axios.get(`/verify-email`, {
      params: {
        token,
        user_id: userId,
      },
    })
    isVerifying.value = false
    // Redirect to home after 2 seconds
    setTimeout(() => {
      router.replace('/')
    }, 2000)
  } catch (err) {
    error.value =
      axios.isAxiosError(err) && err.response?.data?.message
        ? err.response.data.message
        : t('emailVerification.verifying.error')
    isVerifying.value = false
  }
})
</script>
