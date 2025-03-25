<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useUserStore } from '@/stores/user'
import RegisterSuggestion from '@/components/RegisterSuggestion.vue'
import axios from 'axios'

interface Character {
  id: string
  name: string
  world: string
  soulcore_count?: number
}

const userStore = useUserStore()
const characters = ref<Character[]>([])
const loading = ref(false)
const error = ref('')

const fetchCharacters = async () => {
  try {
    loading.value = true
    const response = await axios.get(`/api/users/${userStore.userId}/characters`)
    const chars = response.data
    
    // Fetch soulcore counts for each character
    const charsWithCounts = await Promise.all(chars.map(async (char: Character) => {
      const soulcores = await axios.get(`/api/characters/${char.id}/soulcores`)
      return {
        ...char,
        soulcore_count: soulcores.data.length
      }
    }))
    characters.value = charsWithCounts
  } catch (err) {
    error.value = 'Failed to load characters'
    console.error('Error fetching characters:', err)
  } finally {
    loading.value = false
  }
}

const characterWithMostCores = computed(() => {
  if (characters.value.length === 0) return null
  return characters.value.reduce((prev, current) => 
    (prev.soulcore_count || 0) > (current.soulcore_count || 0) ? prev : current
  )
})

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

    <div class="bg-white shadow sm:rounded-lg">
      <div class="px-4 py-5 sm:px-6 border-b border-gray-200">
        <h3 class="text-2xl font-bold text-gray-900">Profile</h3>
        <p class="mt-1 max-w-2xl text-sm text-gray-500">Your personal information and characters</p>
      </div>
      
      <div class="px-4 py-5 sm:p-6">
        <dl class="grid grid-cols-1 gap-6 sm:grid-cols-3">
          <div class="bg-gray-50 px-4 py-5 rounded-lg">
            <dt class="text-sm font-medium text-gray-500">User ID</dt>
            <dd class="mt-1 text-lg font-semibold text-gray-900">{{ userStore.userId }}</dd>
          </div>

          <div class="bg-gray-50 px-4 py-5 rounded-lg">
            <dt class="text-sm font-medium text-gray-500">Account Type</dt>
            <dd class="mt-1 text-lg font-semibold text-gray-900">
              {{ userStore.isAnonymous ? 'Anonymous' : 'Registered' }}
            </dd>
          </div>

          <div v-if="characterWithMostCores" class="bg-gray-50 px-4 py-5 rounded-lg">
            <dt class="text-sm font-medium text-gray-500">Most Advanced Character</dt>
            <dd class="mt-1 text-lg font-semibold text-gray-900">
              {{ characterWithMostCores.name }}
              <span class="text-sm text-gray-600">({{ characterWithMostCores.soulcore_count }} cores)</span>
            </dd>
          </div>
        </dl>

        <div class="mt-8">
          <h4 class="text-lg font-medium text-gray-900 mb-4">Your Characters</h4>
          <div v-if="loading" class="text-gray-500 flex items-center space-x-2">
            <div class="animate-spin h-5 w-5 border-2 border-gray-500 border-t-transparent rounded-full"></div>
            <span>Loading characters...</span>
          </div>
          <div v-else-if="error" class="text-red-500 bg-red-50 p-4 rounded-lg">{{ error }}</div>
          <div v-else-if="characters.length === 0" class="text-gray-500 italic bg-gray-50 p-4 rounded-lg">
            No characters added yet
          </div>
          <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <router-link 
              v-for="character in characters" 
              :key="character.id"
              :to="{ name: 'character-details', params: { id: character.id }}"
              class="block group"
            >
              <div class="bg-gray-50 p-4 rounded-lg border border-gray-200 hover:border-indigo-500 hover:shadow-md transition-all duration-200">
                <h5 class="font-medium text-gray-900 group-hover:text-indigo-600">{{ character.name }}</h5>
                <p class="text-sm text-gray-500">{{ character.world }}</p>
                <p class="text-sm text-gray-500 mt-1">{{ character.soulcore_count || 0 }} soul cores</p>
              </div>
            </router-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
