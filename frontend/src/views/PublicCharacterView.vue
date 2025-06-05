<template>
  <div class="min-h-screen bg-gray-50">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <BreadcrumbNavigation />

      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center items-center h-64">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
        <p class="ml-4 text-gray-600">{{ t('publicCharacter.loading') }}</p>
      </div>

      <!-- Character Header -->
      <div v-else-if="character" class="space-y-8">
        <!-- Hero Section -->
        <div class="bg-white shadow-lg rounded-lg overflow-hidden">
          <div class="px-6 py-12 border-b border-gray-200 relative overflow-hidden">
            <!-- Animated gradient background -->
            <div
              class="absolute inset-0 animate-gradient-x"
              :style="{
                '--gradient-start': `hsl(${nameHash % 360}, 80%, 80%)`,
                '--gradient-mid': `hsl(${(nameHash + 120) % 360}, 80%, 80%)`,
                '--gradient-end': `hsl(${(nameHash + 240) % 360}, 80%, 80%)`,
                '--gradient-opacity': '0.3',
              }"
            ></div>

            <div class="relative z-10">
              <div class="text-center">
                <h1 class="text-4xl font-bold text-gray-900 animate-fade-in">
                  {{ character.name }}
                </h1>
                <p
                  class="mt-2 text-xl text-gray-600 animate-fade-in-delay flex items-center justify-center"
                >
                  <svg
                    class="w-5 h-5 mr-2 text-gray-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="1.5"
                      d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                  {{ character.world }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Stats Cards -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <!-- XP Boost Progress -->
          <div
            class="bg-white rounded-2xl shadow-lg p-6 hover:shadow-xl transition-shadow duration-300"
          >
            <div class="flex items-center mb-4">
              <div class="p-2 bg-indigo-100 rounded-lg mr-4">
                <svg
                  class="w-6 h-6 text-indigo-600"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="1.5"
                    d="M13 10V3L4 14h7v7l9-11h-7z"
                  />
                </svg>
              </div>
              <h3 class="text-lg font-semibold text-gray-900">
                {{ t('publicCharacter.stats.xpBoost.title') }}
              </h3>
            </div>
            <div class="space-y-4">
              <div class="flex justify-between text-sm text-gray-600">
                <span>{{
                  t('publicCharacter.stats.xpBoost.cores', {
                    current: xpBoostProgress.current,
                    target: xpBoostProgress.target,
                  })
                }}</span>
                <span>{{
                  t('publicCharacter.stats.xpBoost.percentage', {
                    percentage: Math.round(xpBoostProgress.percentage),
                  })
                }}</span>
              </div>
              <div class="h-4 bg-gray-200 rounded-full overflow-hidden">
                <div
                  class="h-full bg-gradient-to-r from-indigo-500 to-purple-500 transition-all duration-500 rounded-full"
                  :style="{ width: `${xpBoostProgress.percentage}%` }"
                ></div>
              </div>
            </div>
          </div>

          <!-- Total Progress -->
          <div
            class="bg-white rounded-2xl shadow-lg p-6 hover:shadow-xl transition-shadow duration-300"
          >
            <div class="flex items-center mb-4">
              <div class="p-2 bg-green-100 rounded-lg mr-4">
                <svg
                  class="w-6 h-6 text-green-600"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="1.5"
                    d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
              </div>
              <h3 class="text-lg font-semibold text-gray-900">
                {{ t('publicCharacter.stats.total.title') }}
              </h3>
            </div>
            <div class="space-y-4">
              <div class="flex justify-between text-sm text-gray-600">
                <span>{{
                  t('publicCharacter.stats.total.cores', {
                    current: totalProgress.current,
                    total: totalProgress.total,
                  })
                }}</span>
                <span>{{
                  t('publicCharacter.stats.total.percentage', {
                    percentage: Math.round(totalProgress.percentage),
                  })
                }}</span>
              </div>
              <div class="h-4 bg-gray-200 rounded-full overflow-hidden">
                <div
                  class="h-full bg-gradient-to-r from-green-500 to-emerald-500 transition-all duration-500 rounded-full"
                  :style="{ width: `${totalProgress.percentage}%` }"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Soul Cores Section -->
        <div class="bg-white rounded-2xl shadow-lg overflow-hidden">
          <div class="px-6 py-4 border-b border-gray-200">
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div class="flex items-center">
                <div class="p-2 bg-blue-100 rounded-lg mr-4">
                  <svg
                    class="w-6 h-6 text-blue-600"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="1.5"
                      d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                    />
                  </svg>
                </div>
                <h2 class="text-xl font-semibold text-gray-900">
                  {{ t('publicCharacter.soulCores.title') }}
                </h2>
              </div>
              <div class="flex flex-col sm:flex-row items-start sm:items-center gap-4">
                <div class="relative w-full sm:w-64">
                  <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <svg
                      class="h-5 w-5 text-gray-400"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="1.5"
                        d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                      />
                    </svg>
                  </div>
                  <input
                    v-model="searchQuery"
                    type="text"
                    class="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                    :placeholder="t('publicCharacter.soulCores.searchPlaceholder')"
                  />
                </div>
                <div class="text-sm text-gray-500">
                  {{ t('publicCharacter.soulCores.count', { count: filteredCores.length }) }}
                </div>
              </div>
            </div>
          </div>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200">
              <thead class="bg-gray-50">
                <tr>
                  <th
                    scope="col"
                    class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 transition-colors duration-200"
                    @click="toggleSort('name')"
                  >
                    {{ t('publicCharacter.soulCores.columns.name') }}
                    <span v-if="sortBy === 'name'" class="ml-1">
                      {{ sortOrder === 'asc' ? '↑' : '↓' }}
                    </span>
                  </th>
                  <th
                    scope="col"
                    class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 transition-colors duration-200"
                    @click="toggleSort('difficulty')"
                  >
                    {{ t('publicCharacter.soulCores.columns.difficulty') }}
                    <span v-if="sortBy === 'difficulty'" class="ml-1">
                      {{ sortOrder === 'asc' ? '↑' : '↓' }}
                    </span>
                  </th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-gray-200">
                <tr
                  v-for="core in filteredCores"
                  :key="core.creature_id"
                  class="hover:bg-gray-50 transition-colors duration-200"
                >
                  <td class="px-6 py-4 whitespace-nowrap">
                    <div class="text-sm font-medium text-gray-900">{{ core.creature_name }}</div>
                  </td>
                  <td class="px-6 py-4 whitespace-nowrap">
                    <span
                      class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium transition-colors duration-200"
                      :class="{
                        'bg-green-100 text-green-800 hover:bg-green-200': core.difficulty <= 1,
                        'bg-blue-100 text-blue-800 hover:bg-blue-200': core.difficulty === 2,
                        'bg-yellow-100 text-yellow-800 hover:bg-yellow-200': core.difficulty === 3,
                        'bg-orange-100 text-orange-800 hover:bg-orange-200': core.difficulty === 4,
                        'bg-red-100 text-red-800 hover:bg-red-200': core.difficulty === 5,
                      }"
                    >
                      {{ getDifficultyLabel(core.difficulty) }}
                    </span>
                  </td>
                </tr>
                <tr v-if="unlockedCores.length === 0">
                  <td colspan="2" class="px-6 py-8 text-center text-gray-500">
                    {{ t('publicCharacter.soulCores.empty') }}
                  </td>
                </tr>
              </tbody>
            </table>
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
import { useSEO } from '@/composables/useSEO'
import BreadcrumbNavigation from '@/components/BreadcrumbNavigation.vue'

const route = useRoute()
const { t } = useI18n()
const { setCharacterSEO } = useSEO()
const characterName = route.params.name as string

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
const sortOrder = ref<'asc' | 'desc'>('desc')
const sortBy = ref<'name' | 'difficulty'>('difficulty')
const searchQuery = ref('')

const totalProgress = computed(() => {
  const progress = unlockedCores.value.length
  const percentage = (progress / totalCreatures.value) * 100
  return {
    current: progress,
    total: totalCreatures.value,
    percentage,
  }
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

const sortedUnlockedCores = computed(() => {
  return [...unlockedCores.value].sort((a, b) => {
    if (sortBy.value === 'difficulty') {
      return sortOrder.value === 'desc' ? b.difficulty - a.difficulty : a.difficulty - b.difficulty
    } else {
      return sortOrder.value === 'desc'
        ? b.creature_name.localeCompare(a.creature_name)
        : a.creature_name.localeCompare(b.creature_name)
    }
  })
})

const filteredCores = computed(() => {
  if (!searchQuery.value) return sortedUnlockedCores.value

  const query = searchQuery.value.toLowerCase()
  return sortedUnlockedCores.value.filter((core) =>
    core.creature_name.toLowerCase().includes(query),
  )
})

const toggleSort = (column: 'name' | 'difficulty') => {
  if (sortBy.value === column) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = column
    sortOrder.value = 'asc'
  }
}

const getDifficultyLabel = (difficulty: number): string => {
  switch (difficulty) {
    case 0:
      return t('difficulty.harmless')
    case 1:
      return t('difficulty.trivial')
    case 2:
      return t('difficulty.easy')
    case 3:
      return t('difficulty.medium')
    case 4:
      return t('difficulty.hard')
    case 5:
      return t('difficulty.challenging')
    default:
      return t('difficulty.unknown')
  }
}

const loadCharacterDetails = async () => {
  try {
    const [characterResponse, creaturesResponse] = await Promise.all([
      axios.get(`/characters/public/${characterName}`),
      axios.get('/creatures'),
    ])

    const creatures = creaturesResponse.data as Creature[]
    const creatureMap = new Map(creatures.map((c) => [c.name, c]))

    character.value = characterResponse.data.character
    unlockedCores.value = characterResponse.data.unlocked_cores.map(
      (core: UnlockedCoreResponse) => ({
        ...core,
        difficulty: creatureMap.get(core.creature_name)?.difficulty ?? 0,
      }),
    )
    totalCreatures.value = creatures.length

    // Set SEO data for the character page
    if (character.value) {
      setCharacterSEO(
        character.value.name,
        character.value.world,
        unlockedCores.value.length
      )
    }
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

// Update name hash computation to be more meaningful
const nameHash = computed(() => {
  if (!character.value) return 0

  // Convert name to lowercase and remove spaces
  const name = character.value.name.toLowerCase().replace(/\s+/g, '')

  // Calculate a hash that's more visually meaningful
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    const char = name.charCodeAt(i)
    // Use different multipliers for vowels and consonants
    const multiplier = 'aeiou'.includes(name[i]) ? 3 : 2
    hash = (hash * multiplier + char) % 360
  }

  // Ensure the hash is positive and within a good range for colors
  return Math.abs(hash)
})
</script>

<style scoped>
@keyframes gradient-x {
  0% {
    background-position: 0% 50%;
  }
  50% {
    background-position: 100% 50%;
  }
  100% {
    background-position: 0% 50%;
  }
}

@keyframes fade-in {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-gradient-x {
  background: linear-gradient(
    90deg,
    var(--gradient-start) 0%,
    var(--gradient-mid) 50%,
    var(--gradient-end) 100%
  );
  background-size: 200% 200%;
  animation: gradient-x 12s ease infinite;
  opacity: var(--gradient-opacity);
}

.animate-fade-in {
  animation: fade-in 0.8s ease-out forwards;
}

.animate-fade-in-delay {
  animation: fade-in 0.8s ease-out 0.3s forwards;
  opacity: 0;
}

/* Mobile-specific styles */
@media (max-width: 640px) {
  .px-4 {
    padding-left: 1rem;
    padding-right: 1rem;
  }

  .text-4xl {
    font-size: 2rem;
  }

  .text-xl {
    font-size: 1.25rem;
  }

  .grid-cols-1 {
    grid-template-columns: 1fr;
  }

  .px-6 {
    padding-left: 1rem;
    padding-right: 1rem;
  }

  .py-12 {
    padding-top: 2rem;
    padding-bottom: 2rem;
  }

  .w-12 {
    width: 2.5rem;
  }

  .h-12 {
    height: 2.5rem;
  }

  .w-5 {
    width: 1.25rem;
  }

  .h-5 {
    height: 1.25rem;
  }

  .p-6 {
    padding: 1rem;
  }

  .text-lg {
    font-size: 1rem;
  }

  .w-6 {
    width: 1.25rem;
  }

  .h-6 {
    height: 1.25rem;
  }

  .px-2\.5 {
    padding-left: 0.625rem;
    padding-right: 0.625rem;
  }

  .py-0\.5 {
    padding-top: 0.125rem;
    padding-bottom: 0.125rem;
  }

  .text-xs {
    font-size: 0.75rem;
  }

  .px-6 {
    padding-left: 1rem;
    padding-right: 1rem;
  }

  .py-4 {
    padding-top: 0.75rem;
    padding-bottom: 0.75rem;
  }

  .py-3 {
    padding-top: 0.5rem;
    padding-bottom: 0.5rem;
  }

  .py-8 {
    padding-top: 1.5rem;
    padding-bottom: 1.5rem;
  }
}
</style>
