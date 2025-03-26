<template>
  <div class="max-w-6xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <div v-if="loading" class="flex items-center justify-center min-h-[400px]">
      <div class="flex flex-col items-center space-y-4">
        <div class="animate-spin h-8 w-8 border-4 border-indigo-500 border-t-transparent rounded-full"></div>
        <p class="text-gray-600">Loading character details...</p>
      </div>
    </div>

    <div v-else-if="character" class="space-y-8">
      <!-- Character Header -->
      <div class="bg-white shadow rounded-lg overflow-hidden">
        <div class="px-6 py-8 border-b border-gray-200">
          <div class="flex items-center justify-between">
            <div>
              <h1 class="text-3xl font-bold text-gray-900">{{ character.name }}</h1>
              <p class="mt-1 text-lg text-gray-600">{{ character.world }}</p>
            </div>
            <router-link 
              to="/profile"
              class="text-gray-600 hover:text-gray-900 flex items-center space-x-2 px-4 py-2 rounded-lg hover:bg-gray-100"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
              <span>Back to Profile</span>
            </router-link>
          </div>
        </div>

        <!-- Stats Overview -->
        <div class="px-6 py-5 bg-gray-50 border-b border-gray-200">
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <p class="text-sm font-medium text-gray-500">Total Soul Cores</p>
              <p class="mt-1 text-2xl font-semibold text-gray-900">{{ unlockedCores.length }}</p>
            </div>
            <div>
              <p class="text-sm font-medium text-gray-500">XP Boost Progress (200)</p>
              <p class="mt-1 text-2xl font-semibold text-gray-900">{{ xpBoostProgress.current }}/{{ xpBoostProgress.target }}</p>
              <div class="mt-2 w-full bg-gray-200 rounded-full h-2">
                <div class="bg-blue-600 rounded-full h-2" :style="{ width: xpBoostProgress.percentage + '%' }"></div>
              </div>
            </div>
            <div>
              <p class="text-sm font-medium text-gray-500">Total Progress</p>
              <p class="mt-1 text-2xl font-semibold text-gray-900">{{ totalProgress.current }}/{{ totalProgress.total }}</p>
              <div class="mt-2 w-full bg-gray-200 rounded-full h-2">
                <div class="bg-green-600 rounded-full h-2" :style="{ width: totalProgress.percentage + '%' }"></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Soul Cores Section -->
      <div class="bg-white shadow rounded-lg overflow-hidden">
        <div class="px-6 py-5 border-b border-gray-200">
          <h2 class="text-xl font-semibold text-gray-900">Unlocked Soul Cores</h2>
        </div>

        <div class="px-6 py-6">
          <div v-if="unlockedCores.length === 0" class="text-center py-8">
            <p class="text-gray-500 text-lg">No soul cores unlocked yet</p>
            <p class="text-gray-400 mt-2">Join or create a list to start tracking your soul cores</p>
          </div>

          <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <div 
              v-for="core in unlockedCores" 
              :key="core.creature_id" 
              class="group relative bg-gray-50 rounded-lg p-4 border border-gray-200 hover:border-red-200"
            >
              <div class="flex justify-between items-start">
                <h3 class="font-medium text-gray-900">{{ core.creature_name }}</h3>
                <button
                  @click="removeSoulcore(core.creature_id)"
                  class="opacity-0 group-hover:opacity-100 transition-opacity duration-200 text-red-600 hover:text-red-800 p-1 hover:bg-red-50 rounded"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Soul Core Suggestions Section -->
      <SoulcoreSuggestions 
        :character-id="characterId" 
        @suggestion-accepted="loadUnlockedCores" 
        class="bg-white shadow rounded-lg overflow-hidden"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
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
const totalCreatures = ref(0)
const loading = ref(true)

const xpBoostProgress = computed(() => {
  const progress = unlockedCores.value.length
  const target = 200
  const percentage = Math.min((progress / target) * 100, 100)
  return {
    current: progress,
    target,
    percentage
  }
})

const totalProgress = computed(() => {
  const progress = unlockedCores.value.length
  const percentage = (progress / totalCreatures.value) * 100
  return {
    current: progress,
    total: totalCreatures.value,
    percentage
  }
})

const loadCharacterDetails = async () => {
  try {
    const response = await axios.get(`/characters/${characterId}`)
    character.value = response.data
  } catch (error) {
    console.error('Failed to load character details:', error)
  }
}

const loadUnlockedCores = async () => {
  try {
    const [soulcoresResponse, creaturesResponse] = await Promise.all([
      axios.get(`/characters/${characterId}/soulcores`),
      axios.get('/creatures')
    ])
    unlockedCores.value = soulcoresResponse.data
    totalCreatures.value = creaturesResponse.data.length
  } catch (error) {
    console.error('Failed to load unlocked cores:', error)
  }
}

const removeSoulcore = async (creatureId: string) => {
  try {
    await axios.delete(`/characters/${characterId}/soulcores/${creatureId}`)
    loadUnlockedCores()
  } catch (error) {
    console.error('Failed to remove soul core:', error)
  }
}

onMounted(async () => {
  try {
    await Promise.all([loadCharacterDetails(), loadUnlockedCores()])
  } catch (error) {
    console.error('Failed to load data:', error)
  } finally {
    loading.value = false
  }
})
</script>