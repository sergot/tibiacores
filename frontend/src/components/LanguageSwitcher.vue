<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { loadLocale } from '@/i18n'
import { GlobeAltIcon, CheckIcon } from '@heroicons/vue/24/outline'
import { Menu, MenuButton, MenuItems, MenuItem } from '@headlessui/vue'

const { t } = useI18n()

const languages: Array<{ code: 'en' | 'pl' | 'de' | 'es' | 'pt'; name: string }> = [
  { code: 'en', name: 'English' },
  { code: 'de', name: 'Deutsch' },
  { code: 'es', name: 'Español' },
  { code: 'pl', name: 'Polski' },
  { code: 'pt', name: 'Português' },
]

const switchLanguage = (locale: 'en' | 'pl' | 'de' | 'es' | 'pt') => {
  loadLocale(locale)
}
</script>

<template>
  <Menu as="div" class="relative">
    <MenuButton
      class="p-2 text-gray-600 hover:text-gray-900 rounded-md hover:bg-gray-100 flex items-center space-x-1 ui-focus-visible:outline-none ui-focus-visible:ring-2 ui-focus-visible:ring-indigo-500"
      :aria-label="t('language.change')"
    >
      <GlobeAltIcon class="h-5 w-5" />
      <span class="text-sm font-medium uppercase">{{ t('_locale') }}</span>
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
        class="absolute right-0 mt-2 w-48 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 z-50 ui-focus-visible:outline-none"
      >
        <div class="py-1">
          <MenuItem
            v-for="lang in languages"
            :key="lang.code"
            v-slot="{ active }"
          >
            <button
              @click="switchLanguage(lang.code)"
              :class="[
                active ? 'bg-gray-100' : '',
                'w-full text-left px-4 py-2 text-sm text-gray-700 flex items-center justify-between'
              ]"
            >
              <span>{{ lang.name }}</span>
              <CheckIcon
                v-if="t('_locale') === lang.code"
                class="h-5 w-5 text-indigo-600"
              />
            </button>
          </MenuItem>
        </div>
      </MenuItems>
    </transition>
  </Menu>
</template>
