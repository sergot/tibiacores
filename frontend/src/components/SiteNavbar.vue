<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useSuggestionsStore } from '@/stores/suggestions'
import { useChatNotificationsStore } from '@/stores/chatNotifications'
import { useI18n } from 'vue-i18n'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import {
  Bars3Icon,
  XMarkIcon,
  ArrowRightStartOnRectangleIcon,
  UserIcon,
  ExclamationTriangleIcon,
  ChatBubbleLeftRightIcon,
  BellIcon,
} from '@heroicons/vue/24/outline'
import { Menu, MenuButton, MenuItems, MenuItem, Dialog, DialogPanel } from '@headlessui/vue'

const { t } = useI18n()
const userStore = useUserStore()
const router = useRouter()
const suggestionsStore = useSuggestionsStore()
const chatNotificationsStore = useChatNotificationsStore()
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
  suggestionsStore.stopPolling()
  chatNotificationsStore.stopPolling()
}

const cancelLogout = () => {
  showLogoutWarning.value = false
}

const handleSuggestionClick = () => {
  suggestionsStore.fetchPendingSuggestions() // Refresh suggestions after navigation
}

const handleChatNotificationClick = (listId: string) => {
  chatNotificationsStore.markAsRead(listId)
}

const handleMobileChatNotificationClick = (listId: string) => {
  isMenuOpen.value = false
  chatNotificationsStore.markAsRead(listId)
}

onMounted(() => {
  if (userStore.isAuthenticated) {
    suggestionsStore.startPolling()
    chatNotificationsStore.startPolling()
  }
})

onBeforeUnmount(() => {
  // Cleanup handled by Headless UI
})
</script>

<template>
  <nav class="bg-white shadow">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between h-16">
        <div class="flex items-center space-x-8">
          <div class="flex-shrink-0">
            <RouterLink to="/" class="flex items-center space-x-2">
              <img src="/logo.png" alt="TibiaCores Logo" class="h-12 w-12" />
              <span class="text-xl font-bold text-gray-800">TibiaCores</span>
            </RouterLink>
          </div>
          <div class="hidden md:flex md:items-center md:space-x-8">
            <RouterLink
              to="/highscores"
              class="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
            >
              {{ t('nav.highscores') }}
            </RouterLink>
            <RouterLink
              to="/blog"
              class="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
            >
              News
            </RouterLink>
            <div class="relative px-3 py-2 rounded-md text-sm font-medium text-gray-400 cursor-not-allowed flex items-center">
              {{ t('nav.marketplace.title') }}
              <span class="ml-2 text-xs bg-gradient-to-r from-indigo-100 to-purple-100 text-indigo-700 font-semibold px-2.5 py-1 rounded-full border border-indigo-200 shadow-sm">
                {{ t('nav.marketplace.comingSoon') }}
              </span>
            </div>
            <RouterLink
              to="/about"
              class="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
            >
              {{ t('nav.about') }}
            </RouterLink>
          </div>
        </div>

        <!-- Desktop Navigation -->
        <div class="hidden md:flex md:items-center md:space-x-4">
          <LanguageSwitcher />
          <div v-if="!userStore.isAuthenticated" class="flex space-x-4">
            <RouterLink
              to="/signin"
              class="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium"
            >
              {{ t('nav.signIn') }}
            </RouterLink>
            <RouterLink
              to="/signup"
              class="bg-indigo-600 text-white hover:bg-indigo-700 px-4 py-2 rounded-md text-sm font-medium transition-colors"
            >
              {{ t('nav.signUp') }}
            </RouterLink>
          </div>
          <div v-else class="flex items-center space-x-4">
            <!-- Notifications Area -->
            <div class="flex items-center space-x-2">
              <!-- Suggestions Notifications -->
              <Menu as="div" class="relative" v-if="suggestionsStore.hasPendingSuggestions">
                <MenuButton
                  class="relative text-gray-600 hover:text-gray-900 p-2 rounded-lg hover:bg-gray-100 ui-focus-visible:outline-none ui-focus-visible:ring-2 ui-focus-visible:ring-indigo-500"
                  :aria-label="t('nav.suggestions.aria')"
                >
                  <BellIcon class="h-6 w-6" />
                  <span
                    v-if="suggestionsStore.totalPendingSuggestions > 0"
                    class="absolute top-0 right-0 inline-flex items-center justify-center px-2 py-1 text-xs font-bold leading-none text-white transform translate-x-1/2 -translate-y-1/2 bg-red-500 rounded-full"
                  >
                    {{ suggestionsStore.totalPendingSuggestions }}
                  </span>
                </MenuButton>

                <transition
                  enter-active-class="transition ease-out duration-100"
                  enter-from-class="transform opacity-0 scale-95"
                  enter-to-class="transform opacity-100 scale-100"
                  leave-active-class="transition ease-in duration-75"
                  leave-from-class="transform opacity-100 scale-100"
                  leave-to-class="transform opacity-0 scale-95"
                >
                  <MenuItems
                    class="absolute right-0 mt-2 w-72 bg-white rounded-lg shadow-lg py-1 z-50 border border-gray-200 ui-focus-visible:outline-none"
                  >
                    <div class="px-4 py-2 text-sm font-medium text-gray-700 border-b border-gray-200">
                      {{ t('nav.suggestions.title') }}
                    </div>
                    <div class="max-h-96 overflow-y-auto">
                      <MenuItem
                        v-for="char in suggestionsStore.pendingSuggestions"
                        :key="char.character_id"
                        v-slot="{ active }"
                      >
                        <RouterLink
                          :to="{ name: 'character-details', params: { id: char.character_id } }"
                          :class="[
                            active ? 'bg-gray-100' : '',
                            'block px-4 py-2 text-sm text-gray-700'
                          ]"
                          @click="handleSuggestionClick"
                        >
                          <div class="flex justify-between items-center">
                            <span>{{ char.character_name }}</span>
                            <span
                              class="bg-blue-100 text-blue-800 text-xs font-medium px-2.5 py-0.5 rounded-full"
                            >
                              {{ t('nav.suggestions.count', { count: char.suggestion_count }) }}
                            </span>
                          </div>
                        </RouterLink>
                      </MenuItem>
                    </div>
                  </MenuItems>
                </transition>
              </Menu>

              <!-- Chat Notifications -->
              <Menu as="div" class="relative" v-if="chatNotificationsStore.hasUnreadMessages">
                <MenuButton
                  class="relative text-gray-600 hover:text-gray-900 p-2 rounded-lg hover:bg-gray-100 ui-focus-visible:outline-none ui-focus-visible:ring-2 ui-focus-visible:ring-indigo-500"
                  :aria-label="t('nav.chat.aria')"
                >
                  <ChatBubbleLeftRightIcon class="h-6 w-6" />
                  <span
                    v-if="chatNotificationsStore.totalUnreadMessages > 0"
                    class="absolute top-0 right-0 inline-flex items-center justify-center px-2 py-1 text-xs font-bold leading-none text-white transform translate-x-1/2 -translate-y-1/2 bg-green-500 rounded-full"
                  >
                    {{ chatNotificationsStore.totalUnreadMessages }}
                  </span>
                </MenuButton>

                <transition
                  enter-active-class="transition ease-out duration-100"
                  enter-from-class="transform opacity-0 scale-95"
                  enter-to-class="transform opacity-100 scale-100"
                  leave-active-class="transition ease-in duration-75"
                  leave-from-class="transform opacity-100 scale-100"
                  leave-to-class="transform opacity-0 scale-95"
                >
                  <MenuItems
                    class="absolute right-0 mt-2 w-72 bg-white rounded-lg shadow-lg py-1 z-50 border border-gray-200 ui-focus-visible:outline-none"
                  >
                    <div class="px-4 py-2 text-sm font-medium text-gray-700 border-b border-gray-200">
                      {{ t('nav.chat.title') }}
                    </div>
                    <div class="max-h-96 overflow-y-auto">
                      <MenuItem
                        v-for="notification in chatNotificationsStore.notifications"
                        :key="notification.list_id"
                        v-slot="{ active }"
                      >
                        <RouterLink
                          :to="{ name: 'list-detail', params: { id: notification.list_id } }"
                          :class="[
                            active ? 'bg-gray-100' : '',
                            'block px-4 py-2 text-sm text-gray-700'
                          ]"
                          @click="handleChatNotificationClick(notification.list_id)"
                        >
                          <div class="flex flex-col">
                            <div class="flex justify-between items-center">
                              <span class="font-medium">{{ notification.list_name }}</span>
                              <span
                                class="bg-green-100 text-green-800 text-xs font-medium px-2.5 py-0.5 rounded-full"
                              >
                                {{ t('nav.chat.count', { count: notification.unread_count }) }}
                              </span>
                            </div>
                            <div class="text-xs text-gray-500 mt-1">
                              {{ t('nav.chat.lastMessage', { character: notification.last_character_name }) }}
                            </div>
                          </div>
                        </RouterLink>
                      </MenuItem>
                      <div v-if="chatNotificationsStore.notifications.length === 0" class="px-4 py-3 text-sm text-gray-500">
                        {{ t('nav.chat.noMessages') }}
                      </div>
                    </div>
                  </MenuItems>
                </transition>
              </Menu>
            </div>

            <RouterLink
              to="/profile"
              class="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium flex items-center space-x-2"
            >
              <UserIcon class="h-5 w-5" />
              <span>{{ t('nav.profile') }}</span>
            </RouterLink>
            <button
              @click="initiateLogout"
              class="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium flex items-center space-x-2"
            >
              <ArrowRightStartOnRectangleIcon class="h-5 w-5" />
              <span>{{ t('nav.signOut') }}</span>
            </button>
          </div>
        </div>

        <!-- Mobile menu button -->
        <div class="flex items-center space-x-2 md:hidden">
          <LanguageSwitcher />
          <button
            @click="toggleMenu"
            class="inline-flex items-center justify-center p-2 rounded-md text-gray-600 hover:text-gray-900 hover:bg-gray-100"
          >
            <span class="sr-only">Open main menu</span>
            <Bars3Icon v-if="!isMenuOpen" class="block h-6 w-6" />
            <XMarkIcon v-else class="block h-6 w-6" />
          </button>
        </div>
      </div>
    </div>

    <!-- Mobile menu -->
    <div v-if="isMenuOpen" class="md:hidden bg-white border-t border-gray-200">
      <div class="px-2 pt-2 pb-3 space-y-1">
        <!-- Main navigation links for all users -->
        <RouterLink
          to="/highscores"
          class="block px-3 py-2 rounded-md text-base font-medium text-gray-600 hover:text-gray-900 hover:bg-gray-100"
          @click="isMenuOpen = false"
        >
          {{ t('nav.highscores') }}
        </RouterLink>
        <RouterLink
          to="/blog"
          class="block px-3 py-2 rounded-md text-base font-medium text-gray-600 hover:text-gray-900 hover:bg-gray-100"
          @click="isMenuOpen = false"
        >
          News
        </RouterLink>
        <div class="block px-3 py-2 rounded-md text-base font-medium text-gray-400 cursor-not-allowed">
          <div class="flex items-center justify-between">
            <span>{{ t('nav.marketplace.title') }}</span>
            <span class="text-xs bg-gradient-to-r from-indigo-100 to-purple-100 text-indigo-700 font-semibold px-2.5 py-1 rounded-full border border-indigo-200 shadow-sm">
              {{ t('nav.marketplace.comingSoon') }}
            </span>
          </div>
        </div>
        <RouterLink
          to="/about"
          class="block px-3 py-2 rounded-md text-base font-medium text-gray-600 hover:text-gray-900 hover:bg-gray-100"
          @click="isMenuOpen = false"
        >
          {{ t('nav.about') }}
        </RouterLink>

        <!-- Authentication-specific sections -->
        <div v-if="!userStore.isAuthenticated" class="space-y-2 pt-2 border-t border-gray-200">
          <RouterLink
            to="/signin"
            class="block px-3 py-2 rounded-md text-base font-medium text-gray-600 hover:text-gray-900 hover:bg-gray-100"
            @click="isMenuOpen = false"
          >
            {{ t('nav.signIn') }}
          </RouterLink>
          <RouterLink
            to="/signup"
            class="block px-3 py-2 rounded-md text-base font-medium bg-indigo-600 text-white hover:bg-indigo-700"
            @click="isMenuOpen = false"
          >
            {{ t('nav.signUp') }}
          </RouterLink>
        </div>
        <div v-else class="space-y-2 pt-2 border-t border-gray-200">
          <!-- Suggestions notifications for mobile -->
          <RouterLink
            v-if="suggestionsStore.hasPendingSuggestions"
            :to="{
              name: 'character-details',
              params: { id: suggestionsStore.pendingSuggestions[0]?.character_id },
            }"
            class="flex justify-between items-center px-3 py-2 rounded-md text-base font-medium text-gray-600 hover:text-gray-900 hover:bg-gray-100"
            @click="isMenuOpen = false"
          >
            <span>{{ t('nav.suggestions.title') }}</span>
            <span class="bg-red-500 text-white text-xs font-bold px-2 py-1 rounded-full">
              {{ suggestionsStore.totalPendingSuggestions }}
            </span>
          </RouterLink>

          <!-- Chat notifications for mobile -->
          <RouterLink
            v-if="chatNotificationsStore.hasUnreadMessages"
            :to="{
              name: 'list-detail',
              params: { id: chatNotificationsStore.notifications[0]?.list_id },
            }"
            class="flex justify-between items-center px-3 py-2 rounded-md text-base font-medium text-gray-600 hover:text-gray-900 hover:bg-gray-100"
            @click="handleMobileChatNotificationClick(chatNotificationsStore.notifications[0]?.list_id)"
          >
            <span>{{ t('nav.chat.title') }}</span>
            <span class="bg-green-500 text-white text-xs font-bold px-2 py-1 rounded-full">
              {{ chatNotificationsStore.totalUnreadMessages }}
            </span>
          </RouterLink>
          <RouterLink
            to="/profile"
            class="block px-3 py-2 rounded-md text-base font-medium text-gray-600 hover:text-gray-900 hover:bg-gray-100"
            @click="isMenuOpen = false"
          >
            {{ t('nav.profile') }}
          </RouterLink>
          <button
            @click="initiateLogout"
            class="w-full text-left px-3 py-2 rounded-md text-base font-medium text-gray-600 hover:text-gray-900 hover:bg-gray-100"
          >
            {{ t('nav.signOut') }}
          </button>
        </div>
      </div>
    </div>
  </nav>

  <!-- Logout warning modal -->
  <Dialog :open="showLogoutWarning" @close="cancelLogout">
    <div class="fixed inset-0 bg-black/75 backdrop-blur-sm z-50 flex items-center justify-center p-4">
      <DialogPanel class="bg-white rounded-lg max-w-md w-full p-6">
        <div class="flex items-start">
          <div class="flex-shrink-0">
            <ExclamationTriangleIcon class="h-6 w-6 text-yellow-400" />
          </div>
          <div class="ml-3">
            <h3 class="text-lg font-medium text-gray-900">{{ t('nav.logout.warning.title') }}</h3>
            <div class="mt-2">
              <p class="text-sm text-gray-500">
                {{ t('nav.logout.warning.message') }}
              </p>
            </div>
            <div class="mt-4 flex flex-col sm:flex-row sm:space-x-4 space-y-2 sm:space-y-0">
              <RouterLink
                to="/signup"
                class="inline-flex justify-center px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-md"
                @click="showLogoutWarning = false"
              >
                {{ t('nav.logout.warning.register') }}
              </RouterLink>
              <button
                type="button"
                class="inline-flex justify-center px-4 py-2 text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 border border-gray-300 rounded-md"
                @click="handleLogout"
              >
                {{ t('nav.logout.warning.signOut') }}
              </button>
              <button
                type="button"
                class="inline-flex justify-center px-4 py-2 text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 border border-gray-300 rounded-md"
                @click="cancelLogout"
              >
                {{ t('nav.logout.warning.cancel') }}
              </button>
            </div>
          </div>
        </div>
      </DialogPanel>
    </div>
  </Dialog>
</template>
