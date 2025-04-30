// src/i18n/index.ts
import { createI18n } from 'vue-i18n'
import en from '@/i18n/locales/en.json'
import ro from '@/i18n/locales/ro.json'

// Get locale from browser or stored preference
const getBrowserLocale = () => {
  const navigatorLocale = navigator.languages !== undefined
    ? navigator.languages[0]
    : navigator.language

  return navigatorLocale.split('-')[0]
}

const storedLocale = localStorage.getItem('userLocale')
const defaultLocale = storedLocale || getBrowserLocale() || 'en'

// Create i18n instance
const i18n = createI18n({
  legacy: false, // Use Composition API
  locale: defaultLocale,
  fallbackLocale: 'en',
  messages: {
    en,
    ro
  }
})

export default i18n

// Function to change locale
export const setLocale = (locale: "en" | "ro") => {
  i18n.global.locale.value = locale
  localStorage.setItem('userLocale', locale)
  document.querySelector('html')?.setAttribute('lang', locale)
}
