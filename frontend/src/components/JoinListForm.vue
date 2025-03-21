<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const shareCode = ref('')
const characterName = ref('')
const error = ref('')
const loading = ref(false)

const extractShareCode = (input: string): string => {
  try {
    const url = new URL(input)
    const parts = url.pathname.split('/')
    return parts[parts.length - 1]
  } catch {
    return input
  }
}

const handleSubmit = () => {
  const code = extractShareCode(shareCode.value)
  if (!code) {
    error.value = 'Please enter a valid share code or URL'
    return
  }

  // Navigate to the join list view with the character name and share code
  router.push({
    name: 'join-list',
    params: { share_code: code },
    query: characterName.value ? { character: characterName.value } : undefined
  })
}
</script>

<template>
  <div class="p-6 rounded-lg">
    <h2 class="mb-4 text-2xl">Join a list</h2>

    <div v-if="error" class="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <form @submit.prevent="handleSubmit" class="space-y-4">
      <div>
        <label for="shareCode" class="block text-sm font-medium text-gray-700 mb-1">
          Share Code or URL
        </label>
        <input
          id="shareCode"
          v-model="shareCode"
          type="text"
          placeholder="Enter share code or paste URL"
          required
          :disabled="loading"
          class="w-full p-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
        />
      </div>

      <div>
        <label for="characterName" class="block text-sm font-medium text-gray-700 mb-1">
          Character Name (Optional)
        </label>
        <input
          id="characterName"
          v-model="characterName"
          type="text"
          placeholder="Enter your character name"
          :disabled="loading"
          class="w-full p-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
        />
      </div>

      <button
        type="submit"
        :disabled="loading || !shareCode"
        class="w-full px-4 py-2 bg-green-500 text-white rounded-md hover:bg-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 disabled:bg-gray-400"
      >
        Continue
      </button>
    </form>
  </div>
</template>
