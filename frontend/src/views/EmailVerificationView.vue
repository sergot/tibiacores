<template>
  <div class="min-h-screen bg-gray-100 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md mx-auto bg-white rounded-lg shadow-md p-8">
      <div v-if="isVerifying" class="text-center">
        <svg class="animate-spin h-10 w-10 text-indigo-500 mx-auto mb-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <h2 class="text-xl font-semibold mb-2">Verifying your email...</h2>
        <p class="text-gray-600">Please wait while we verify your email address.</p>
      </div>

      <div v-else-if="error" class="text-center">
        <svg class="h-12 w-12 text-red-500 mx-auto mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        <h2 class="text-xl font-semibold text-red-600 mb-2">Verification Failed</h2>
        <p class="text-gray-600 mb-4">{{ error }}</p>
        <router-link to="/" class="text-indigo-600 hover:text-indigo-500">Return to Home</router-link>
      </div>

      <div v-else class="text-center">
        <svg class="h-12 w-12 text-green-500 mx-auto mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
        <h2 class="text-xl font-semibold text-green-600 mb-2">Email Verified!</h2>
        <p class="text-gray-600 mb-4">Your email has been successfully verified.</p>
        <router-link to="/" class="text-indigo-600 hover:text-indigo-500">Return to Home</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'

const route = useRoute()
const isVerifying = ref(true)
const error = ref('')

onMounted(async () => {
  const token = route.query.token
  const userId = route.query.user_id

  if (!token || !userId) {
    error.value = 'Invalid verification link'
    isVerifying.value = false
    return
  }

  try {
    await axios.get(`/verify-email`, {
      params: {
        token,
        user_id: userId
      }
    })
    
    isVerifying.value = false
  } catch (err) {
    error.value = axios.isAxiosError(err) && err.response?.data?.message 
      ? err.response.data.message 
      : 'Failed to verify email'
    isVerifying.value = false
  }
})
</script>