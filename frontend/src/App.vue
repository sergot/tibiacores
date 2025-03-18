<script setup lang="ts">
import Navbar from './components/Navbar.vue'
import Footer from './components/Footer.vue'
</script>

<template>
  <div class="flex flex-col min-h-screen">
    <Navbar />
    
    <main class="flex-grow bg-gray-100">
      <RouterView />
    </main>

    <Footer />
  </div>
</template>
