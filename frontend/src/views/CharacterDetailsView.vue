<template>
  <div class="max-w-4xl mx-auto py-6 px-4">
    <div v-if="character" class="bg-white shadow rounded-lg p-6">
      <div class="mb-6">
        <h1 class="text-2xl font-bold">{{ character.name }}</h1>
        <p class="text-gray-600">{{ character.world }}</p>
      </div>

      <div class="mb-6">
        <h2 class="text-xl font-semibold mb-3">Unlocked Soul Cores</h2>
        <div v-if="unlockedCores.length > 0" class="grid grid-cols-2 md:grid-cols-3 gap-4">
          <div v-for="core in unlockedCores" :key="core.creature_id" class="p-3 bg-gray-50 rounded">
            <span class="font-medium">{{ core.creature_name }}</span>
          </div>
        </div>
        <p v-else class="text-gray-500">No soul cores unlocked yet.</p>
      </div>

      <!-- Soul Core Suggestions component -->
      <SoulcoreSuggestions :character-id="characterId" @suggestion-accepted="loadUnlockedCores" />
    </div>
    <div v-else class="text-center py-8">
      <p class="text-gray-500">Loading character details...</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'
import SoulcoreSuggestions from '@/components/SoulcoreSuggestions.vue'

const route = useRoute()
const characterId = route.params.id as string

interface Character {
  id: string
  name: string
  world: string
}

interface UnlockedCore {
  creature_id: string
  creature_name: string
}

const character = ref<Character | null>(null)
const unlockedCores = ref<UnlockedCore[]>([])

const loadCharacterDetails = async () => {
  try {
    const response = await axios.get(`/api/characters/${characterId}`)
    character.value = response.data
  } catch (error) {
    console.error('Failed to load character details:', error)
  }
}

const loadUnlockedCores = async () => {
  try {
    const response = await axios.get(`/api/characters/${characterId}/soulcores`)
    unlockedCores.value = response.data
  } catch (error) {
    console.error('Failed to load unlocked cores:', error)
  }
}

onMounted(() => {
  loadCharacterDetails()
  loadUnlockedCores()
})
</script>