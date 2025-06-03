<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import { useUserStore } from '@/stores/user'
import { useI18n } from 'vue-i18n'
import { useChatNotificationsStore } from '@/stores/chatNotifications'
import axios from 'axios'

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
  listId: string
  isCompact?: boolean
}>()

const { t } = useI18n()
const userStore = useUserStore()
const chatNotificationsStore = useChatNotificationsStore()
const messages = ref<ChatMessage[]>([])
const newMessage = ref('')
const loading = ref(false)
const error = ref('')
const characters = ref<Character[]>([])
const selectedCharacterId = ref('')
const pollingInterval = ref<number | null>(null)
const lastMessageTimestamp = ref('')
const isExpanded = ref(!props.isCompact)

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
      error.value = err.response?.data?.message || 'Failed to fetch characters'
    } else {
      error.value = 'Failed to fetch characters'
    }
  }
}

// Fetch chat messages for the list
const fetchMessages = async (isInitial = false) => {
  try {
    loading.value = true
    let url = `/lists/${props.listId}/chat`

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

    loading.value = false

    // Mark messages as read when viewed
    if (messages.value.length > 0) {
      chatNotificationsStore.markAsRead(props.listId)
    }
  } catch (err) {
    loading.value = false
    if (axios.isAxiosError(err)) {
      error.value = err.response?.data?.message || 'Failed to fetch messages'
    } else {
      error.value = 'Failed to fetch messages'
    }
  }
}

// Send a new message
const sendMessage = async () => {
  if (!newMessage.value.trim() || !selectedCharacterId.value) return

  try {
    await axios.post(`/lists/${props.listId}/chat`, {
      message: newMessage.value.trim(),
      character_id: selectedCharacterId.value
    })

    newMessage.value = ''
    await fetchMessages() // Refresh messages
  } catch (err) {
    if (axios.isAxiosError(err)) {
      error.value = err.response?.data?.message || 'Failed to send message'
    } else {
      error.value = 'Failed to send message'
    }
  }
}

// Set up message polling for real-time updates
const setupPolling = () => {
  pollingInterval.value = window.setInterval(async () => {
    await fetchMessages()
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

// Delete a message (only own messages)
const deleteMessage = async (messageId: string) => {
  try {
    await axios.delete(`/lists/${props.listId}/chat/${messageId}`)
    // Remove the message from the local list
    messages.value = messages.value.filter(msg => msg.id !== messageId)
  } catch (err) {
    if (axios.isAxiosError(err)) {
      error.value = err.response?.data?.message || 'Failed to delete message'
    } else {
      error.value = 'Failed to delete message'
    }
  }
}

// Toggle expanded/collapsed state
const toggleExpanded = () => {
  isExpanded.value = !isExpanded.value
}

// Watch for prop changes to reload messages
watch(() => props.listId, async () => {
  if (props.listId) {
    messages.value = []
    lastMessageTimestamp.value = ''
    await fetchMessages(true)
  }
})

// Auto-scroll to bottom when new messages arrive
const chatContainer = ref<HTMLElement | null>(null)
watch(() => messages.value.length, () => {
  setTimeout(() => {
    if (chatContainer.value) {
      chatContainer.value.scrollTop = chatContainer.value.scrollHeight
    }
  }, 50)
})

onMounted(async () => {
  await fetchUserCharacters()
  await fetchMessages(true)
  setupPolling()

  // Mark all messages as read when component is mounted
  if (props.listId) {
    chatNotificationsStore.markAsRead(props.listId)
  }
})

onUnmounted(() => {
  if (pollingInterval.value !== null) {
    clearInterval(pollingInterval.value)
  }
})
</script>

<template>
  <div class="flex flex-col h-full">
    <div v-if="error" class="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <div class="flex items-center justify-between mb-2">
      <div class="flex items-center space-x-2">
        <span v-if="characters.length > 0" class="text-sm text-gray-600">
          {{ t('listDetail.chat.chatAs') }} <span class="font-medium">{{ characters[0]?.name }}</span>
        </span>
      </div>
      <button
        @click="toggleExpanded"
        class="text-gray-500 hover:text-gray-700"
        :title="isExpanded ? t('listDetail.chat.collapse') : t('listDetail.chat.expand')"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd"
            :d="isExpanded
              ? 'M5 10a1 1 0 0 1 1-1h8a1 1 0 1 1 0 2H6a1 1 0 0 1-1-1z'
              : 'M10 5a1 1 0 0 1 1 1v3h3a1 1 0 1 1 0 2h-3v3a1 1 0 1 1-2 0v-3H6a1 1 0 1 1 0-2h3V6a1 1 0 0 1 1-1z'"
            clip-rule="evenodd"
          />
        </svg>
      </button>
    </div>

    <div v-if="isExpanded">
      <!-- Messages container -->
      <div
        ref="chatContainer"
        class="flex-1 overflow-y-auto p-3 space-y-2 border border-gray-200 rounded-lg bg-gray-50 mb-3"
        :style="props.isCompact ? 'min-height: 150px; max-height: 300px;' : 'min-height: 300px; max-height: 500px;'"
      >
        <div v-if="loading && messages.length === 0" class="flex justify-center items-center h-full">
          <div class="animate-spin h-5 w-5 border-2 border-indigo-500 border-t-transparent rounded-full"></div>
          <span class="ml-2 text-gray-600">{{ t('listDetail.chat.loading') }}</span>
        </div>

        <div v-else-if="messages.length === 0" class="flex justify-center items-center h-full">
          <p class="text-gray-500">{{ t('listDetail.chat.noMessages') }}</p>
        </div>

        <div
          v-for="message in messages"
          :key="message.id"
          :class="[
            'p-2 rounded-lg max-w-3/4 break-words text-sm',
            isOwnMessage(message)
              ? 'bg-indigo-100 text-indigo-900 ml-auto'
              : 'bg-white border border-gray-200'
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
</template>
