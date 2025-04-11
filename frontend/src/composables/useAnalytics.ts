import { ref } from 'vue'
import { useUserStore } from '@/stores/user'

declare global {
  interface Window {
    gtag: (...args: unknown[]) => void
    dataLayer: unknown[]
  }
}

export function useAnalytics() {
  const userStore = useUserStore()
  const isEnabled = ref(false)

  // Initialize GA4
  const initGA4 = (measurementId: string) => {
    const script = document.createElement('script')
    script.async = true
    script.src = `https://www.googletagmanager.com/gtag/js?id=${measurementId}`
    document.head.appendChild(script)

    window.dataLayer = window.dataLayer || []
    window.gtag = function gtag(...args: unknown[]) {
      window.dataLayer.push(args)
    }
    window.gtag('js', new Date())
    window.gtag('config', measurementId, {
      send_page_view: false, // We'll handle page views manually
    })
  }

  // Track page view
  const trackPageView = (path: string) => {
    if (!isEnabled.value) return
    window.gtag('event', 'page_view', {
      page_path: path,
      user_id: userStore.userId,
    })
  }

  // Track custom event
  const trackEvent = (name: string, params?: Record<string, unknown>) => {
    if (!isEnabled.value) return
    window.gtag('event', name, {
      ...params,
      user_id: userStore.userId,
    })
  }

  // Enable/disable tracking based on cookie consent
  const setEnabled = (enabled: boolean) => {
    isEnabled.value = enabled
  }

  return {
    initGA4,
    trackPageView,
    trackEvent,
    setEnabled,
    isEnabled,
  }
}
