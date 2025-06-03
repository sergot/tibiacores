import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import axios from 'axios'
import { useUserStore } from './user'

interface ChatNotification {
  list_id: string
  list_name: string
  last_message_time: string
  unread_count: number
  last_character_name: string
}

export const useChatNotificationsStore = defineStore('chatNotifications', () => {
  const userStore = useUserStore()
  const notifications = ref<ChatNotification[]>([])
  const loading = ref(false)
  const error = ref('')

  const totalUnreadMessages = computed(() => {
    return notifications.value.reduce((total, notification) => total + notification.unread_count, 0)
  })

  const hasUnreadMessages = computed(() => totalUnreadMessages.value > 0)

  const fetchChatNotifications = async () => {
    if (!userStore.isAuthenticated) return

    loading.value = true
    error.value = ''

    try {
      const response = await axios.get('/chat-notifications')
      notifications.value = response.data
    } catch (err) {
      console.error('Failed to fetch chat notifications:', err)
      error.value = 'Failed to load chat notifications'
      notifications.value = []
    } finally {
      loading.value = false
    }
  }

  // Mark a specific list's messages as read
  const markAsRead = async (listId: string) => {
    if (!userStore.isAuthenticated) return

    try {
      await axios.post(`/lists/${listId}/chat/read`)
      // Update the local state
      notifications.value = notifications.value.filter(n => n.list_id !== listId)
    } catch (err) {
      console.error('Failed to mark messages as read:', err)
    }
  }

  // Start polling when store is initialized
  let pollInterval: number | undefined

  const startPolling = () => {
    if (pollInterval) return
    fetchChatNotifications()
    pollInterval = window.setInterval(fetchChatNotifications, 30000) // Check every 30 seconds
  }

  const stopPolling = () => {
    if (pollInterval) {
      window.clearInterval(pollInterval)
      pollInterval = undefined
    }
  }

  // Clean up on store destruction
  const $dispose = () => {
    stopPolling()
  }

  return {
    notifications,
    loading,
    error,
    hasUnreadMessages,
    totalUnreadMessages,
    fetchChatNotifications,
    markAsRead,
    startPolling,
    stopPolling,
    $dispose,
  }
})
