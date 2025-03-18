import { ref } from 'vue'
import { defineStore } from 'pinia'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('session_token') || '')
  const isAuthenticated = ref(!!token.value)

  function setToken(newToken: string) {
    token.value = newToken
    localStorage.setItem('session_token', newToken)
    isAuthenticated.value = true
  }

  function clearToken() {
    token.value = ''
    localStorage.removeItem('session_token')
    isAuthenticated.value = false
  }

  return {
    token,
    isAuthenticated,
    setToken,
    clearToken
  }
})
