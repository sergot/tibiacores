import { defineStore } from 'pinia'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('session_token') || '',
    userId: localStorage.getItem('user_id') || '',
    isAuthenticated: !!localStorage.getItem('session_token'),
    isAnonymous: localStorage.getItem('is_anonymous') === 'true',
  }),
  actions: {
    setUser(data: { session_token: string; id: string; is_anonymous: boolean }) {
      // Only set user if there is no existing authenticated user or current user is anonymous
      if (!this.isAuthenticated || this.isAnonymous) {
        this.token = data.session_token
        this.userId = data.id
        this.isAnonymous = data.is_anonymous
        this.isAuthenticated = true

        localStorage.setItem('session_token', data.session_token)
        localStorage.setItem('user_id', data.id)
        localStorage.setItem('is_anonymous', String(data.is_anonymous))
      }
    },
    clearUser() {
      this.token = ''
      this.userId = ''
      this.isAnonymous = false
      this.isAuthenticated = false

      localStorage.removeItem('session_token')
      localStorage.removeItem('user_id')
      localStorage.removeItem('is_anonymous')
    },
  },
})
