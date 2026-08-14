import { defineStore } from 'pinia'
import { ref } from 'vue'

const keys = {
  theme: 'theme',
  locale: 'locale'
}
// Create the auth store
export const usePreferences = defineStore('preferences', () => {
  // State
  const theme = ref<string | null>(localStorage.getItem(keys.theme))
  const locale = ref<string | null>(localStorage.getItem(keys.locale))

  function setTheme(newTheme: string) {
    theme.value = newTheme
    localStorage.setItem(keys.theme, newTheme)
  }

  function setLocale(newLocale: string) {
    locale.value = newLocale
    localStorage.setItem(keys.locale, newLocale)
  }

  // Return store
  return {
    // State
    theme,
    locale,

    // Actions
    setTheme,
    setLocale
  }
})
