import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'
import config from '@/config'
import type { LoginRequest, LoginResponse } from '@/types/api'
import { authApi } from '@/services/api'

const keys = {
  theme: 'theme'
}
// Create the auth store
export const usePreferences = defineStore('preferences', () => {
  // State
  const theme = ref<string | null>(localStorage.getItem(keys.theme))

  function setTheme(newTheme: string) {
    theme.value = newTheme
    localStorage.setItem(keys.theme, newTheme)
  }

  // Return store
  return {
    // State
    theme,


    // Actions
    setTheme
  }
})
