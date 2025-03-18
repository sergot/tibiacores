<script setup lang="ts">
import CreateListForm from '../components/CreateListForm.vue'
import JoinListForm from '../components/JoinListForm.vue'
</script>

<template>
  <div class="w-full">
    <main class="max-w-6xl mx-auto px-8 py-8">
      <div class="text-center mb-8">
        <h1 class="text-4xl mb-2">Welcome to Fiendlist</h1>
        <p class="text-xl text-gray-600">Your companion app for managing and sharing gaming lists</p>
      </div>
      
      <div class="grid md:grid-cols-2 gap-8 mb-8">
        <div class="p-6 rounded-lg shadow-sm bg-white">
          <CreateListForm />
        </div>
        <div class="p-6 rounded-lg shadow-sm bg-white">
          <JoinListForm />
        </div>
      </div>

      <div class="mt-8 p-6 rounded-lg bg-white text-center">
        <h2 class="mb-4 text-2xl">About Fiendlist</h2>
        <p class="text-gray-600 leading-relaxed">Fiendlist is your companion app for managing and sharing gaming lists. Create characters, build your collections, and share them with friends. Join existing lists using share codes or direct links.</p>
      </div>
    </main>
  </div>
</template>

<style scoped>
</style>
