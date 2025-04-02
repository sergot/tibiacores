import { createI18n } from 'vue-i18n'
import en from '@/i18n/locales/en.json'
import pl from '@/i18n/locales/pl.json'
import de from '@/i18n/locales/de.json'
import es from '@/i18n/locales/es.json'
import pt from '@/i18n/locales/pt.json'

type Locale = 'en' | 'pl' | 'de' | 'es' | 'pt'
const SUPPORTED_LOCALES = ['en', 'pl', 'de', 'es', 'pt'] as const

export const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  messages: {
    en,
    pl,
    de,
    es,
    pt,
  },
})

function isValidLocale(locale: string): locale is Locale {
  return SUPPORTED_LOCALES.includes(locale as Locale)
}

export function getBrowserLocale(): Locale {
  const savedLocale = localStorage.getItem('user-locale')
  if (savedLocale && isValidLocale(savedLocale)) {
    return savedLocale
  }

  const navigatorLocale =
    navigator.languages !== undefined ? navigator.languages[0] : navigator.language

  if (!navigatorLocale) {
    return 'en'
  }

  const locale = navigatorLocale.trim().split(/-|_/)[0]
  return isValidLocale(locale) ? locale : 'en'
}

export function loadLocale(locale: Locale) {
  if (i18n.global.locale) {
    i18n.global.locale.value = locale
    localStorage.setItem('user-locale', locale)
    document.querySelector('html')?.setAttribute('lang', locale)
  }
}
