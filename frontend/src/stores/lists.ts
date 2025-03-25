import { defineStore } from 'pinia'
import { api } from '@/services/api'
import { useUserStore } from './user'
import type { AxiosError } from 'axios'
import type { ListDetails, APIErrorResponse } from '@/services/api'

interface List extends ListDetails {
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
        const lists = await api.lists.getAll(userStore.userId)
        // Add is_author flag based on whether the user ID matches the author ID
        this.lists = lists.map(list => ({
          ...list,
          is_author: list.author_id === userStore.userId
        }))
      } catch (err) {
        const axiosError = err as AxiosError<APIErrorResponse>
        console.error('Failed to fetch lists:', err)
        this.error = axiosError.response?.data?.message || 'Failed to fetch lists'
        this.lists = []
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
