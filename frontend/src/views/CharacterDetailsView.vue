<template>
  <div class="max-w-6xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <div v-if="loading" class="flex items-center justify-center min-h-[400px]">
      <div class="flex flex-col items-center space-y-4">
        <div
          class="animate-spin h-8 w-8 border-4 border-indigo-500 border-t-transparent rounded-full"
        ></div>
        <p class="text-gray-600">{{ t('characterDetails.soulcores.loading') }}</p>
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
            <div class="flex items-center space-x-4">
              <button
                @click="showShareDialog = true"
                class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
              >
                <svg class="h-5 w-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"
                  />
                </svg>
                {{ t('characterDetails.shareButton') }}
              </button>
              <router-link
                to="/profile"
                class="text-gray-600 hover:text-gray-900 flex items-center space-x-2 px-4 py-2 rounded-lg hover:bg-gray-100"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  class="h-5 w-5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M10 19l-7-7m0 0l7-7m-7 7h18"
                  />
                </svg>
                <span>{{ t('profile.title') }}</span>
              </router-link>
            </div>
          </div>
        </div>

        <!-- Stats Overview -->
        <div class="px-6 py-5 bg-gray-50 border-b border-gray-200">
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <p class="text-sm font-medium text-gray-500">
                {{ t('characterDetails.soulcores.total') }}
              </p>
              <p class="mt-1 text-2xl font-semibold text-gray-900">{{ unlockedCores.length }}</p>
            </div>
            <div>
              <p class="text-sm font-medium text-gray-500">{{ t('listDetail.xpBoostProgress') }}</p>
              <p class="mt-1 text-2xl font-semibold text-gray-900">
                {{ xpBoostProgress.current }}/{{ xpBoostProgress.target }}
              </p>
              <div class="mt-2 w-full bg-gray-200 rounded-full h-2">
                <div
                  class="bg-blue-600 rounded-full h-2"
                  :style="{ width: xpBoostProgress.percentage + '%' }"
                ></div>
              </div>
            </div>
            <div>
              <p class="text-sm font-medium text-gray-500">Total Progress</p>
              <p class="mt-1 text-2xl font-semibold text-gray-900">
                {{ totalProgress.current }}/{{ totalProgress.total }}
              </p>
              <div class="mt-2 w-full bg-gray-200 rounded-full h-2">
                <div
                  class="bg-green-600 rounded-full h-2"
                  :style="{ width: totalProgress.percentage + '%' }"
                ></div>
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

      <!-- Soul Cores Section -->
      <div class="bg-white shadow rounded-lg overflow-hidden">
        <div class="px-6 py-5 border-b border-gray-200">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <h2 class="text-xl font-semibold text-gray-900">
              {{ t('characterDetails.soulcores.title') }}
            </h2>
            <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-4">
              <div class="w-full sm:w-80">
                <CreatureSelect
                  v-model="selectedCreatureName"
                  :creatures="creatures"
                  :existing-soul-cores="
                    unlockedCores.map((core) => ({
                      creature_id: core.creature_id,
                      creature_name: core.creature_name,
                      status: 'obtained',
                      added_by: '',
                      added_by_user_id: '',
                    }))
                  "
                />
              </div>
              <button
                @click="addSoulcore"
                :disabled="!selectedCreatureName"
                class="w-full sm:w-auto min-w-[120px] bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {{ t('characterDetails.soulcores.addButton') }}
              </button>
            </div>
          </div>
        </div>

        <div class="px-6 py-6">
          <div v-if="unlockedCores.length === 0" class="text-center py-8">
            <p class="text-gray-500 text-lg">{{ t('characterDetails.soulcores.empty') }}</p>
            <p class="text-gray-400 mt-2">{{ t('profile.lists.empty') }}</p>
          </div>

          <div
            v-if="unlockedCores.length > 0"
            class="mt-6 grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4"
          >
            <div
              v-for="core in sortedUnlockedCores"
              :key="core.creature_id"
              class="group relative bg-gray-50 rounded-lg p-4 border border-gray-200 hover:border-red-200"
            >
              <div class="flex justify-between items-start">
                <div class="space-y-2">
                  <h3 class="font-medium text-gray-900">{{ core.creature_name }}</h3>
                </div>
                <button
                  @click="removeSoulcore(core.creature_id)"
                  class="opacity-0 group-hover:opacity-100 transition-opacity duration-200 text-red-600 hover:text-red-800 p-1 hover:bg-red-50 rounded"
                  :title="t('characterDetails.soulcores.removeButton')"
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    class="h-5 w-5"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                    />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Share Dialog -->
      <Dialog :open="showShareDialog" @close="showShareDialog = false">
        <div class="fixed inset-0 bg-black/50 flex items-center justify-center p-4">
          <DialogPanel class="bg-white rounded-xl p-6 max-w-lg w-full">
            <DialogTitle class="text-xl font-semibold mb-4">{{ t('characterDetails.shareCharacter') }}</DialogTitle>
            <p class="text-gray-600 mb-4">
              {{ t('characterDetails.shareCharacterDescription') }}
            </p>
            <div class="flex gap-2">
              <input
                type="text"
                readonly
                :value="shareUrl"
                class="flex-1 p-2 border border-gray-300 rounded-lg bg-gray-50"
              />
              <button
                @click="copyShareUrl"
                class="relative px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
              >
                <span
                  v-if="showCopiedMessage"
                  class="absolute -top-8 left-1/2 transform -translate-x-1/2 bg-black text-white px-2 py-1 rounded text-sm"
                >
                  {{ t('characterDetails.copied') }}
                </span>
                {{ t('characterDetails.copyLink') }}
              </button>
            </div>
          </DialogPanel>
        </div>
      </Dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import axios from 'axios'
import SoulcoreSuggestions from '@/components/SoulcoreSuggestions.vue'
import CreatureSelect from '@/components/CreatureSelect.vue'
import { Dialog, DialogPanel, DialogTitle } from '@headlessui/vue'

const route = useRoute()
const { t } = useI18n()
const characterId = route.params.id as string

interface Character {
  id: string
  name: string
  world: string
}

interface UnlockedCoreResponse {
  creature_id: string
  creature_name: string
}

interface UnlockedCore extends UnlockedCoreResponse {
  difficulty: number
}

interface Creature {
  id: string
  name: string
  difficulty: number
}

const character = ref<Character | null>(null)
const unlockedCores = ref<UnlockedCore[]>([])
const totalCreatures = ref(0)
const loading = ref(true)
const selectedCreatureName = ref('')
const creatures = ref<Array<{ id: string; name: string }>>([])
const showShareDialog = ref(false)
const showCopiedMessage = ref(false)
const sortOrder = ref<'asc' | 'desc'>('asc')

const sortedUnlockedCores = computed(() => {
  return [...unlockedCores.value].sort((a, b) => {
    return sortOrder.value === 'asc'
      ? a.creature_name.localeCompare(b.creature_name)
      : b.creature_name.localeCompare(a.creature_name)
  })
})

const shareUrl = computed(() => {
  return `${window.location.origin}/characters/public/${character.value?.name}`
})

const getSelectedCreature = computed(() => {
  return creatures.value.find((c) => c.name === selectedCreatureName.value)
})

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
      axios.get('/creatures'),
    ])

    const creaturesData = creaturesResponse.data as Creature[]
    const creatureMap = new Map(creaturesData.map((c) => [c.name, c]))

    unlockedCores.value = soulcoresResponse.data.map((core: UnlockedCoreResponse) => ({
      ...core,
      difficulty: creatureMap.get(core.creature_name)?.difficulty ?? 0,
    }))
    totalCreatures.value = creaturesData.length
    creatures.value = creaturesData
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

const addSoulcore = async () => {
  const creature = getSelectedCreature.value
  if (!creature) return

  try {
    await axios.post(`/characters/${characterId}/soulcores`, {
      creature_id: creature.id,
    })
    await loadUnlockedCores()
    selectedCreatureName.value = ''
  } catch (err) {
    console.error('Failed to add soul core:', err)
  }
}

const copyShareUrl = async () => {
  try {
    await navigator.clipboard.writeText(shareUrl.value)
    showCopiedMessage.value = true
    setTimeout(() => {
      showCopiedMessage.value = false
    }, 2000)
  } catch (err) {
    console.error('Failed to copy URL:', err)
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

<style scoped></style>
