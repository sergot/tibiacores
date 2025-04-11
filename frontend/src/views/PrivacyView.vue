<template>
  <div class="w-full bg-gradient-to-b from-gray-50 to-gray-100 min-h-[calc(100vh-8rem)]">
    <main class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <div class="text-center mb-12">
        <h1
          class="text-5xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-indigo-600 to-blue-600 mb-4"
        >
          {{ t('privacy.title') }}
        </h1>
        <p class="text-xl text-gray-600">{{ t('privacy.lastUpdated') }}</p>
      </div>

      <div class="grid gap-8 md:grid-cols-2">
        <!-- Introduction section -->
        <div
          class="group bg-white p-8 rounded-xl shadow-sm border border-gray-200 hover:shadow-lg transition-all duration-300 hover:border-indigo-200"
        >
          <div class="flex items-start space-x-4 mb-4">
            <div
              class="flex-shrink-0 w-12 h-12 rounded-full bg-indigo-100 flex items-center justify-center"
            >
              <svg
                class="w-6 h-6 text-indigo-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <h2 class="text-2xl font-semibold text-gray-800">
              {{ t('privacy.introduction.title') }}
            </h2>
          </div>
          <p class="text-gray-600 leading-relaxed group-hover:text-gray-700 transition-colors">
            {{ t('privacy.introduction.content') }}
          </p>
        </div>

        <!-- Cookies section -->
        <div
          class="group bg-white p-8 rounded-xl shadow-sm border border-gray-200 hover:shadow-lg transition-all duration-300 hover:border-green-200"
        >
          <div class="flex items-start space-x-4 mb-4">
            <div
              class="flex-shrink-0 w-12 h-12 rounded-full bg-green-100 flex items-center justify-center"
            >
              <svg
                class="w-6 h-6 text-green-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
                />
              </svg>
            </div>
            <h2 class="text-2xl font-semibold text-gray-800">{{ t('privacy.cookies.title') }}</h2>
          </div>
          <p class="text-gray-600 mb-4 leading-relaxed">
            {{ t('privacy.cookies.content') }}
          </p>
          <ul class="space-y-2 text-gray-600">
            <li class="flex items-start">
              <svg
                class="w-5 h-5 text-green-500 mt-1 mr-2"
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
              {{ t('privacy.cookies.types.necessary') }}
            </li>
            <li class="flex items-start">
              <svg
                class="w-5 h-5 text-green-500 mt-1 mr-2"
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
              {{ t('privacy.cookies.types.preferences') }}
            </li>
            <li class="flex items-start">
              <svg
                class="w-5 h-5 text-green-500 mt-1 mr-2"
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
              {{ t('privacy.cookies.types.analytics') }}
            </li>
          </ul>
        </div>

        <!-- Data Collection section -->
        <div
          class="group bg-white p-8 rounded-xl shadow-sm border border-gray-200 hover:shadow-lg transition-all duration-300 hover:border-purple-200"
        >
          <div class="flex items-start space-x-4 mb-4">
            <div
              class="flex-shrink-0 w-12 h-12 rounded-full bg-purple-100 flex items-center justify-center"
            >
              <svg
                class="w-6 h-6 text-purple-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                />
              </svg>
            </div>
            <h2 class="text-2xl font-semibold text-gray-800">{{ t('privacy.data.title') }}</h2>
          </div>
          <p class="text-gray-600 leading-relaxed group-hover:text-gray-700 transition-colors">
            {{ t('privacy.data.content') }}
          </p>
        </div>

        <!-- Security section -->
        <div
          class="group bg-white p-8 rounded-xl shadow-sm border border-gray-200 hover:shadow-lg transition-all duration-300 hover:border-yellow-200"
        >
          <div class="flex items-start space-x-4 mb-4">
            <div
              class="flex-shrink-0 w-12 h-12 rounded-full bg-yellow-100 flex items-center justify-center"
            >
              <svg
                class="w-6 h-6 text-yellow-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
                />
              </svg>
            </div>
            <h2 class="text-2xl font-semibold text-gray-800">{{ t('privacy.security.title') }}</h2>
          </div>
          <p class="text-gray-600 leading-relaxed group-hover:text-gray-700 transition-colors">
            {{ t('privacy.security.content') }}
          </p>
        </div>

        <!-- Contact section -->
        <div
          class="group bg-white p-8 rounded-xl shadow-sm border border-gray-200 hover:shadow-lg transition-all duration-300 hover:border-red-200"
        >
          <div class="flex items-start space-x-4 mb-4">
            <div
              class="flex-shrink-0 w-12 h-12 rounded-full bg-red-100 flex items-center justify-center"
            >
              <svg
                class="w-6 h-6 text-red-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
                />
              </svg>
            </div>
            <h2 class="text-2xl font-semibold text-gray-800">{{ t('privacy.contact.title') }}</h2>
          </div>
          <p class="text-gray-600 leading-relaxed group-hover:text-gray-700 transition-colors">
            {{ t('privacy.contact.content') }}
          </p>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>
