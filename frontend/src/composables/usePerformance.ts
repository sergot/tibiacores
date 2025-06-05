// Core Web Vitals and Performance Optimization for TibiaCores
// This composable provides utilities to monitor and improve Core Web Vitals

import { ref, onMounted } from 'vue'

interface WebVitalsMetrics {
  fcp?: number // First Contentful Paint
  lcp?: number // Largest Contentful Paint
  fid?: number // First Input Delay
  cls?: number // Cumulative Layout Shift
  ttfb?: number // Time to First Byte
}

export function useWebVitals() {
  const metrics = ref<WebVitalsMetrics>({})
  const isSupported = ref(false)

  const measureFCP = () => {
    try {
      const observer = new PerformanceObserver((list) => {
        const entries = list.getEntries()
        const fcpEntry = entries.find(entry => entry.name === 'first-contentful-paint')
        if (fcpEntry) {
          metrics.value.fcp = fcpEntry.startTime
          observer.disconnect()
        }
      })
      observer.observe({ entryTypes: ['paint'] })
    } catch (error) {
      console.warn('FCP measurement not supported:', error)
    }
  }

  const measureLCP = () => {
    try {
      const observer = new PerformanceObserver((list) => {
        const entries = list.getEntries()
        const lastEntry = entries[entries.length - 1]
        metrics.value.lcp = lastEntry.startTime
      })
      observer.observe({ entryTypes: ['largest-contentful-paint'] })
    } catch (error) {
      console.warn('LCP measurement not supported:', error)
    }
  }

  const measureCLS = () => {
    try {
      let clsValue = 0
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          const layoutShiftEntry = entry as PerformanceEntry & {
            hadRecentInput?: boolean
            value: number
          }
          if (!layoutShiftEntry.hadRecentInput) {
            clsValue += layoutShiftEntry.value
            metrics.value.cls = clsValue
          }
        }
      })
      observer.observe({ entryTypes: ['layout-shift'] })
    } catch (error) {
      console.warn('CLS measurement not supported:', error)
    }
  }

  const measureTTFB = () => {
    try {
      const navigationEntry = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming
      if (navigationEntry) {
        metrics.value.ttfb = navigationEntry.responseStart - navigationEntry.requestStart
      }
    } catch (error) {
      console.warn('TTFB measurement not supported:', error)
    }
  }

  const initializeMetrics = () => {
    if (typeof window !== 'undefined' && 'PerformanceObserver' in window) {
      isSupported.value = true
      measureFCP()
      measureLCP()
      measureCLS()
      measureTTFB()
    }
  }

  const getMetricsReport = () => {
    const report = {
      ...metrics.value,
      scores: {
        fcp: metrics.value.fcp ? (metrics.value.fcp < 1800 ? 'good' : metrics.value.fcp < 3000 ? 'needs-improvement' : 'poor') : 'unknown',
        lcp: metrics.value.lcp ? (metrics.value.lcp < 2500 ? 'good' : metrics.value.lcp < 4000 ? 'needs-improvement' : 'poor') : 'unknown',
        cls: metrics.value.cls ? (metrics.value.cls < 0.1 ? 'good' : metrics.value.cls < 0.25 ? 'needs-improvement' : 'poor') : 'unknown',
        ttfb: metrics.value.ttfb ? (metrics.value.ttfb < 800 ? 'good' : metrics.value.ttfb < 1800 ? 'needs-improvement' : 'poor') : 'unknown',
      }
    }
    return report
  }

  onMounted(() => {
    initializeMetrics()
  })

  return {
    metrics,
    isSupported,
    getMetricsReport
  }
}

// Performance optimization utilities
export const performanceUtils = {
  // Preload critical resources
  preloadResource(href: string, as: string, type?: string) {
    const link = document.createElement('link')
    link.rel = 'preload'
    link.href = href
    link.as = as
    if (type) link.type = type
    document.head.appendChild(link)
  },

  // Prefetch next page resources
  prefetchResource(href: string) {
    const link = document.createElement('link')
    link.rel = 'prefetch'
    link.href = href
    document.head.appendChild(link)
  },

  // Lazy load images with intersection observer
  lazyLoadImage(img: HTMLImageElement, src: string) {
    if ('IntersectionObserver' in window) {
      const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
          if (entry.isIntersecting) {
            img.src = src
            img.classList.remove('lazy')
            observer.unobserve(img)
          }
        })
      })
      observer.observe(img)
    } else {
      // Fallback for browsers without IntersectionObserver
      img.src = src
    }
  },

  // Debounce function for search inputs
  debounce<T extends (...args: unknown[]) => unknown>(func: T, wait: number): (...args: Parameters<T>) => void {
    let timeout: number
    return (...args: Parameters<T>) => {
      clearTimeout(timeout)
      timeout = setTimeout(() => func.apply(this, args), wait)
    }
  },

  // Throttle function for scroll events
  throttle<T extends (...args: unknown[]) => unknown>(func: T, limit: number): (...args: Parameters<T>) => void {
    let inThrottle: boolean
    return (...args: Parameters<T>) => {
      if (!inThrottle) {
        func.apply(this, args)
        inThrottle = true
        setTimeout(() => inThrottle = false, limit)
      }
    }
  },

  // Check if user prefers reduced motion
  prefersReducedMotion(): boolean {
    return window.matchMedia('(prefers-reduced-motion: reduce)').matches
  },

  // Get connection quality information
  getConnectionInfo() {
    const connection = (navigator as NavigatorWithConnection).connection || 
                      (navigator as NavigatorWithConnection).mozConnection || 
                      (navigator as NavigatorWithConnection).webkitConnection
    if (connection) {
      return {
        effectiveType: connection.effectiveType,
        downlink: connection.downlink,
        rtt: connection.rtt,
        saveData: connection.saveData
      }
    }
    return null
  }
}

// Type for navigator with connection API
interface NavigatorWithConnection extends Navigator {
  connection?: {
    effectiveType: string
    downlink: number
    rtt: number
    saveData: boolean
  }
  mozConnection?: {
    effectiveType: string
    downlink: number
    rtt: number
    saveData: boolean
  }
  webkitConnection?: {
    effectiveType: string
    downlink: number
    rtt: number
    saveData: boolean
  }
}

// Image optimization utilities
export const imageUtils = {
  // Generate responsive image srcset
  generateSrcSet(baseUrl: string, sizes: number[]): string {
    return sizes.map(size => `${baseUrl}?w=${size} ${size}w`).join(', ')
  },

  // Get optimal image format based on browser support
  getOptimalFormat(): 'webp' | 'avif' | 'jpg' {
    if (this.supportsAvif()) return 'avif'
    if (this.supportsWebp()) return 'webp'
    return 'jpg'
  },

  supportsWebp(): boolean {
    const canvas = document.createElement('canvas')
    canvas.width = 1
    canvas.height = 1
    return canvas.toDataURL('image/webp').indexOf('data:image/webp') === 0
  },

  supportsAvif(): boolean {
    const canvas = document.createElement('canvas')
    canvas.width = 1
    canvas.height = 1
    return canvas.toDataURL('image/avif').indexOf('data:image/avif') === 0
  }
}
