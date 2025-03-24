<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import RegisterSuggestion from '@/components/RegisterSuggestion.vue'
import axios from 'axios'

interface Character {
  id: string
  name: string
  world: string
}

const userStore = useUserStore()
const characters = ref<Character[]>([])
const loading = ref(false)
const error = ref('')

const fetchCharacters = async () => {
  try {
    loading.value = true
    const response = await axios.get(`/api/users/${userStore.userId}/characters`)
    characters.value = response.data
  } catch (err) {
    error.value = 'Failed to load characters'
    console.error('Error fetching characters:', err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (userStore.isAuthenticated) {
    fetchCharacters()
  }
})
</script>

<template>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div v-if="userStore.isAnonymous" class="mb-8">
      <RegisterSuggestion />
    </div>

    <div class="bg-white shadow overflow-hidden sm:rounded-lg">
      <div class="px-4 py-5 sm:px-6">
        <h3 class="text-lg leading-6 font-medium text-gray-900">Profile</h3>
        <p class="mt-1 max-w-2xl text-sm text-gray-500">Your personal information and characters</p>
      </div>
      <div class="border-t border-gray-200 px-4 py-5 sm:p-0">
        <dl class="sm:divide-y sm:divide-gray-200">
          <div class="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
            <dt class="text-sm font-medium text-gray-500">User ID</dt>
            <dd class="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{{ userStore.userId }}</dd>
          </div>
          <div class="py-4 sm:py-5 sm:px-6">
            <dt class="text-sm font-medium text-gray-500 mb-4">Characters</dt>
            <dd class="mt-1 text-sm text-gray-900">
              <div v-if="loading" class="text-gray-500">Loading characters...</div>
              <div v-else-if="error" class="text-red-500">{{ error }}</div>
              <div v-else-if="characters.length === 0" class="text-gray-500 italic">
                No characters added yet
              </div>
              <ul v-else class="divide-y divide-gray-200">
                <li v-for="character in characters" :key="character.id" class="py-3">
                  <div class="flex items-center justify-between">
                    <div>
                      <router-link 
                        :to="{ name: 'character-details', params: { id: character.id }}"
                        class="text-sm font-medium text-indigo-600 hover:text-indigo-800"
                      >
                        {{ character.name }}
                      </router-link>
                      <p class="text-sm text-gray-500">{{ character.world }}</p>
                    </div>
                  </div>
                </li>
              </ul>
            </dd>
          </div>
        </dl>
      </div>
    </div>
  </div>
</template>
