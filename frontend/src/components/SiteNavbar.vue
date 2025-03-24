<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useSuggestionsStore } from '@/stores/suggestions'
import {
  Bars3Icon,
  XMarkIcon,
  ArrowRightStartOnRectangleIcon,
  UserPlusIcon,
  UserIcon,
  ExclamationTriangleIcon,
} from '@heroicons/vue/24/outline'

const userStore = useUserStore()
const router = useRouter()
const suggestionsStore = useSuggestionsStore()
const isMenuOpen = ref(false)
const showLogoutWarning = ref(false)
const showSuggestions = ref(false)

const toggleMenu = () => {
  isMenuOpen.value = !isMenuOpen.value
}

const initiateLogout = () => {
  if (userStore.isAnonymous) {
    showLogoutWarning.value = true
  } else {
    handleLogout()
  }
}

const handleLogout = () => {
  userStore.clearUser()
  router.push('/signin')
  isMenuOpen.value = false
  showLogoutWarning.value = false
  suggestionsStore.stopPolling()
}

const cancelLogout = () => {
  showLogoutWarning.value = false
}

// Close dropdown when clicking outside
const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (!target.closest('.relative')) {
    showSuggestions.value = false
  }
}

const handleSuggestionClick = () => {
  showSuggestions.value = false
  suggestionsStore.fetchPendingSuggestions() // Refresh suggestions after navigation
}

onMounted(() => {
  if (userStore.isAuthenticated) {
    suggestionsStore.startPolling()
  }
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <nav class="bg-white shadow">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between h-16">
        <div class="flex">
          <div class="flex-shrink-0 flex items-center">
            <RouterLink to="/" class="text-xl font-bold text-gray-800">FiendList</RouterLink>
          </div>
        </div>
        <div class="flex items-center">
          <div v-if="!userStore.isAuthenticated" class="flex space-x-4">
            <RouterLink
              to="/signin"
              class="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
            >
              Sign in
            </RouterLink>
            <RouterLink
              to="/signup"
              class="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
            >
              Sign up
            </RouterLink>
          </div>
          <div v-else class="flex items-center space-x-4">
            <!-- Suggestions Dropdown -->
            <div class="relative" v-if="suggestionsStore.hasPendingSuggestions">
              <button
                @click="showSuggestions = !showSuggestions"
                class="relative text-gray-600 hover:text-gray-900 p-2 rounded-lg hover:bg-gray-100"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                </svg>
                <span
                  class="absolute top-0 right-0 inline-flex items-center justify-center px-2 py-1 text-xs font-bold leading-none text-white transform translate-x-1/2 -translate-y-1/2 bg-red-500 rounded-full"
                >
                  {{ suggestionsStore.totalPendingSuggestions }}
                </span>
              </button>
              
              <!-- Dropdown menu -->
              <div
                v-if="showSuggestions"
                class="absolute right-0 mt-2 w-72 bg-white rounded-md shadow-lg py-1 z-50"
                @click.stop
              >
                <div class="px-4 py-2 text-sm font-medium text-gray-700 border-b border-gray-200">
                  Pending Soul Core Suggestions
                </div>
                <div class="max-h-96 overflow-y-auto">
                  <RouterLink
                    v-for="char in suggestionsStore.pendingSuggestions"
                    :key="char.character_id"
                    :to="{ name: 'character-details', params: { id: char.character_id }}"
                    class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                    @click="handleSuggestionClick"
                  >
                    <div class="flex justify-between items-center">
                      <span>{{ char.character_name }}</span>
                      <span class="bg-blue-100 text-blue-800 text-xs font-medium px-2 py-0.5 rounded-full">
                        {{ char.suggestion_count }} {{ char.suggestion_count === 1 ? 'suggestion' : 'suggestions' }}
                      </span>
                    </div>
                  </RouterLink>
                </div>
              </div>
            </div>

            <RouterLink
              to="/profile"
              class="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
            >
              Profile
            </RouterLink>
            <RouterLink
              to="/about"
              class="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
            >
              About
            </RouterLink>
            <button
              @click="initiateLogout"
              class="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
            >
              Log out
            </button>
          </div>
        </div>
      </div>
    </div>
  </nav>

  <!-- Logout warning modal -->
  <div v-if="showLogoutWarning" class="fixed inset-0 bg-black/75 backdrop-blur-sm z-50 flex items-center justify-center p-4">
    <div class="bg-white rounded-lg max-w-md w-full p-6">
      <div class="flex items-start">
        <div class="flex-shrink-0">
          <ExclamationTriangleIcon class="h-6 w-6 text-yellow-400" />
        </div>
        <div class="ml-3">
          <h3 class="text-lg font-medium text-gray-900">Warning: You're using an anonymous account</h3>
          <div class="mt-2">
            <p class="text-sm text-gray-500">
              If you sign out now, you'll lose access to your current session and all your data. Consider registering an account to save your progress.
            </p>
          </div>
          <div class="mt-4 flex space-x-4">
            <RouterLink
              to="/signup"
              class="inline-flex justify-center px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-md"
              @click="showLogoutWarning = false"
            >
              Register now
            </RouterLink>
            <button
              type="button"
              class="inline-flex justify-center px-4 py-2 text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 border border-gray-300 rounded-md"
              @click="handleLogout"
            >
              Sign out anyway
            </button>
            <button
              type="button"
              class="inline-flex justify-center px-4 py-2 text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 border border-gray-300 rounded-md"
              @click="cancelLogout"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
