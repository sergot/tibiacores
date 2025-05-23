<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import axios from 'axios'
import CreatureSelect from '@/components/CreatureSelect.vue'
import { useUserStore } from '@/stores/user'

interface ListDetails {
  id: string
  author_id: string
  name: string
  share_code: string
  world: string
  created_at: string
  updated_at: string
  members: MemberStats[]
  soul_cores: SoulCore[]
}

interface MemberStats {
  user_id: string
  character_name: string
  obtained_count: number
  unlocked_count: number
  is_active: boolean
}

interface SoulCore {
  creature_id: string
  creature_name: string
  status: 'obtained' | 'unlocked'
  added_by: string | null
  added_by_user_id: string
}

interface Creature {
  id: string
  name: string
}

interface UnlockStats {
  creature_id: string
  unlocked_count: number
  unlocked_by: Array<{
    character_name: string
    list_name: string
  }>
}

interface ListMemberWithUnlocks {
  user_id: string
  character_id: string
  character_name: string
  unlocked_creatures: Array<{
    creature_id: string
    creature_name: string
  }>
  obtained_count: number
  unlocked_count: number
  is_active: boolean
}

const props = defineProps<{
  id: string
}>()

const router = useRouter()
const listDetails = ref<ListDetails | null>(null)
const loading = ref(true)
const error = ref('')
const showShareDialog = ref(false)
const creatures = ref<Creature[]>([])
const selectedCreatureName = ref('')
const sortField = ref<'creature_name' | 'status'>('creature_name')
const sortDirection = ref<'asc' | 'desc'>('asc')
const searchQuery = ref('')
const hideUnlocked = ref(true)
const showCopiedMessage = ref(false)
const unlockStats = ref<Record<string, UnlockStats>>({})
const membersWithUnlocks = ref<ListMemberWithUnlocks[]>([])

// Add computed property for unlocked cores count
const unlockedCoresCount = computed(() => {
  if (!listDetails.value) return 0
  return listDetails.value.soul_cores.filter((core) => core.status === 'unlocked').length
})

// Add computed property for total creatures count
const totalCreaturesCount = computed(() => {
  return creatures.value.length
})

const userStore = useUserStore()

const getSelectedCreature = computed(() => {
  return creatures.value.find((c) => c.name === selectedCreatureName.value)
})

const availableCreatures = computed(() => {
  if (!creatures.value) return []
  return creatures.value
})

const fetchListDetails = async () => {
  try {
    const response = await axios.get<ListDetails>(`/lists/${props.id}`)
    listDetails.value = response.data
  } catch (err) {
    if (axios.isAxiosError(err)) {
      error.value = err.response?.data?.message || 'Failed to fetch list details'
    } else {
      error.value = 'Failed to fetch list details'
    }
  }
}

const fetchCreatures = async () => {
  try {
    const response = await axios.get<Creature[]>('/creatures')
    creatures.value = response.data

    // Build unlock stats from membersWithUnlocks data
    const stats: Record<string, UnlockStats> = {}
    membersWithUnlocks.value.forEach((member) => {
      member.unlocked_creatures.forEach((creature) => {
        if (!stats[creature.creature_id]) {
          stats[creature.creature_id] = {
            creature_id: creature.creature_id,
            unlocked_count: 0,
            unlocked_by: [],
          }
        }
        stats[creature.creature_id].unlocked_count++
        stats[creature.creature_id].unlocked_by.push({
          character_name: member.character_name,
          list_name: listDetails.value?.name || '',
        })
      })
    })
    unlockStats.value = stats
  } catch (err) {
    if (axios.isAxiosError(err)) {
      error.value = err.response?.data?.message || 'Failed to fetch creatures'
    } else {
      error.value = 'Failed to fetch creatures'
    }
  }
}

const fetchListMembers = async () => {
  try {
    const response = await axios.get<ListMemberWithUnlocks[]>(`/lists/${props.id}/members`)
    membersWithUnlocks.value = response.data
  } catch (err) {
    if (axios.isAxiosError(err)) {
      error.value = err.response?.data?.message || 'Failed to fetch list members'
    } else {
      error.value = 'Failed to fetch list members'
    }
  }
}

const addSoulcore = async () => {
  const creature = getSelectedCreature.value
  if (!creature) return

  try {
    await axios.post(`/lists/${props.id}/soulcores`, {
      creature_id: creature.id,
      status: 'obtained',
    })
    await fetchListDetails()
    selectedCreatureName.value = '' // Clear input after adding
  } catch (err) {
    if (axios.isAxiosError(err)) {
      error.value = err.response?.data?.message || 'Failed to add soul core'
    } else {
      error.value = 'Failed to add soul core'
    }
  }
}

const updateSoulcoreStatus = async (creatureId: string, status: 'obtained' | 'unlocked') => {
  try {
    await axios.put(`/lists/${props.id}/soulcores`, {
      creature_id: creatureId,
      status,
    })
    await fetchListDetails()
  } catch (err) {
    if (axios.isAxiosError(err)) {
      error.value = err.response?.data?.message || 'Failed to update soul core'
    } else {
      error.value = 'Failed to update soul core'
    }
  }
}

const removeSoulcore = async (creatureId: string) => {
  try {
    await axios.delete(`/lists/${props.id}/soulcores/${creatureId}`)
    await fetchListDetails()
  } catch (err) {
    if (axios.isAxiosError(err)) {
      error.value = err.response?.data?.message || 'Failed to remove soul core'
    } else {
      error.value = 'Failed to remove soul core'
    }
  }
}

const origin = window.location.origin

const copyShareLink = async () => {
  if (!listDetails.value) return

  const shareUrl = `${origin}/join/${listDetails.value.share_code}`
  try {
    await navigator.clipboard.writeText(shareUrl)
    showCopiedMessage.value = true
    setTimeout(() => {
      showCopiedMessage.value = false
    }, 2000)
  } catch {
    error.value = 'Failed to copy share link'
  }
}

const toggleSort = (field: 'creature_name' | 'status') => {
  if (sortField.value === field) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortField.value = field
    sortDirection.value = 'asc'
  }
}

const sortedAndFilteredSoulCores = computed(() => {
  if (!listDetails.value) return []

  let filtered = [...listDetails.value.soul_cores]

  // Apply search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter((core) => core.creature_name.toLowerCase().includes(query))
  }

  // Hide unlocked cores if enabled
  if (hideUnlocked.value) {
    filtered = filtered.filter((core) => core.status !== 'unlocked')
  }

  // Apply sorting
  return filtered.sort((a, b) => {
    const aValue = sortField.value === 'creature_name' ? a.creature_name : a.status
    const bValue = sortField.value === 'creature_name' ? b.creature_name : b.status
    const modifier = sortDirection.value === 'asc' ? 1 : -1

    return aValue > bValue ? modifier : -modifier
  })
})

const canModifySoulcore = (soulcore: SoulCore) => {
  return (
    soulcore.added_by_user_id === userStore.userId ||
    listDetails.value?.author_id === userStore.userId
  )
}

const { t } = useI18n()

onMounted(async () => {
  try {
    await fetchListDetails()
    await fetchListMembers()
    await fetchCreatures() // This needs to run after members are loaded
  } catch (err) {
    console.error('Error loading data:', err)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 py-8">
    <div v-if="error" class="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <div v-if="loading" class="text-center py-12">
      <div
        class="animate-spin h-8 w-8 border-4 border-blue-500 border-t-transparent rounded-full mx-auto mb-4"
      ></div>
      <p class="text-gray-600 font-medium">{{ t('characterDetails.soulcores.loading') }}</p>
    </div>

    <template v-else-if="listDetails">
      <div class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4 mb-8">
        <div>
          <div class="flex items-center gap-4 mb-2">
            <button
              @click="router.push('/')"
              class="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg"
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
            </button>
            <h1 class="text-2xl sm:text-3xl font-semibold">{{ listDetails.name }}</h1>
          </div>
          <p class="text-gray-600">
            {{ t('listDetail.created') }}
            {{ new Date(listDetails.created_at).toLocaleDateString() }}
          </p>
        </div>
        <button
          @click="showShareDialog = true"
          class="w-full sm:w-auto px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors flex items-center justify-center gap-2"
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
              d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"
            />
          </svg>
          {{ t('listDetail.shareList') }}
        </button>
      </div>

      <!-- Stats Section -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
        <!-- Soul Core Stats -->
        <div class="bg-white p-6 rounded-xl shadow-sm border border-gray-200">
          <h2 class="text-xl font-semibold mb-4">{{ t('characterDetails.soulcores.title') }}</h2>
          <div class="space-y-4">
            <div class="flex justify-between text-sm text-gray-600">
              <span>{{ t('listDetail.xpBoostProgress') }}</span>
              <span
                >{{
                  listDetails.soul_cores.filter(
                    (sc) => sc.status === 'obtained' || sc.status === 'unlocked',
                  ).length
                }}
                / 200</span
              >
            </div>
            <div class="w-full bg-gray-200 rounded-full h-2.5">
              <div
                class="bg-blue-600 h-2.5 rounded-full"
                :style="{
                  width: `${Math.min(
                    (listDetails.soul_cores.filter(
                      (sc) => sc.status === 'obtained' || sc.status === 'unlocked',
                    ).length /
                      200) *
                      100,
                    100,
                  )}%`,
                }"
              ></div>
            </div>
            <div class="flex justify-between text-sm text-gray-600">
              <span>{{ t('listDetail.totalProgress') }}</span>
              <span
                >{{
                  listDetails.soul_cores.filter(
                    (sc) => sc.status === 'obtained' || sc.status === 'unlocked',
                  ).length
                }}
                / {{ totalCreaturesCount }}</span
              >
            </div>
            <div class="w-full bg-gray-200 rounded-full h-2.5">
              <div
                class="bg-green-600 h-2.5 rounded-full"
                :style="{
                  width: `${(listDetails.soul_cores.filter((sc) => sc.status === 'obtained' || sc.status === 'unlocked').length / totalCreaturesCount) * 100}%`,
                }"
              ></div>
            </div>
          </div>
        </div>

        <!-- Member Stats -->
        <div class="bg-white p-6 rounded-xl shadow-sm border border-gray-200">
          <h2 class="text-xl font-semibold mb-4">{{ t('listDetail.memberContributions') }}</h2>
          <div class="space-y-4">
            <div
              v-for="member in listDetails.members"
              :key="member.user_id"
              class="flex flex-col gap-2 p-3 rounded-lg bg-gray-50"
            >
              <div class="flex justify-between items-center">
                <div class="flex items-center gap-1">
                  <span class="font-medium" :class="{ 'text-gray-400': !member.is_active }">
                    {{ member.character_name }}
                  </span>
                  <div v-if="!member.is_active" class="group relative">
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      class="h-4 w-4 text-gray-400"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                    <div
                      class="invisible group-hover:visible opacity-0 group-hover:opacity-100 transition-opacity absolute left-1/2 -translate-x-1/2 -top-2 transform -translate-y-full bg-gray-900 text-white text-xs rounded py-1 px-2 whitespace-nowrap"
                    >
                      {{ t('listDetail.inactiveCharacterTooltip') }}
                    </div>
                  </div>
                </div>
                <span
                  class="px-2 py-1 text-xs font-medium rounded-full"
                  :class="{
                    'bg-blue-100 text-blue-800': member.is_active,
                    'bg-gray-100 text-gray-600': !member.is_active,
                  }"
                >
                  obtained {{ member.obtained_count }} {{ t('listDetail.obtained') }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Soul Cores Table -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
        <div class="p-6 border-b border-gray-200">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-4">
            <div class="flex items-center gap-4">
              <h2 class="text-xl font-semibold">{{ t('characterDetails.soulcores.title') }}</h2>
              <button
                @click="hideUnlocked = !hideUnlocked"
                class="px-3 py-1.5 text-sm border rounded-lg hover:bg-gray-50 whitespace-nowrap"
                :class="
                  hideUnlocked
                    ? 'border-indigo-600 text-indigo-600'
                    : 'border-gray-300 text-gray-600'
                "
              >
                {{
                  hideUnlocked
                    ? t('listDetail.showUnlocked', { count: unlockedCoresCount })
                    : t('listDetail.hideUnlocked', { count: unlockedCoresCount })
                }}
              </button>
            </div>
            <div class="flex flex-col sm:flex-row sm:items-center gap-2">
              <CreatureSelect
                v-model="selectedCreatureName"
                :creatures="availableCreatures"
                :existing-soul-cores="listDetails?.soul_cores || []"
                :unlock-stats="unlockStats"
              />
              <button
                @click="addSoulcore"
                :disabled="!getSelectedCreature"
                class="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors disabled:bg-gray-400"
              >
                {{ t('characterDetails.soulcores.addButton') }}
              </button>
            </div>
          </div>

          <div class="mb-4">
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('characterDetails.soulcores.filters.search')"
              class="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
            />
          </div>

          <div class="overflow-x-auto">
            <table class="w-full">
              <thead>
                <tr class="border-b border-gray-200">
                  <th
                    @click="toggleSort('creature_name')"
                    class="px-2 sm:px-4 py-2 text-left text-sm font-medium text-gray-600 cursor-pointer hover:text-gray-900"
                  >
                    {{ t('listDetail.creatureName') }}
                    <span v-if="sortField === 'creature_name'">
                      {{ sortDirection === 'asc' ? '↑' : '↓' }}
                    </span>
                  </th>
                  <th
                    @click="toggleSort('status')"
                    class="hidden sm:table-cell px-4 py-2 text-left text-sm font-medium text-gray-600 cursor-pointer hover:text-gray-900"
                  >
                    {{ t('listDetail.status') }}
                    <span v-if="sortField === 'status'">
                      {{ sortDirection === 'asc' ? '↑' : '↓' }}
                    </span>
                  </th>
                  <th
                    class="hidden sm:table-cell px-4 py-2 text-left text-sm font-medium text-gray-600"
                  >
                    {{ t('listDetail.addedBy') }}
                  </th>
                  <th class="px-2 sm:px-4 py-2 text-right text-sm font-medium text-gray-600">
                    {{ t('listDetail.actions') }}
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="core in sortedAndFilteredSoulCores"
                  :key="core.creature_id"
                  class="border-b border-gray-200 last:border-0"
                >
                  <td class="px-2 sm:px-4 py-2">{{ core.creature_name }}</td>
                  <td class="hidden sm:table-cell px-4 py-2">
                    <span
                      :class="{
                        'px-2 py-1 text-xs font-medium rounded-full': true,
                        'bg-green-100 text-green-800': core.status === 'unlocked',
                        'bg-blue-100 text-blue-800': core.status === 'obtained',
                      }"
                    >
                      {{ t(core.status) }}
                    </span>
                  </td>
                  <td class="hidden sm:table-cell px-4 py-2 text-gray-600">
                    {{ core.added_by || '-' }}
                  </td>
                  <td class="px-2 sm:px-4 py-2">
                    <div class="flex items-center justify-end gap-2">
                      <button
                        v-if="canModifySoulcore(core) && core.status === 'obtained'"
                        @click="updateSoulcoreStatus(core.creature_id, 'unlocked')"
                        class="text-sm text-indigo-600 hover:text-indigo-800"
                      >
                        {{ t('listDetail.markAsUnlocked') }}
                      </button>
                      <button
                        v-if="canModifySoulcore(core) && core.status === 'unlocked'"
                        @click="updateSoulcoreStatus(core.creature_id, 'obtained')"
                        class="text-sm text-indigo-600 hover:text-indigo-800"
                      >
                        {{ t('listDetail.markAsObtained') }}
                      </button>
                      <button
                        v-if="canModifySoulcore(core)"
                        @click="removeSoulcore(core.creature_id)"
                        class="text-sm text-red-600 hover:text-red-800"
                      >
                        {{ t('characterDetails.soulcores.removeButton') }}
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </template>

    <!-- Share Dialog -->
    <div
      v-if="showShareDialog"
      class="fixed inset-0 bg-black/50 flex items-center justify-center p-4"
      @click="showShareDialog = false"
    >
      <div class="bg-white rounded-xl p-6 max-w-lg w-full" @click.stop>
        <h3 class="text-xl font-semibold mb-4">{{ t('listDetail.shareList') }}</h3>
        <p class="text-gray-600 mb-4">
          {{ t('listDetail.shareListDescription') }}
        </p>
        <div class="flex gap-2">
          <input
            type="text"
            readonly
            :value="`${origin}/join/${listDetails?.share_code}`"
            class="flex-1 p-2 border border-gray-300 rounded-lg bg-gray-50"
          />
          <button
            @click="copyShareLink"
            class="relative px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
          >
            <span
              v-if="showCopiedMessage"
              class="absolute -top-8 left-1/2 transform -translate-x-1/2 bg-black text-white px-2 py-1 rounded text-sm"
            >
              {{ t('listDetail.copied') }}
            </span>
            {{ t('listDetail.copyLink') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
/* Remove old multiselect styles */
</style>
