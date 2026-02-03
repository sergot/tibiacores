import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import axios from 'axios'

import App from './App.vue'
import router from './router'
import { i18n, getBrowserLocale, loadLocale } from './i18n'

// Define Umami type for TypeScript
declare global {
  interface Window {
    umami?: {
      track: (event_name: string, event_data?: Record<string, unknown>) => void
    }
  }
}

// Setup Umami analytics
const UMAMI_SCRIPT_ID = 'umami-analytics-script'

// Function to load Umami analytics
export const loadUmamiAnalytics = () => {
  const umamiWebsiteId = import.meta.env.VITE_UMAMI_WEBSITE_ID
  if (!umamiWebsiteId) return

  // Skip if already loaded
  if (document.getElementById(UMAMI_SCRIPT_ID)) return

  // Initialize Umami
  const script = document.createElement('script')
  script.id = UMAMI_SCRIPT_ID
  script.async = true
  script.defer = true
  script.setAttribute('data-website-id', umamiWebsiteId)
  script.src = 'https://umami.tibiacores.com/script.js'
  document.head.appendChild(script)
}

// Helper function to track custom events (if needed elsewhere in the app)
export const trackEvent = (name: string, params?: Record<string, unknown>) => {
  if (window.umami) {
    window.umami.track(name, params)
  }
}

// Check consent on initial load
const consentData = localStorage.getItem('cookie_consent')
if (consentData) {
  const consent = JSON.parse(consentData)
  if (consent.analytics) {
    loadUmamiAnalytics()
  }
}

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(router)
app.use(i18n)

// Initialize locale
const locale = getBrowserLocale()
loadLocale(locale)

// Configure axios base URL
axios.defaults.baseURL = import.meta.env.VITE_API_URL || '/api'
// Enable cookies for OAuth state validation (Double Submit Cookie pattern)
axios.defaults.withCredentials = true

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
      const currentRoute = router.currentRoute.value

      // Check if this is likely an authorization error (accessing specific resource)
      // vs authentication error (invalid/expired token)
      const isResourceAccessError = currentRoute.matched.some(record => record.meta.requiresAuth) &&
                                   userStore.isAuthenticated

      if (isResourceAccessError) {
        // For resource access errors, just redirect to home without clearing user data
        router.replace('/')
      } else {
        // For authentication errors, clear user data and redirect
        userStore.clearUser()
        listsStore.clearLists()

        // Only redirect if not already on signin/signup/home pages
        const authRoutes = ['/signin', '/signup', '/']
        if (!authRoutes.includes(currentRoute.path)) {
          router.replace('/')
        }
      }
    }
    return Promise.reject(error)
  },
)

app.mount('#app')
