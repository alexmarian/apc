import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'
import config from '@/config'
import type { LoginRequest, LoginResponse } from '@/types/api'
import { authApi } from '@/services/api'
import { onLogout, offLogout, clearAuthTokens, attemptTokenRefresh } from '@/services/auth-service'

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
  async function login(credentials: LoginRequest) {
    loading.value = true
    error.value = null

    try {
      const response = await authApi.login(credentials)
      const authData = response.data

      // Set auth data
      token.value = authData.token
      refreshToken.value = authData.refresh_token
      user.value = authData.login

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
      const success = await attemptTokenRefresh()
      
      if (success) {
        // Update token in store
        token.value = localStorage.getItem(config.authTokenKey)
        return true
      } else {
        // If refresh fails, logout
        logout()
        return false
      }
    } catch (err) {
      console.error('Token refresh failed:', err)
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
    clearAuthTokens()
    localStorage.removeItem('user')
  }

  // Internal logout handler for auth service
  function handleLogout() {
    token.value = null
    refreshToken.value = null
    user.value = null
  }

  // Register logout handler
  onLogout(handleLogout)

  // Cleanup on store dispose
  function dispose() {
    offLogout(handleLogout)
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
    logout,
    dispose
  }
})
