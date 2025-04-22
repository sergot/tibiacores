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

// Enable sending cookies with cross-origin requests (required for HttpOnly cookie auth)
axios.defaults.withCredentials = true

// Configure axios after Pinia is initialized
import { useUserStore } from './stores/user'

// Request interceptor
axios.interceptors.request.use(async (config) => {
  const userStore = useUserStore()

  // Skip token refresh for token refresh requests to avoid infinite loops
  if (config.url === '/auth/refresh') {
    return config
  }

  // Check if user is logged in and token is expired
  if (userStore.isAuthenticated && userStore.isTokenExpired) {
    await userStore.refreshAccessToken()
  }

  return config
})

// Response interceptor to handle 401 Unauthorized errors
axios.interceptors.response.use(
  (response) => response,
  async (error) => {
    const userStore = useUserStore()
    const originalRequest = error.config

    // If the request failed due to 401 Unauthorized and we haven't tried refreshing yet
    if (error.response?.status === 401 && !originalRequest._retry) {
      // Mark this request as retried
      originalRequest._retry = true

      try {
        // Try to refresh the token
        await userStore.refreshAccessToken()

        // Retry the original request - cookies will be sent automatically
        return axios(originalRequest)
      } catch (refreshError) {
        // If refresh fails, redirect to login
        userStore.clearUser()
        window.location.href = '/signin'
        return Promise.reject(refreshError)
      }
    }

    return Promise.reject(error)
  },
)

app.mount('#app')
