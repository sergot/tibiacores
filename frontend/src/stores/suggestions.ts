import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import axios from 'axios'
import { useUserStore } from './user'

interface PendingSuggestion {
  character_id: string
  character_name: string
  suggestion_count: number
}

export const useSuggestionsStore = defineStore('suggestions', () => {
  const userStore = useUserStore()
  const pendingSuggestions = ref<PendingSuggestion[]>([])
  const loading = ref(false)
  const error = ref('')

  const totalPendingSuggestions = computed(() => {
    return pendingSuggestions.value.reduce((total, char) => total + char.suggestion_count, 0)
  })

  const hasPendingSuggestions = computed(() => totalPendingSuggestions.value > 0)

  const fetchPendingSuggestions = async () => {
    if (!userStore.isAuthenticated) return

    loading.value = true
    error.value = ''

    try {
      const response = await axios.get('/api/pending-suggestions')
      pendingSuggestions.value = response.data
    } catch (err) {
      console.error('Failed to fetch pending suggestions:', err)
      error.value = 'Failed to load suggestions'
      pendingSuggestions.value = []
    } finally {
      loading.value = false
    }
  }

  // Start polling when store is initialized
  let pollInterval: number | undefined
  
  const startPolling = () => {
    if (pollInterval) return
    fetchPendingSuggestions()
    pollInterval = window.setInterval(fetchPendingSuggestions, 60000) // Check every minute
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
    pendingSuggestions,
    loading,
    error,
    hasPendingSuggestions,
    totalPendingSuggestions,
    fetchPendingSuggestions,
    startPolling,
    stopPolling,
    $dispose
  }
})