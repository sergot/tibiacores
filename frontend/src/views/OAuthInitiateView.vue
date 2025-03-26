<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'

const route = useRoute()
const router = useRouter()

onMounted(async () => {
  const provider = route.params.provider
  if (!provider) {
    router.push('/signin')
    return
  }

  try {
    // Get the OAuth URL from our backend which will include the proper state
    const response = await axios.get(`/auth/oauth/${provider}`)
    window.location.href = response.data
  } catch (err) {
    console.error('Failed to initiate OAuth:', err)
    router.push('/signin')
  }
})
</script>

<template>
  <div class="min-h-[calc(100vh-8rem)] flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8 bg-gray-100">
    <div class="max-w-md w-full space-y-8">
      <div class="text-center">
        <svg class="mx-auto h-12 w-12 text-indigo-600 animate-spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
          Redirecting to authentication...
        </h2>
      </div>
    </div>
  </div>
</template>