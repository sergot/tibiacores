<template>
  <div class="min-h-screen bg-gray-100">
    <SiteNavbar />
    <router-view></router-view>
    <SiteFooter />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount } from 'vue'
import { useUserStore } from './stores/user'
import { useSuggestionsStore } from './stores/suggestions'
import SiteNavbar from './components/SiteNavbar.vue'
import SiteFooter from './components/SiteFooter.vue'

const userStore = useUserStore()
const suggestionsStore = useSuggestionsStore()

onMounted(() => {
  if (userStore.isAuthenticated) {
    suggestionsStore.startPolling()
  }
})

onBeforeUnmount(() => {
  if (userStore.isAuthenticated) {
    suggestionsStore.stopPolling()
  }
})
</script>
