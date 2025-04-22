import { defineStore } from 'pinia'
import axios from 'axios'

interface UserData {
  id: string
  has_email: boolean
  expires_in?: number
}

export const useUserStore = defineStore('user', {
  state: () => ({
    userId: localStorage.getItem('user_id') || '',
    hasEmail: localStorage.getItem('has_email') === 'true',
    tokenExpiry: parseInt(localStorage.getItem('token_expiry') || '0'),
    refreshInProgress: false,
  }),

  getters: {
    hasAccount: (state) => !!state.userId,
    isAnonymous: (state) => !!state.userId && !state.hasEmail,
    isAuthenticated: (state) => !!state.userId,
    isTokenExpired: (state) => {
      if (!state.tokenExpiry) return true
      // Consider token expired 30 seconds before actual expiry
      // to prevent edge cases
      return Date.now() > state.tokenExpiry - 30000
    },
  },

  actions: {
    setUser(data: UserData) {
      this.userId = data.id
      this.hasEmail = data.has_email

      // Calculate token expiry time if expires_in is provided
      if (data.expires_in) {
        this.tokenExpiry = Date.now() + data.expires_in * 1000
        localStorage.setItem('token_expiry', String(this.tokenExpiry))
      }

      localStorage.setItem('user_id', data.id)
      localStorage.setItem('has_email', String(data.has_email))
    },

    clearUser() {
      this.userId = ''
      this.hasEmail = false
      this.tokenExpiry = 0

      localStorage.removeItem('token_expiry')
      localStorage.removeItem('user_id')
      localStorage.removeItem('has_email')
    },

    async refreshAccessToken() {
      // Prevent multiple refresh attempts
      if (this.refreshInProgress) {
        return
      }

      this.refreshInProgress = true

      try {
        const response = await axios.post('/auth/refresh')

        if (response.data.expires_in) {
          this.tokenExpiry = Date.now() + response.data.expires_in * 1000
          localStorage.setItem('token_expiry', String(this.tokenExpiry))
        }
      } catch (error) {
        console.error('Failed to refresh token:', error)
        // Clear user data on refresh token failure
        this.clearUser()
      } finally {
        this.refreshInProgress = false
      }
    },

    async logout() {
      try {
        await axios.post('/auth/logout')
      } catch (error) {
        console.error('Error during logout:', error)
      } finally {
        this.clearUser()
      }
    },
  },
})
