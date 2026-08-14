import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'
import config from '@/config'
import type { LoginRequest } from '@/types/api'
import { authApi } from '@/services/api'
import {
  onLogout,
  offLogout,
  onTokenRefreshed,
  offTokenRefreshed,
  attemptTokenRefresh,
  performLogout,
  scheduleTokenRefresh,
  cancelTokenRefresh,
} from '@/services/auth-service'
import { isTokenExpired } from '@/utils/jwt'

// Create the auth store
export const useAuthStore = defineStore('auth', () => {
  // State
  const token = ref<string | null>(localStorage.getItem(config.authTokenKey))
  const refreshToken = ref<string | null>(localStorage.getItem(config.refreshTokenKey))
  const user = ref<string | null>(
    localStorage.getItem('user') ? JSON.parse(localStorage.getItem('user')!) : null,
  )
  const loading = ref(false)
  const error = ref<string | null>(null)

  /*
   * Checks the token's `exp`, not merely that a string is present. A JWT lives
   * one hour; treating any non-empty value as a session meant an expired token
   * still rendered the whole authenticated shell, with every request behind it
   * failing.
   *
   * This does not re-evaluate as time passes, and does not need to: the only
   * way the token reaches its expiry is with the scheduled refresh armed, and
   * that either replaces `token` or logs out. Both write to the ref, which is
   * what invalidates this.
   */
  const isAuthenticated = computed(() => !isTokenExpired(token.value))

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

      // Arm the renewal for this session; it re-arms itself thereafter.
      scheduleTokenRefresh()

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

  /*
   * Delegates to the auth service so a manual sign-out takes the same path as
   * an expiry: clear storage, cancel the refresh timer, notify listeners *and*
   * navigate. It used to only clear this store's refs, which left the user
   * looking at the authenticated shell with no session — every subsequent
   * click was then bounced back to /login by the guard, so the UI appeared
   * frozen rather than logged out. performLogout calls handleLogout below.
   */
  function logout() {
    performLogout()
  }

  // Internal logout handler for auth service
  function handleLogout() {
    token.value = null
    refreshToken.value = null
    user.value = null
  }

  // Register logout handler
  onLogout(handleLogout)

  // Follow tokens minted by a background refresh, which reach localStorage
  // directly; without this the store keeps serving the previous, expired one.
  function handleTokenRefreshed(newToken: string) {
    token.value = newToken
  }

  onTokenRefreshed(handleTokenRefreshed)

  // A reload lands here with a token already in localStorage, so the renewal
  // has to be armed on startup as well as after a login.
  if (isAuthenticated.value) {
    scheduleTokenRefresh()
  }

  // Cleanup on store dispose
  function dispose() {
    offLogout(handleLogout)
    offTokenRefreshed(handleTokenRefreshed)
    cancelTokenRefresh()
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
    dispose,
  }
})
