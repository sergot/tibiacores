<template>
  <div class="w-full bg-gradient-to-b from-gray-50 to-gray-100 min-h-[calc(100vh-8rem)]">
    <main class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <div class="text-center mb-12">
        <h1
          class="text-4xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-pink-600 to-rose-600 mb-4"
        >
          {{ t('sponsor.title') }}
        </h1>
        <p class="text-xl text-gray-600 max-w-2xl mx-auto">
          {{ t('sponsor.subtitle') }}
        </p>
      </div>

      <div class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden mb-8">
        <div class="p-6">
          <h2 class="text-2xl font-semibold text-gray-800 mb-4">{{ t('sponsor.why.title') }}</h2>
          <p class="text-gray-600 mb-6">{{ t('sponsor.why.description') }}</p>

          <div class="grid gap-4 md:grid-cols-3">
            <div class="bg-pink-50 p-4 rounded-lg">
              <div class="flex items-center justify-center w-12 h-12 bg-pink-100 rounded-full mb-3">
                <svg
                  class="w-6 h-6 text-pink-600"
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
              </div>
              <h3 class="font-medium text-gray-900 mb-1">{{ t('sponsor.why.reasons.hosting') }}</h3>
            </div>
            <div class="bg-pink-50 p-4 rounded-lg">
              <div class="flex items-center justify-center w-12 h-12 bg-pink-100 rounded-full mb-3">
                <svg
                  class="w-6 h-6 text-pink-600"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M12 6v6m0 0v6m0-6h6m-6 0H6"
                  />
                </svg>
              </div>
              <h3 class="font-medium text-gray-900 mb-1">
                {{ t('sponsor.why.reasons.features') }}
              </h3>
            </div>
            <div class="bg-pink-50 p-4 rounded-lg">
              <div class="flex items-center justify-center w-12 h-12 bg-pink-100 rounded-full mb-3">
                <svg
                  class="w-6 h-6 text-pink-600"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M13 10V3L4 14h7v7l9-11h-7z"
                  />
                </svg>
              </div>
              <h3 class="font-medium text-gray-900 mb-1">
                {{ t('sponsor.why.reasons.motivation') }}
              </h3>
            </div>
          </div>
        </div>
      </div>

      <!-- GitHub Sponsors -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden mb-8">
        <div class="bg-gradient-to-r from-pink-500 to-rose-500 py-4 px-6">
          <h2 class="text-xl font-bold text-white flex items-center">
            <svg class="h-6 w-6 mr-2" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
              <path
                fill-rule="evenodd"
                d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1.41 15.06V16h2.67v1.06c0 .55-.45 1-1 1h-.67c-.55 0-1-.45-1-1zm0-9.12v4.18c0 .55.45 1 1 1h.67c.55 0 1-.45 1-1V7.94c0-.55-.45-1-1-1h-.67c-.55 0-1 .45-1 1zm6.88 5.06c0 2.76-2.24 5-5 5s-5-2.24-5-5V10h1.94v2.94c0 1.71 1.35 3.06 3.06 3.06s3.06-1.35 3.06-3.06V10h1.94v3z"
                clip-rule="evenodd"
              />
            </svg>
            GitHub Sponsors
          </h2>
        </div>
        <div class="p-6">
          <p class="text-gray-600 mb-6">{{ t('sponsor.github.description') }}</p>
          <div class="flex flex-col sm:flex-row gap-4">
            <a
              href="https://github.com/sponsors/sergot"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex justify-center items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-pink-600 hover:bg-pink-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-pink-500"
            >
              <svg class="h-5 w-5 mr-2" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <path
                  fill-rule="evenodd"
                  d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1.41 15.06V16h2.67v1.06c0 .55-.45 1-1 1h-.67c-.55 0-1-.45-1-1zm0-9.12v4.18c0 .55.45 1 1 1h.67c.55 0 1-.45 1-1V7.94c0-.55-.45-1-1-1h-.67c-.55 0-1 .45-1 1zm6.88 5.06c0 2.76-2.24 5-5 5s-5-2.24-5-5V10h1.94v2.94c0 1.71 1.35 3.06 3.06 3.06s3.06-1.35 3.06-3.06V10h1.94v3z"
                  clip-rule="evenodd"
                />
              </svg>
              {{ t('sponsor.github.button') }}
            </a>
          </div>
        </div>
      </div>

      <!-- In-game Donations -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden mb-8">
        <div class="bg-gradient-to-r from-yellow-500 to-amber-500 py-4 px-6">
          <h2 class="text-xl font-bold text-white flex items-center">
            <img src="/assets/tibia-icon.gif" alt="Tibia Icon" class="w-6 h-6 mr-2" />
            {{ t('sponsor.ingame.title') }}
          </h2>
        </div>
        <div class="p-6">
          <p class="text-gray-600 mb-6">{{ t('sponsor.ingame.description') }}</p>
          <div class="flex items-center p-4 bg-yellow-50 rounded-lg">
            <img src="/assets/tibia-icon.gif" alt="Tibia Icon" class="w-8 h-8 mr-3" />
            <div>
              <div class="font-medium">Corescom</div>
              <a
                href="https://www.tibia.com/community/?name=Corescom"
                target="_blank"
                rel="noopener noreferrer"
                class="text-blue-600 hover:text-blue-800 hover:underline text-sm"
              >
                {{ t('sponsor.ingame.viewProfile') }}
              </a>
            </div>
          </div>
        </div>
      </div>

      <!-- FAQ -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        <div class="bg-gradient-to-r from-indigo-500 to-blue-500 py-4 px-6">
          <h2 class="text-xl font-bold text-white">{{ t('sponsor.faq.title') }}</h2>
        </div>
        <div class="p-6">
          <div class="space-y-6">
            <div>
              <h3 class="text-lg font-medium text-gray-900 mb-2">
                {{ t('sponsor.faq.questions.howMuch.question') }}
              </h3>
              <p class="text-gray-600">{{ t('sponsor.faq.questions.howMuch.answer') }}</p>
            </div>
            <div>
              <h3 class="text-lg font-medium text-gray-900 mb-2">
                {{ t('sponsor.faq.questions.benefit.question') }}
              </h3>
              <p class="text-gray-600">{{ t('sponsor.faq.questions.benefit.answer') }}</p>
            </div>
            <div>
              <h3 class="text-lg font-medium text-gray-900 mb-2">
                {{ t('sponsor.faq.questions.other.question') }}
              </h3>
              <p class="text-gray-600">{{ t('sponsor.faq.questions.other.answer') }}</p>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n({
  useScope: 'global',
})
</script>
