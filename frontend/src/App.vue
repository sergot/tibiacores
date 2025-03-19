<script setup lang="ts">
import SiteNavbar from './components/SiteNavbar.vue'
import SiteFooter from './components/SiteFooter.vue'
</script>

<template>
  <div class="flex flex-col min-h-screen">
    <SiteNavbar />

    <main class="flex-grow bg-gray-100">
      <RouterView />
    </main>

    <SiteFooter />
  </div>
</template>
