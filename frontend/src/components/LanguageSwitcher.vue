<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { loadLocale } from '@/i18n'
import { GlobeAltIcon } from '@heroicons/vue/24/outline'

const { t } = useI18n()
const isOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

const languages: Array<{ code: 'en' | 'pl' | 'de' | 'es' | 'pt'; name: string }> = [
  { code: 'en', name: 'English' },
  { code: 'de', name: 'Deutsch' },
  { code: 'es', name: 'Español' },
  { code: 'pl', name: 'Polski' },
  { code: 'pt', name: 'Português' },
]

const switchLanguage = (locale: 'en' | 'pl' | 'de' | 'es' | 'pt') => {
  loadLocale(locale)
  isOpen.value = false
}

const handleClickOutside = (event: MouseEvent) => {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div class="relative" ref="dropdownRef">
    <button
      @click.stop="isOpen = !isOpen"
      class="p-2 text-gray-600 hover:text-gray-900 rounded-md hover:bg-gray-100 flex items-center space-x-1"
      aria-label="Change language"
    >
      <GlobeAltIcon class="h-5 w-5" />
      <span class="text-sm font-medium uppercase">{{ t('_locale') }}</span>
    </button>

    <div
      v-if="isOpen"
      class="absolute right-0 mt-2 w-48 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 z-50"
      :class="{
        'md:origin-top-right md:right-0 md:left-auto': true,
        'origin-top-right right-0 left-auto': true,
      }"
    >
      <div class="py-1" role="menu" aria-orientation="vertical">
        <button
          v-for="lang in languages"
          :key="lang.code"
          @click="switchLanguage(lang.code)"
          class="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 flex items-center justify-between"
          role="menuitem"
        >
          <span>{{ lang.name }}</span>
          <span v-if="t('_locale') === lang.code" class="text-indigo-600">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              class="h-5 w-5"
              viewBox="0 0 20 20"
              fill="currentColor"
            >
              <path
                fill-rule="evenodd"
                d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                clip-rule="evenodd"
              />
            </svg>
          </span>
        </button>
      </div>
    </div>
  </div>
</template>
