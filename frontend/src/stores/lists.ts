import { defineStore } from 'pinia'
import axios from 'axios'
import { useUserStore } from './user'

export interface List {
  id: string
  name: string
  world: string
  share_code: string
  created_at: string
  updated_at: string
  is_author: boolean
  character_name?: string
}

interface ListsState {
  lists: List[]
  isLoading: boolean
  error: string | null
}

export const useListsStore = defineStore('lists', {
  state: (): ListsState => ({
    lists: [],
    isLoading: false,
    error: null,
  }),

  getters: {
    hasLists: (state) => state.lists.length > 0,
    getListById: (state) => {
      return (id: string) => state.lists.find((list) => list.id === id)
    },
  },

  actions: {
    async fetchUserLists() {
      const userStore = useUserStore()
      if (!userStore.userId) {
        this.error = 'No user ID available'
        return
      }

      this.isLoading = true
      this.error = null

      try {
        const response = await axios.get<List[]>(`/api/lists/${userStore.userId}`)
        this.lists = response.data
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to fetch lists'
        console.error('Failed to fetch lists:', err)
      } finally {
        this.isLoading = false
      }
    },

    clearLists() {
      this.lists = []
      this.error = null
    },
  },
})
