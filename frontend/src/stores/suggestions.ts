import { defineStore } from 'pinia'
import { api } from '@/services/api'

interface PendingSuggestion {
  character_id: string
  character_name: string
  suggestion_count: number
}

interface SuggestionsState {
  pendingSuggestions: PendingSuggestion[]
  error: string | null
  pollingInterval: ReturnType<typeof setInterval> | null
}

export const useSuggestionsStore = defineStore('suggestions', {
  state: (): SuggestionsState => ({
    pendingSuggestions: [],
    error: null,
    pollingInterval: null,
  }),

  getters: {
    hasPendingSuggestions: (state) => state.pendingSuggestions.length > 0,
    totalPendingSuggestions: (state) =>
      state.pendingSuggestions.reduce((total, char) => total + char.suggestion_count, 0),
  },

  actions: {
    async fetchPendingSuggestions() {
      try {
        this.pendingSuggestions = await api.suggestions.getPending()
      } catch (err) {
        console.error('Failed to fetch pending suggestions:', err)
        this.error = 'Failed to fetch pending suggestions'
      }
    },

    startPolling() {
      // Start polling every 30 seconds
      this.pollingInterval = setInterval(() => {
        this.fetchPendingSuggestions()
      }, 30000)
      // Initial fetch
      this.fetchPendingSuggestions()
    },

    stopPolling() {
      if (this.pollingInterval) {
        clearInterval(this.pollingInterval)
        this.pollingInterval = null
      }
    },
  },
})