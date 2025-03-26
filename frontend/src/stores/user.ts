import { defineStore } from 'pinia'
import axios from 'axios'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('session_token') || '',
    userId: localStorage.getItem('user_id') || '',
    hasEmail: localStorage.getItem('has_email') === 'true',
  }),

  getters: {
    hasAccount: (state) => !!state.token,
    isAnonymous: (state) => !!state.token && !state.hasEmail,
    isAuthenticated: (state) => !!state.token,
  },

  actions: {
    setUser(data: { session_token: string; id: string; has_email: boolean }) {
      this.token = data.session_token
      this.userId = data.id
      this.hasEmail = data.has_email

      localStorage.setItem('session_token', data.session_token)
      localStorage.setItem('user_id', data.id)
      localStorage.setItem('has_email', String(data.has_email))
      
      axios.defaults.headers.common['Authorization'] = `Bearer ${data.session_token}`
    },

    clearUser() {
      this.token = ''
      this.userId = ''
      this.hasEmail = false

      localStorage.removeItem('session_token')
      localStorage.removeItem('user_id')
      localStorage.removeItem('has_email')
      
      delete axios.defaults.headers.common['Authorization']
    }
  }
})
