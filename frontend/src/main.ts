import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import axios from 'axios'

import App from './App.vue'
import router from './router'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(router)

// Configure axios after Pinia is initialized
import { useUserStore } from './stores/user'
import { useListsStore } from './stores/lists'

// Request interceptor
axios.interceptors.request.use((config) => {
  const userStore = useUserStore()
  if (userStore.token) {
    config.headers.Authorization = `Bearer ${userStore.token}`
  }
  return config
})

// Response interceptor
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const userStore = useUserStore()
      const listsStore = useListsStore()

      // Clear both user data and lists
      userStore.clearUser()
      listsStore.clearLists()

      // Only redirect if not already on signin/signup pages
      const authRoutes = ['/signin', '/signup']
      if (!authRoutes.includes(router.currentRoute.value.path)) {
        router.push('/signin')
      }
    }
    return Promise.reject(error)
  }
)

app.mount('#app')
