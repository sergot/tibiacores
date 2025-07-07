<template>
  <div class="bg-indigo-50 border border-indigo-200 rounded-lg p-6">
    <h3 class="text-lg font-semibold text-indigo-900 mb-3">{{ t('newsletter.title') }}</h3>
    <p class="text-indigo-700 mb-4 text-sm">
      {{ t('newsletter.description') }}
    </p>

    <div v-if="!isSubscribed" class="space-y-4">
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
        class="w-full bg-indigo-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-indigo-700 disabled:bg-indigo-400 disabled:cursor-not-allowed transition-colors"
      >
        {{ loading ? t('newsletter.subscribing') : t('newsletter.subscribe') }}
      </button>
    </div>

    <div v-else class="space-y-4">
      <div class="flex items-center space-x-2 text-green-700">
        <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
        <span class="text-sm font-medium">{{ t('newsletter.subscribed') }}</span>
      </div>

      <button
        @click="unsubscribe"
        :disabled="loading"
        class="w-full bg-gray-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-gray-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
      >
        {{ loading ? t('newsletter.unsubscribing') : t('newsletter.unsubscribe') }}
      </button>
    </div>

    <div v-if="message" class="mt-4 p-3 rounded-md text-sm" :class="messageClass">
      {{ message }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '@/stores/user'
import { newsletterService } from '@/services/newsletter'
import axios from 'axios'

const { t } = useI18n()
const userStore = useUserStore()

const email = ref('')
const loading = ref(false)
const message = ref('')
const messageType = ref<'success' | 'error' | 'info'>('info')
const isSubscribed = ref(false)
const checkingStatus = ref(false)
const userEmail = ref('')

const messageClass = computed(() => {
  switch (messageType.value) {
    case 'success':
      return 'bg-green-100 border border-green-200 text-green-800'
    case 'error':
      return 'bg-red-100 border border-red-200 text-red-800'
    case 'info':
    default:
      return 'bg-blue-100 border border-blue-200 text-blue-800'
  }
})

const showMessage = (msg: string, type: 'success' | 'error' | 'info' = 'info') => {
  message.value = msg
  messageType.value = type
  setTimeout(() => {
    message.value = ''
  }, 5000)
}

const fetchUserEmail = async () => {
  if (!userStore.isAuthenticated) return

  try {
    const response = await axios.get(`/users/${userStore.userId}`)
    if (response.data.email) {
      userEmail.value = response.data.email
      email.value = response.data.email
    }
  } catch (error) {
    console.error('Failed to fetch user email:', error)
  }
}

const checkSubscriptionStatus = async () => {
  if (!email.value.trim()) return

  checkingStatus.value = true
  try {
    const response = await newsletterService.checkSubscriptionStatus(email.value)
    isSubscribed.value = response.subscribed
  } catch (error) {
    console.error('Failed to check subscription status:', error)
  } finally {
    checkingStatus.value = false
  }
}

const subscribe = async () => {
  if (!email.value.trim()) return

  loading.value = true
  try {
    const response = await newsletterService.subscribe(email.value)

    if (response.status === 'subscribed') {
      showMessage(t('newsletter.messages.subscribed'), 'success')
      isSubscribed.value = true
    } else if (response.status === 'already_subscribed') {
      showMessage(t('newsletter.messages.alreadySubscribed'), 'info')
      isSubscribed.value = true
    } else if (response.status === 'resubscribed') {
      showMessage(t('newsletter.messages.resubscribed'), 'success')
      isSubscribed.value = true
    }
  } catch (error: any) {
    console.error('Newsletter subscription error:', error)
    if (error.response?.data?.message) {
      showMessage(error.response.data.message, 'error')
    } else {
      showMessage(t('newsletter.messages.error'), 'error')
    }
  } finally {
    loading.value = false
  }
}

const unsubscribe = async () => {
  if (!email.value.trim()) return

  loading.value = true
  try {
    const response = await newsletterService.unsubscribe(email.value)

    if (response.status === 'unsubscribed') {
      showMessage(t('newsletter.messages.unsubscribed'), 'success')
      isSubscribed.value = false
    }
  } catch (error: any) {
    console.error('Newsletter unsubscription error:', error)
    if (error.response?.data?.message) {
      showMessage(error.response.data.message, 'error')
    } else {
      showMessage(t('newsletter.messages.error'), 'error')
    }
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  // Clear any previous messages
  message.value = ''

  // If user is authenticated, try to auto-fill email and check status
  if (userStore.isAuthenticated) {
    await fetchUserEmail()
    if (email.value) {
      await checkSubscriptionStatus()
    }
  }
})
</script>
