<template>
  <div class="min-h-screen bg-gray-50">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center items-center h-64">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
        <p class="ml-4 text-gray-600">{{ t('publicCharacter.loading') }}</p>
      </div>

      <!-- Character Header -->
      <div v-else-if="character" class="space-y-8">
        <!-- Hero Section -->
        <div class="bg-white rounded-2xl shadow-lg overflow-hidden">
          <div class="relative h-48 bg-gradient-to-r from-indigo-600 to-purple-600">
            <div class="absolute inset-0 bg-black/20"></div>
            <div class="relative h-full flex items-center justify-center">
              <div class="text-center text-white">
                <h1 class="text-4xl font-bold mb-2">{{ character.name }}</h1>
                <p class="text-xl">{{ character.world }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Stats Cards -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <!-- XP Boost Progress -->
          <div class="bg-white rounded-2xl shadow-lg p-6">
            <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ t('publicCharacter.stats.xpBoost.title') }}</h3>
            <div class="space-y-4">
              <div class="flex justify-between text-sm text-gray-600">
                <span>{{ t('publicCharacter.stats.xpBoost.cores', { current: xpBoostProgress.current, target: xpBoostProgress.target }) }}</span>
                <span>{{ t('publicCharacter.stats.xpBoost.percentage', { percentage: Math.round(xpBoostProgress.percentage) }) }}</span>
              </div>
              <div class="h-4 bg-gray-200 rounded-full overflow-hidden">
                <div
                  class="h-full bg-gradient-to-r from-indigo-500 to-purple-500 transition-all duration-500"
                  :style="{ width: `${xpBoostProgress.percentage}%` }"
                ></div>
              </div>
            </div>
          </div>

          <!-- Total Progress -->
          <div class="bg-white rounded-2xl shadow-lg p-6">
            <h3 class="text-lg font-semibold text-gray-900 mb-4">{{ t('publicCharacter.stats.total.title') }}</h3>
            <div class="space-y-4">
              <div class="flex justify-between text-sm text-gray-600">
                <span>{{ t('publicCharacter.stats.total.cores', { current: totalProgress.current, total: totalProgress.total }) }}</span>
                <span>{{ t('publicCharacter.stats.total.percentage', { percentage: Math.round(totalProgress.percentage) }) }}</span>
              </div>
              <div class="h-4 bg-gray-200 rounded-full overflow-hidden">
                <div
                  class="h-full bg-gradient-to-r from-green-500 to-emerald-500 transition-all duration-500"
                  :style="{ width: `${totalProgress.percentage}%` }"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Soul Cores Section -->
        <div class="bg-white rounded-2xl shadow-lg overflow-hidden">
          <div class="px-6 py-4 border-b border-gray-200">
            <h2 class="text-xl font-semibold text-gray-900">{{ t('publicCharacter.soulCores.title') }}</h2>
          </div>
          <div class="divide-y divide-gray-200">
            <div
              v-for="core in unlockedCores"
              :key="core.creature_id"
              class="px-6 py-4 hover:bg-gray-50 transition-colors duration-150"
            >
              <div class="flex items-center justify-between">
                <span class="text-gray-900 font-medium">{{ core.creature_name }}</span>
                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                  {{ t('publicCharacter.soulCores.unlocked') }}
                </span>
              </div>
            </div>
            <div v-if="unlockedCores.length === 0" class="px-6 py-8 text-center text-gray-500">
              {{ t('publicCharacter.soulCores.empty') }}
            </div>
          </div>
        </div>
      </div>

      <!-- Error State -->
      <div v-else class="text-center py-12">
        <h3 class="text-lg font-medium text-gray-900">{{ t('publicCharacter.notFound.title') }}</h3>
        <p class="mt-2 text-gray-500">{{ t('publicCharacter.notFound.description') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import axios from 'axios'

const route = useRoute()
const { t } = useI18n()
const characterName = route.params.name as string

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
    percentage,
  }
})

const totalProgress = computed(() => {
  const progress = unlockedCores.value.length
  const percentage = (progress / totalCreatures.value) * 100
  return {
    current: progress,
    total: totalCreatures.value,
    percentage,
  }
})

const loadCharacterDetails = async () => {
  try {
    const [characterResponse, creaturesResponse] = await Promise.all([
      axios.get(`/characters/public/${characterName}`),
      axios.get('/creatures')
    ])
    character.value = characterResponse.data.character
    unlockedCores.value = characterResponse.data.unlocked_cores
    totalCreatures.value = creaturesResponse.data.length
  } catch (error) {
    console.error('Failed to load character details:', error)
  }
}

onMounted(async () => {
  try {
    await loadCharacterDetails()
  } catch (error) {
    console.error('Failed to load data:', error)
  } finally {
    loading.value = false
  }
})
</script> 