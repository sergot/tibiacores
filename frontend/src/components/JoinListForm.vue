<script setup lang="ts">
import { ref } from 'vue'

const shareCode = ref('')

const joinList = () => {
  // TODO: Implement join list logic
  console.log('Joining list with code:', shareCode.value)
  shareCode.value = ''
}
</script>

<template>
  <div class="p-6 rounded-lg">
    <h2 class="mb-4 text-2xl">Join a list</h2>
    <form @submit.prevent="joinList">
      <div class="flex gap-2">
        <input 
          v-model="shareCode"
          type="text"
          placeholder="Enter share code or paste URL"
          required
          class="flex-1 p-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
        >
        <button 
          type="submit"
          class="px-4 py-2 bg-green-500 text-white rounded-md hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2"
        >
          Join
        </button>
      </div>
    </form>
  </div>
</template>