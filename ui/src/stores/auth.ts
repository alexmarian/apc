import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'
import config from '@/config'

interface LoginCredentials {
  login: string
  password: string
  totp: string
}

interface AuthPayload {
  login: string
  token: string
  refresh_token: string
}

// Create the auth store
export const useAuthStore = defineStore('auth', () => {
  // State
  const token = ref<string | null>(localStorage.getItem(config.authTokenKey))
  const refreshToken = ref<string | null>(localStorage.getItem(config.refreshTokenKey))
  const user = ref<string | null>(localStorage.getItem('user') ? JSON.parse(localStorage.getItem('user')!) : null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Computed
  const isAuthenticated = computed(() => !!token.value)

  // Actions
  async function login(credentials: LoginCredentials) {
    loading.value = true
    error.value = null

    try {
      const response = await axios.post<AuthPayload>(`${config.apiBaseUrl}/login`, credentials)

      // Set auth data
      token.value = response.data.token
      refreshToken.value = response.data.refresh_token
      user.value = response.data.login

      // Save to localStorage
      localStorage.setItem(config.authTokenKey, token.value)
      localStorage.setItem(config.refreshTokenKey, refreshToken.value)
      localStorage.setItem('user', JSON.stringify(user.value))

      return true
    } catch (err) {
      if (axios.isAxiosError(err)) {
        error.value = err.response?.data?.msg || 'Login failed'
      } else {
        error.value = 'Login failed. Please try again.'
      }
      return false
    } finally {
      loading.value = false
    }
  }

  async function refreshAccessToken() {
    if (!refreshToken.value) {
      return false
    }

    try {
      const response = await axios.post<{ token: string }>(`${config.apiBaseUrl}/refresh`, null, {
        headers: {
          Authorization: `Bearer ${refreshToken.value}`
        }
      })

      // Update token
      token.value = response.data.token
      localStorage.setItem(config.authTokenKey, token.value)

      return true
    } catch (err) {
      // If refresh fails, logout
      logout()
      return false
    }
  }

  function logout() {
    // Clear auth data
    token.value = null
    refreshToken.value = null
    user.value = null

    // Clear localStorage
    localStorage.removeItem(config.authTokenKey)
    localStorage.removeItem(config.refreshTokenKey)
    localStorage.removeItem('user')
  }

  // Return store
  return {
    // State
    token,
    refreshToken,
    user,
    loading,
    error,

    // Computed
    isAuthenticated,

    // Actions
    login,
    refreshAccessToken,
    logout
  }
})
