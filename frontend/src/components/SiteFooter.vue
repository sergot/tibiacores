<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <footer class="bg-gray-800 text-white py-4">
    <div class="container mx-auto px-4">
      <div class="flex flex-col md:flex-row justify-between items-center">
        <div class="mb-4 md:mb-0">
          <p class="text-sm">
            &copy; {{ new Date().getFullYear() }} TibiaCores. {{ t('footer.rights') }}
          </p>
          <p class="text-sm mt-1">
            {{
              t(
                'footer.tibiaCopyright',
                'Tibia and all products related to Tibia are copyrighted by',
              )
            }}
            <a
              href="https://www.cipsoft.com/en/"
              target="_blank"
              rel="noopener noreferrer"
              class="text-blue-400 hover:underline"
              >CipSoft GmbH</a
            >.
            {{ t('footer.tibiaWebsite', 'The official website of Tibia is') }}
            <a
              href="https://www.tibia.com/news/?subtopic=latestnews"
              target="_blank"
              rel="noopener noreferrer"
              class="text-blue-400 hover:underline"
              >tibia.com</a
            >.
          </p>
          <p class="text-sm mt-1">
            <RouterLink to="/privacy" class="text-blue-400 hover:underline">
              {{ t('footer.privacyPolicy', 'Privacy Policy') }}
            </RouterLink>
          </p>
        </div>
        <div class="flex space-x-4">
          <RouterLink
            to="/sponsor"
            class="bg-pink-600 hover:bg-pink-700 text-white px-4 py-2 rounded-md text-sm font-medium"
          >
            {{ t('sponsor.support') || 'Support TibiaCores' }}
          </RouterLink>
        </div>
      </div>
    </div>
  </footer>
</template>
