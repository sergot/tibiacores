<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/stores/user'
import axios from 'axios'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const userStore = useUserStore()
const error = ref('')

onMounted(async () => {
  const code = route.query.code as string
  const state = route.query.state as string
  const provider = route.params.provider

  if (!code) {
    error.value = t('oauth.callback.error.missingCode')
    setTimeout(() => router.push('/signin'), 3000)
    return
  }

  if (!state) {
    error.value = t('oauth.callback.error.missingState')
    setTimeout(() => router.push('/signin'), 3000)
    return
  }

  if (!provider) {
    error.value = t('oauth.callback.error.invalidProvider')
    setTimeout(() => router.push('/signin'), 3000)
    return
  }

  try {
    // Exchange code for token with our backend
    const response = await axios.get(`/auth/oauth/${provider}/callback`, {
      params: { code, state },
    })

    // Set the user in the store - cookies will handle authentication
    userStore.setUser({
      id: response.data.id,
      has_email: response.data.has_email,
    })

    // Navigate to profile page
    router.push('/profile')
  } catch (err) {
    if (axios.isAxiosError(err) && err.response?.status === 400) {
      error.value = t('oauth.callback.error.missingState')
    } else {
      error.value = t('oauth.callback.error.generic')
    }
    setTimeout(() => {
      router.push('/signin')
    }, 3000)
  }
})
</script>

<template>
  <div
    class="min-h-[calc(100vh-8rem)] flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8 bg-gray-100"
  >
    <div class="max-w-md w-full space-y-8">
      <div class="text-center">
        <div v-if="!error">
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
            {{ t('oauth.callback.loading') }}
          </h2>
        </div>
        <div v-else class="text-red-600">
          <svg class="mx-auto h-12 w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
            />
          </svg>
          <h2 class="mt-6 text-center text-xl font-medium">{{ error }}</h2>
          <p class="mt-2 text-sm">{{ t('oauth.callback.error.generic') }}</p>
          <p class="mt-2 text-sm text-gray-500">{{ t('emailVerification.verifying.redirect') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>
