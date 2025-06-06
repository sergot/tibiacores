<template>
  <div class="w-full bg-gradient-to-b from-gray-50 to-gray-100 min-h-[calc(100vh-8rem)]">
    <main class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <BreadcrumbNavigation />

      <div class="text-center mb-12">
        <h1
          class="text-5xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-indigo-600 to-blue-600 mb-4"
        >
          {{ t('about.title') }}
        </h1>
        <p class="text-xl text-gray-600 max-w-2xl mx-auto">{{ t('about.subtitle') }}</p>
      </div>

      <div class="grid gap-8 md:grid-cols-2">
        <!-- What is TibiaCores section -->
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
                  d="M13 10V3L4 14h7v7l9-11h-7z"
                />
              </svg>
            </div>
            <h2 class="text-2xl font-semibold text-gray-800">{{ t('about.whatIs.title') }}</h2>
          </div>
          <p class="text-gray-600 leading-relaxed group-hover:text-gray-700 transition-colors">
            {{ t('about.whatIs.description') }}
          </p>
        </div>

        <!-- Features section -->
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
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </div>
            <h2 class="text-2xl font-semibold text-gray-800">{{ t('about.features.title') }}</h2>
          </div>
          <ul class="space-y-4 text-gray-600">
            <li
              class="flex items-center space-x-3 group-hover:transform group-hover:translate-x-1 transition-transform"
            >
              <span
                class="flex-shrink-0 w-6 h-6 bg-green-100 text-green-600 rounded-full flex items-center justify-center shadow-sm"
                >•</span
              >
              <span>{{ t('about.features.list.lists') }}</span>
            </li>
            <li
              class="flex items-center space-x-3 group-hover:transform group-hover:translate-x-1 transition-transform"
            >
              <span
                class="flex-shrink-0 w-6 h-6 bg-green-100 text-green-600 rounded-full flex items-center justify-center shadow-sm"
                >•</span
              >
              <span>{{ t('about.features.list.collaboration') }}</span>
            </li>
            <li
              class="flex items-center space-x-3 group-hover:transform group-hover:translate-x-1 transition-transform"
            >
              <span
                class="flex-shrink-0 w-6 h-6 bg-green-100 text-green-600 rounded-full flex items-center justify-center shadow-sm"
                >•</span
              >
              <span>{{ t('about.features.list.tracking') }}</span>
            </li>
            <li
              class="flex items-center space-x-3 group-hover:transform group-hover:translate-x-1 transition-transform"
            >
              <span
                class="flex-shrink-0 w-6 h-6 bg-green-100 text-green-600 rounded-full flex items-center justify-center shadow-sm"
                >•</span
              >
              <span>{{ t('about.features.list.organization') }}</span>
            </li>
          </ul>
        </div>

        <!-- Contribute section -->
        <div
          class="group bg-white p-8 rounded-xl shadow-sm border border-gray-200 hover:shadow-lg transition-all duration-300 hover:border-blue-200"
        >
          <div class="flex items-start space-x-4 mb-4">
            <div
              class="flex-shrink-0 w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center"
            >
              <svg
                class="w-6 h-6 text-blue-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                />
              </svg>
            </div>
            <h2 class="text-2xl font-semibold text-gray-800">{{ t('about.contribute.title') }}</h2>
          </div>
          <p class="text-gray-600 mb-6 leading-relaxed">
            {{ t('about.contribute.description') }}
          </p>
          <ul class="space-y-4 text-gray-600">
            <li
              class="flex items-center space-x-3 group-hover:transform group-hover:translate-x-1 transition-transform"
            >
              <svg class="w-6 h-6 text-gray-700" viewBox="0 0 24 24" fill="currentColor">
                <path
                  d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"
                />
              </svg>
              <div class="flex-1">
                <a
                  href="https://github.com/sergot/tibiacores"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-blue-600 hover:text-blue-800 hover:underline"
                >
                  github.com/sergot/tibiacores
                </a>
              </div>
            </li>
            <li
              class="flex items-center space-x-3 group-hover:transform group-hover:translate-x-1 transition-transform"
            >
              <span
                class="flex-shrink-0 w-6 h-6 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center shadow-sm"
                >•</span
              >
              <span>{{ t('about.contribute.list.bugs') }}</span>
            </li>
            <li
              class="flex items-center space-x-3 group-hover:transform group-hover:translate-x-1 transition-transform"
            >
              <span
                class="flex-shrink-0 w-6 h-6 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center shadow-sm"
                >•</span
              >
              <span>{{ t('about.contribute.list.code') }}</span>
            </li>
            <li
              class="flex items-center space-x-3 group-hover:transform group-hover:translate-x-1 transition-transform"
            >
              <span
                class="flex-shrink-0 w-6 h-6 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center shadow-sm"
                >•</span
              >
              <span>{{ t('about.contribute.list.docs') }}</span>
            </li>
          </ul>
        </div>

        <!-- Coming Soon section -->
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
                  d="M11.933 12.8a1 1 0 000-1.6L6.6 7.2A1 1 0 005 8v8a1 1 0 001.6.8l5.333-4zM19.933 12.8a1 1 0 000-1.6l-5.333-4A1 1 0 0013 8v8a1 1 0 001.6.8l5.333-4z"
                />
              </svg>
            </div>
            <h2 class="text-2xl font-semibold text-gray-800">{{ t('about.comingSoon.title') }}</h2>
          </div>
          <ul class="space-y-4 text-gray-600">
            <li
              class="flex items-center space-x-3 group-hover:transform group-hover:translate-x-1 transition-transform"
            >
              <span
                class="flex-shrink-0 w-6 h-6 bg-purple-100 text-purple-600 rounded-full flex items-center justify-center shadow-sm"
                >•</span
              >
              <span>{{ t('about.comingSoon.list.market') }}</span>
            </li>
            <li
              class="flex items-center space-x-3 group-hover:transform group-hover:translate-x-1 transition-transform"
            >
              <span
                class="flex-shrink-0 w-6 h-6 bg-purple-100 text-purple-600 rounded-full flex items-center justify-center shadow-sm"
                >•</span
              >
              <span>{{ t('about.comingSoon.list.achievements') }}</span>
            </li>
            <li
              class="flex items-center space-x-3 group-hover:transform group-hover:translate-x-1 transition-transform"
            >
              <span
                class="flex-shrink-0 w-6 h-6 bg-purple-100 text-purple-600 rounded-full flex items-center justify-center shadow-sm"
                >•</span
              >
              <span>{{ t('about.comingSoon.list.highscores') }}</span>
            </li>
            <li
              class="flex items-center space-x-3 group-hover:transform group-hover:translate-x-1 transition-transform"
            >
              <span
                class="flex-shrink-0 w-6 h-6 bg-purple-100 text-purple-600 rounded-full flex items-center justify-center shadow-sm"
                >•</span
              >
              <span>{{ t('about.comingSoon.list.more') }}</span>
            </li>
          </ul>
        </div>

        <!-- Donation section -->
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
                  d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>
            <h2 class="text-2xl font-semibold text-gray-800">{{ t('about.donation.title') }}</h2>
          </div>
          <p class="text-gray-600 mb-6 leading-relaxed">
            {{ t('about.donation.description') }}
          </p>
          <div class="space-y-4">
            <!-- GitHub Sponsors -->
            <div
              class="flex items-center space-x-3 p-4 bg-gradient-to-r from-pink-50 to-pink-100 rounded-lg group-hover:transform group-hover:translate-x-1 transition-transform"
            >
              <svg class="w-6 h-6 text-pink-600" viewBox="0 0 24 24" fill="currentColor">
                <path
                  fill-rule="evenodd"
                  d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1.41 15.06V16h2.67v1.06c0 .55-.45 1-1 1h-.67c-.55 0-1-.45-1-1zm0-9.12v4.18c0 .55.45 1 1 1h.67c.55 0 1-.45 1-1V7.94c0-.55-.45-1-1-1h-.67c-.55 0-1 .45-1 1zm6.88 5.06c0 2.76-2.24 5-5 5s-5-2.24-5-5V10h1.94v2.94c0 1.71 1.35 3.06 3.06 3.06s3.06-1.35 3.06-3.06V10h1.94v3z"
                  clip-rule="evenodd"
                />
              </svg>
              <a
                href="https://github.com/sponsors/sergot"
                target="_blank"
                rel="noopener noreferrer"
                class="text-blue-600 hover:text-blue-800 hover:underline"
              >
                GitHub Sponsors
              </a>
            </div>

            <!-- In-game donation -->
            <div
              class="flex items-center space-x-3 p-4 bg-gradient-to-r from-yellow-50 to-yellow-100 rounded-lg group-hover:transform group-hover:translate-x-1 transition-transform"
            >
              <img src="/assets/tibia-icon.gif" alt="Tibia Icon" class="w-6 h-6" />
              <a
                href="https://www.tibia.com/community/?name=Corescom"
                target="_blank"
                rel="noopener noreferrer"
                class="text-blue-600 hover:text-blue-800 hover:underline"
              >
                Corescom
              </a>
            </div>
          </div>
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
            <h2 class="text-2xl font-semibold text-gray-800">{{ t('about.contact.title') }}</h2>
          </div>
          <p class="text-gray-600 mb-6 leading-relaxed">
            {{ t('about.contact.description') }}
          </p>
          <ul class="space-y-4 text-gray-600">
            <li
              class="flex items-center space-x-3 p-4 bg-gradient-to-r from-[#5865F2]/10 to-[#5865F2]/5 rounded-lg group-hover:transform group-hover:translate-x-1 transition-transform"
            >
              <svg class="w-6 h-6 text-[#5865F2]" viewBox="0 0 24 24" fill="currentColor">
                <path
                  d="M20.317 4.492c-1.53-.69-3.17-1.2-4.885-1.49a.075.075 0 0 0-.079.036c-.21.369-.444.85-.608 1.23a18.566 18.566 0 0 0-5.487 0 12.36 12.36 0 0 0-.617-1.23A.077.077 0 0 0 8.562 3c-1.714.29-3.354.8-4.885 1.491a.07.07 0 0 0-.032.027C.533 9.093-.32 13.555.099 17.961a.08.08 0 0 0 .031.055 20.03 20.03 0 0 0 5.993 2.98.078.078 0 0 0 .084-.026 13.83 13.83 0 0 0 1.226-1.963.074.074 0 0 0-.041-.104 13.175 13.175 0 0 1-1.872-.878.075.075 0 0 1-.008-.125c.126-.093.252-.19.372-.287a.075.075 0 0 1 .078-.01c3.927 1.764 8.18 1.764 12.061 0a.075.075 0 0 1 .079.009c.12.098.245.195.372.288a.075.075 0 0 1-.006.125c-.598.344-1.22.635-1.873.877a.075.075 0 0 0-.041.105c.36.687.772 1.341 1.225 1.962a.077.077 0 0 0 .084.028 19.963 19.963 0 0 0 6.002-2.981.076.076 0 0 0 .032-.054c.5-5.094-.838-9.52-3.549-13.442a.06.06 0 0 0-.031-.028zM8.02 15.278c-1.182 0-2.157-1.069-2.157-2.38 0-1.312.956-2.38 2.157-2.38 1.21 0 2.176 1.077 2.157 2.38 0 1.312-.956 2.38-2.157 2.38zm7.975 0c-1.183 0-2.157-1.069-2.157-2.38 0-1.312.955-2.38 2.157-2.38 1.21 0 2.176 1.077 2.157 2.38 0 1.312-.946 2.38-2.157 2.38z"
                />
              </svg>
              <div class="flex-1">
                <span>{{ t('about.contact.discord') }}</span>
                <a
                  href="https://discord.gg/py9RhemV"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-blue-600 hover:text-blue-800 hover:underline"
                >
                  discord.gg/py9RhemV
                </a>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSEO } from '@/composables/useSEO'
import BreadcrumbNavigation from '@/components/BreadcrumbNavigation.vue'

const { t } = useI18n({
  useScope: 'global',
})
const { setPageSEO } = useSEO()

onMounted(() => {
  setPageSEO({
    title: 'About TibiaCores - Tibia Soulcore Management',
    description: 'Learn about TibiaCores, the ultimate tool for managing and tracking your Tibia soulcore collection. Collaborate with friends and optimize your hunting strategies.',
    keywords: 'Tibia, soulcore, about, features, collaboration, hunting, management',
    canonical: `${window.location.origin}/about`
  })
})
</script>
