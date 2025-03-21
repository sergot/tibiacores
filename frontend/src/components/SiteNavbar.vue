<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
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
const isMenuOpen = ref(false)
const showLogoutWarning = ref(false)

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
}

const cancelLogout = () => {
  showLogoutWarning.value = false
}
</script>

<template>
  <nav class="w-full bg-gray-800">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="relative flex items-center justify-between h-16">
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <img class="h-8 w-8" src="@/assets/logo.svg" alt="Logo" />
          </div>
          <!-- Desktop menu -->
          <div class="hidden md:block">
            <div class="ml-10 flex items-baseline space-x-4">
              <RouterLink
                to="/"
                class="text-gray-300 hover:bg-gray-700 hover:text-white px-3 py-2 rounded-md text-sm font-medium"
                :class="{ 'bg-gray-900 text-white': $route.path === '/' }"
              >
                Home
              </RouterLink>
              <RouterLink
                to="/about"
                class="text-gray-300 hover:bg-gray-700 hover:text-white px-3 py-2 rounded-md text-sm font-medium"
                :class="{ 'bg-gray-900 text-white': $route.path === '/about' }"
              >
                About
              </RouterLink>
            </div>
          </div>
        </div>

        <!-- Auth buttons (desktop) -->
        <div class="hidden md:flex items-center space-x-4">
          <template v-if="!userStore.isAuthenticated">
            <RouterLink
              to="/signin"
              class="text-gray-300 hover:bg-gray-700 hover:text-white px-3 py-2 rounded-md text-sm font-medium inline-flex items-center"
            >
              <ArrowRightStartOnRectangleIcon class="h-5 w-5 mr-2" />
              Sign in
            </RouterLink>
            <RouterLink
              to="/signup"
              class="bg-indigo-600 text-white hover:bg-indigo-700 px-3 py-2 rounded-md text-sm font-medium inline-flex items-center"
            >
              <UserPlusIcon class="h-5 w-5 mr-2" />
              Sign up
            </RouterLink>
          </template>
          <template v-else>
            <RouterLink
              to="/profile"
              class="text-gray-300 hover:bg-gray-700 hover:text-white px-3 py-2 rounded-md text-sm font-medium inline-flex items-center"
              :class="{ 'bg-gray-900 text-white': $route.path === '/profile' }"
            >
              <UserIcon class="h-5 w-5 mr-2" />
              Profile
            </RouterLink>
            <button
              @click="initiateLogout"
              class="text-gray-300 hover:bg-gray-700 hover:text-white px-3 py-2 rounded-md text-sm font-medium inline-flex items-center"
            >
              <ArrowRightStartOnRectangleIcon class="h-5 w-5 mr-2" />
              Sign out
            </button>
          </template>
        </div>

        <!-- Mobile menu button -->
        <div class="md:hidden">
          <button
            @click="toggleMenu"
            class="inline-flex items-center justify-center p-2 rounded-md text-gray-400 hover:text-white hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white"
          >
            <span class="sr-only">Open main menu</span>
            <Bars3Icon v-if="!isMenuOpen" class="block h-6 w-6" />
            <XMarkIcon v-else class="block h-6 w-6" />
          </button>
        </div>
      </div>

      <!-- Mobile menu -->
      <div v-show="isMenuOpen" class="md:hidden w-full">
        <div class="px-2 pt-2 pb-3 space-y-1 sm:px-3">
          <RouterLink
            to="/"
            class="block text-gray-300 hover:bg-gray-700 hover:text-white px-3 py-2 rounded-md text-base font-medium"
            :class="{ 'bg-gray-900 text-white': $route.path === '/' }"
            @click="isMenuOpen = false"
          >
            Home
          </RouterLink>
          <RouterLink
            to="/about"
            class="block text-gray-300 hover:bg-gray-700 hover:text-white px-3 py-2 rounded-md text-base font-medium"
            :class="{ 'bg-gray-900 text-white': $route.path === '/about' }"
            @click="isMenuOpen = false"
          >
            About
          </RouterLink>

          <!-- Auth buttons (mobile) -->
          <template v-if="!userStore.isAuthenticated">
            <RouterLink
              to="/signin"
              class="block text-gray-300 hover:bg-gray-700 hover:text-white px-3 py-2 rounded-md text-base font-medium inline-flex items-center"
              @click="isMenuOpen = false"
            >
              <ArrowRightStartOnRectangleIcon class="h-5 w-5 mr-2" />
              Sign in
            </RouterLink>
            <RouterLink
              to="/signup"
              class="block text-gray-300 hover:bg-gray-700 hover:text-white px-3 py-2 rounded-md text-base font-medium inline-flex items-center"
              @click="isMenuOpen = false"
            >
              <UserPlusIcon class="h-5 w-5 mr-2" />
              Sign up
            </RouterLink>
          </template>
          <template v-else>
            <RouterLink
              to="/profile"
              class="block text-gray-300 hover:bg-gray-700 hover:text-white px-3 py-2 rounded-md text-base font-medium inline-flex items-center"
              :class="{ 'bg-gray-900 text-white': $route.path === '/profile' }"
              @click="isMenuOpen = false"
            >
              <UserIcon class="h-5 w-5 mr-2" />
              Profile
            </RouterLink>
            <button
              @click="initiateLogout"
              class="block text-gray-300 hover:bg-gray-700 hover:text-white px-3 py-2 rounded-md text-base font-medium inline-flex items-center w-full text-left"
            >
              <ArrowRightStartOnRectangleIcon class="h-5 w-5 mr-2" />
              Sign out
            </button>
          </template>
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
