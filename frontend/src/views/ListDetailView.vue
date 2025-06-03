<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import axios from 'axios'
import CreatureSelect from '@/components/CreatureSelect.vue'
import { useUserStore } from '@/stores/user'
import { useChatNotificationsStore } from '@/stores/chatNotifications'

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

interface ChatMessage {
  id: string
  list_id: string
  user_id: string
  character_name: string
  message: string
  created_at: string
}

interface Character {
  id: string
  name: string
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
const isChatOpen = ref(false)
const unreadChatCount = ref(0)
const messages = ref<ChatMessage[]>([])
const newMessage = ref('')
const chatLoading = ref(false)
const chatError = ref('')
const characters = ref<Character[]>([])
const selectedCharacterId = ref('')
const pollingInterval = ref<number | null>(null)
const lastMessageTimestamp = ref('')
const chatContainer = ref<HTMLElement | null>(null)

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
const chatNotificationsStore = useChatNotificationsStore()

// Get current user's characters
const fetchUserCharacters = async () => {
  try {
    const response = await axios.get<Character[]>(`/users/${userStore.userId}/characters`)
    characters.value = response.data

    // Auto-select the first character that is a member of this list
    if (characters.value.length > 0) {
      selectedCharacterId.value = characters.value[0].id
    }
  } catch (err) {
    if (axios.isAxiosError(err)) {
      chatError.value = err.response?.data?.message || 'Failed to fetch characters'
    } else {
      chatError.value = 'Failed to fetch characters'
    }
  }
}

// Fetch chat messages for the list
const fetchMessages = async (isInitial = false) => {
  try {
    chatLoading.value = true
    let url = `/lists/${props.id}/chat`

    if (!isInitial && lastMessageTimestamp.value) {
      url += `?since=${encodeURIComponent(lastMessageTimestamp.value)}`
    }

    const response = await axios.get<ChatMessage[]>(url)

    if (isInitial) {
      messages.value = response.data.reverse() // Reverse for chronological order
    } else if (response.data.length > 0) {
      // Add only new messages
      messages.value = [...messages.value, ...response.data]
    }

    // Update timestamp for polling
    if (messages.value.length > 0) {
      lastMessageTimestamp.value = messages.value[messages.value.length - 1].created_at
    }

    chatLoading.value = false

    // Mark messages as read when viewed
    if (messages.value.length > 0 && isChatOpen.value) {
      chatNotificationsStore.markAsRead(props.id)
      unreadChatCount.value = 0
    }
  } catch (err) {
    chatLoading.value = false
    if (axios.isAxiosError(err)) {
      chatError.value = err.response?.data?.message || 'Failed to fetch messages'
    } else {
      chatError.value = 'Failed to fetch messages'
    }
  }
}

// Send a new message
const sendMessage = async () => {
  if (!newMessage.value.trim() || !selectedCharacterId.value) return

  try {
    await axios.post(`/lists/${props.id}/chat`, {
      message: newMessage.value.trim(),
      character_id: selectedCharacterId.value
    })

    newMessage.value = ''
    await fetchMessages() // Refresh messages
  } catch (err) {
    if (axios.isAxiosError(err)) {
      chatError.value = err.response?.data?.message || 'Failed to send message'
    } else {
      chatError.value = 'Failed to send message'
    }
  }
}

// Delete a message (only own messages)
const deleteMessage = async (messageId: string) => {
  try {
    await axios.delete(`/lists/${props.id}/chat/${messageId}`)
    // Remove the message from the local list
    messages.value = messages.value.filter(msg => msg.id !== messageId)
  } catch (err) {
    if (axios.isAxiosError(err)) {
      chatError.value = err.response?.data?.message || 'Failed to delete message'
    } else {
      chatError.value = 'Failed to delete message'
    }
  }
}

// Set up message polling for real-time updates
const setupPolling = () => {
  pollingInterval.value = window.setInterval(async () => {
    if (isChatOpen.value) { // Only poll when chat is open
      await fetchMessages()
    }
  }, 5000) // Poll every 5 seconds
}

// Check if message is from current user
const isOwnMessage = (message: ChatMessage) => {
  return message.user_id === userStore.userId
}

// Format the timestamp to display only time if it's today, otherwise date and time
const formatTimestamp = (timestamp: string) => {
  const date = new Date(timestamp)
  const now = new Date()

  if (
    date.getDate() === now.getDate() &&
    date.getMonth() === now.getMonth() &&
    date.getFullYear() === now.getFullYear()
  ) {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  return date.toLocaleString([], {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// Watch for unread chat messages in this list
watch(() => chatNotificationsStore.notifications, (notifications) => {
  const listNotifications = notifications.filter((n: { list_id: string }) => n.list_id === props.id)
  unreadChatCount.value = listNotifications.length
}, { deep: true, immediate: true })

// Mark messages as read when chat is opened
watch(() => isChatOpen.value, (isOpen: boolean) => {
  if (isOpen) {
    if (unreadChatCount.value > 0) {
      chatNotificationsStore.markAsRead(props.id)
      unreadChatCount.value = 0
    }
    fetchMessages(true) // Load initial messages when opening chat
  }
})

// Auto-scroll to bottom when new messages arrive
watch(() => messages.value.length, () => {
  setTimeout(() => {
    if (chatContainer.value) {
      chatContainer.value.scrollTop = chatContainer.value.scrollHeight
    }
  }, 50)
})

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

    // Get chat notifications for this list
    if (userStore.isAuthenticated) {
      await chatNotificationsStore.fetchChatNotifications()
      await fetchUserCharacters()
      setupPolling()
    }
  } catch (err) {
    console.error('Error loading data:', err)
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  if (pollingInterval.value !== null) {
    clearInterval(pollingInterval.value)
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

      <!-- Chat window is now implemented as a floating bubble -->

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
      class="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-40"
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

  <!-- Floating Chat Bubble -->
  <div class="fixed bottom-6 right-6 z-30">
    <!-- Chat Bubble when closed -->
    <button
      v-if="!isChatOpen"
      @click="isChatOpen = true"
      class="bg-indigo-600 hover:bg-indigo-700 text-white rounded-full p-3 shadow-lg flex items-center justify-center relative transition-transform hover:scale-110"
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
      </svg>

      <!-- Unread messages indicator -->
      <span
        v-if="unreadChatCount > 0"
        class="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full h-5 w-5 flex items-center justify-center animate-pulse"
      >
        {{ unreadChatCount }}
      </span>
    </button>

    <!-- Chat Panel when open -->
    <div
      v-else
      class="bg-white border border-gray-200 rounded-lg shadow-xl w-80 md:w-96 h-[450px] flex flex-col animate-slide-up overflow-hidden"
    >
      <div class="p-3 border-b border-gray-200 flex justify-between items-center bg-indigo-50">
        <h3 class="font-semibold flex items-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-indigo-600" viewBox="0 0 20 20" fill="currentColor">
            <path d="M2 5a2 2 0 012-2h7a2 2 0 012 2v4a2 2 0 01-2 2H9l-3 3v-3H4a2 2 0 01-2-2V5z" />
            <path d="M15 7v2a4 4 0 01-4 4H9.828l-1.766 1.767c.28.149.599.233.938.233h2l3 3v-3h2a2 2 0 002-2V9a2 2 0 00-2-2h-1z" />
          </svg>
          {{ t('listDetail.chat.title') }}
        </h3>
        <button @click="isChatOpen = false" class="text-gray-500 hover:text-gray-700">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
          </svg>
        </button>
      </div>

      <!-- Chat content -->
      <div class="flex-1 overflow-y-auto p-3 space-y-2 bg-gray-50" ref="chatContainer">
        <div v-if="chatLoading && !messages.length" class="flex justify-center items-center h-full">
          <div class="animate-spin h-5 w-5 border-2 border-indigo-500 border-t-transparent rounded-full"></div>
          <span class="ml-2 text-gray-600">{{ t('listDetail.chat.loading') }}</span>
        </div>

        <div v-else-if="!messages.length" class="flex justify-center items-center h-full">
          <p class="text-gray-500">{{ t('listDetail.chat.noMessages') }}</p>
        </div>

        <div
          v-for="message in messages"
          :key="message.id"
          class="p-2 rounded-lg break-words text-sm"
          :class="[
            isOwnMessage(message)
              ? 'bg-indigo-100 text-indigo-900 ml-auto max-w-[80%]'
              : 'bg-white border border-gray-200 max-w-[80%]'
          ]"
        >
          <div class="flex justify-between items-start mb-1">
            <span class="font-medium text-xs">{{ message.character_name }}</span>
            <span class="text-xs text-gray-500">{{ formatTimestamp(message.created_at) }}</span>
          </div>
          <p>{{ message.message }}</p>
          <div v-if="isOwnMessage(message)" class="mt-1 flex justify-end">
            <button
              @click="deleteMessage(message.id)"
              class="text-xs text-red-600 hover:text-red-800"
            >
              {{ t('listDetail.chat.delete') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Input area -->
      <div class="p-3 border-t border-gray-200 bg-white">
        <div v-if="characters.length > 0" class="flex items-center mb-2">
          <span class="text-xs text-gray-600">
            {{ t('listDetail.chat.chatAs') }} <span class="font-medium">{{ characters[0]?.name }}</span>
          </span>
        </div>
        <div class="flex gap-2">
          <input
            v-model="newMessage"
            type="text"
            :placeholder="t('listDetail.chat.typingMessage')"
            class="flex-1 p-2 border border-gray-300 rounded-l-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
            @keyup.enter="sendMessage"
          />
          <button
            @click="sendMessage"
            class="px-4 py-2 bg-indigo-600 text-white rounded-r-lg hover:bg-indigo-700 transition-colors"
            :disabled="!newMessage.trim() || !selectedCharacterId"
          >
            {{ t('listDetail.chat.send') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Chat bubble animations */
@keyframes slide-up {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-slide-up {
  animation: slide-up 0.3s ease forwards;
}

@keyframes pulse {
  0% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
  100% {
    transform: scale(1);
  }
}

.animate-pulse {
  animation: pulse 1.5s infinite;
}
</style>
