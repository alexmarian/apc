import axios from 'axios'
import config from '@/config'
import { isAuthPath } from '@/router/authPages'
import { getTokenExpiry, isTokenExpired } from '@/utils/jwt'

/**
 * How far ahead of expiry to renew the token. Recovery used to begin only after
 * a request had already failed, so the first action after expiry always errored.
 */
const REFRESH_LEAD_MS = 60_000

// Global state for logout handling
let isLoggingOut = false
let logoutCallbacks: (() => void)[] = []
let tokenCallbacks: ((token: string) => void)[] = []

interface RouterNavigationService {
  push(to: { path: string; query?: Record<string, string> }): void
}

let routerService: RouterNavigationService | null = null

// Register router service for navigation
export function registerRouterService(router: RouterNavigationService) {
  routerService = router
}

// Register callback for logout events
export function onLogout(callback: () => void) {
  logoutCallbacks.push(callback)
}

// Remove logout callback
export function offLogout(callback: () => void) {
  logoutCallbacks = logoutCallbacks.filter((cb) => cb !== callback)
}

/**
 * Register a callback for tokens obtained by a background refresh. Without
 * this the new token reaches localStorage but not the Pinia store, leaving the
 * store holding the previous, expired string.
 */
export function onTokenRefreshed(callback: (token: string) => void) {
  tokenCallbacks.push(callback)
}

export function offTokenRefreshed(callback: (token: string) => void) {
  tokenCallbacks = tokenCallbacks.filter((cb) => cb !== callback)
}

function notifyLogout() {
  for (const callback of logoutCallbacks) {
    try {
      callback()
    } catch (error) {
      console.error('Error in logout callback:', error)
    }
  }
}

function notifyTokenRefreshed(token: string) {
  for (const callback of tokenCallbacks) {
    try {
      callback(token)
    } catch (error) {
      console.error('Error in token refresh callback:', error)
    }
  }
}

function navigateToLogin() {
  const path = typeof window !== 'undefined' ? window.location.pathname : '/'

  // Nothing worth returning to if the session died on an auth page itself.
  const query = isAuthPath(path)
    ? undefined
    : { redirect: path + (typeof window !== 'undefined' ? window.location.search : '') }

  if (routerService) {
    routerService.push({ path: '/login', query })
  } else if (typeof window !== 'undefined' && !isAuthPath(path)) {
    // Fallback for a logout raised before the router registered itself.
    window.location.href = '/login'
  }
}

// Perform logout with proper cleanup
export function performLogout() {
  if (isLoggingOut) {
    return // Prevent multiple simultaneous logouts
  }

  isLoggingOut = true

  try {
    cancelTokenRefresh()

    localStorage.removeItem(config.authTokenKey)
    localStorage.removeItem(config.refreshTokenKey)
    localStorage.removeItem('user')

    notifyLogout()
    navigateToLogin()
  } finally {
    isLoggingOut = false
  }
}

let refreshTimer: ReturnType<typeof setTimeout> | null = null

/**
 * Renew the token shortly before it expires. Re-armed after every successful
 * refresh, so one call after login covers the whole session.
 */
export function scheduleTokenRefresh(): void {
  cancelTokenRefresh()

  const expiresAt = getTokenExpiry(getAuthToken())
  if (expiresAt === null) {
    return
  }

  // Already inside the refresh window (or past it) clamps to 0 and runs now.
  const delay = Math.max(0, expiresAt - REFRESH_LEAD_MS - Date.now())
  refreshTimer = setTimeout(() => {
    void attemptTokenRefresh().then((success) => {
      if (!success) {
        performLogout()
      }
    })
  }, delay)
}

export function cancelTokenRefresh(): void {
  if (refreshTimer !== null) {
    clearTimeout(refreshTimer)
    refreshTimer = null
  }
}

let inFlightRefresh: Promise<boolean> | null = null

/**
 * Refresh the access token, collapsing concurrent callers onto one request.
 *
 * A page load can fire several requests that all 401 together; without the
 * shared promise each one starts its own refresh and they race to write the
 * token.
 */
export function attemptTokenRefresh(): Promise<boolean> {
  if (!inFlightRefresh) {
    inFlightRefresh = requestNewToken().finally(() => {
      inFlightRefresh = null
    })
  }
  return inFlightRefresh
}

async function requestNewToken(): Promise<boolean> {
  const refreshToken = localStorage.getItem(config.refreshTokenKey)

  if (!refreshToken) {
    return false
  }

  try {
    /*
     * apiBaseUrl carries a trailing slash, so interpolating '/refresh' produced
     * '/v1/api//refresh'. Go's mux 301-redirects that to the clean path, the
     * browser downgrades the redirected POST to GET, and it no longer matches
     * `POST /v1/api/refresh` -- refresh could never succeed. Everything else
     * goes through the axios instance, whose baseURL join normalises this.
     */
    const url = `${config.apiBaseUrl.replace(/\/+$/, '')}/refresh`

    const response = await axios.post<{ token: string }>(url, null, {
      headers: {
        Authorization: `Bearer ${refreshToken}`,
        'Content-Type': 'application/json',
      },
      timeout: config.apiTimeout,
    })

    const newToken = response.data.token
    localStorage.setItem(config.authTokenKey, newToken)

    notifyTokenRefreshed(newToken)
    scheduleTokenRefresh()

    return true
  } catch (error) {
    console.error('Token refresh failed:', error)
    return false
  }
}

/**
 * True only for a token that is present *and* still inside its lifetime. No
 * skew here on purpose: a token in its last minute is still accepted by the
 * API, and the scheduled refresh is what renews it. Applying the lead time
 * would sign the user out a minute early instead.
 */
export function isAuthenticated(): boolean {
  return !isTokenExpired(getAuthToken())
}

// Get current auth token
export function getAuthToken(): string | null {
  return localStorage.getItem(config.authTokenKey)
}

// Get current refresh token
export function getRefreshToken(): string | null {
  return localStorage.getItem(config.refreshTokenKey)
}

// Clear auth tokens
export function clearAuthTokens(): void {
  cancelTokenRefresh()
  localStorage.removeItem(config.authTokenKey)
  localStorage.removeItem(config.refreshTokenKey)
}
