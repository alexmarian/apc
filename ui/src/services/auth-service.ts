import axios from 'axios'
import config from '@/config'

// Global state for logout handling
let isLoggingOut = false
let logoutCallbacks: (() => void)[] = []

// Interface for router navigation
interface RouterNavigationService {
  push(path: string): void
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
  logoutCallbacks = logoutCallbacks.filter(cb => cb !== callback)
}

// Perform logout with proper cleanup
export function performLogout() {
  if (isLoggingOut) {
    return // Prevent multiple simultaneous logouts
  }

  isLoggingOut = true

  try {
    // Clear all auth data
    localStorage.removeItem(config.authTokenKey)
    localStorage.removeItem(config.refreshTokenKey)
    localStorage.removeItem('user')

    // Notify all listeners
    logoutCallbacks.forEach(callback => {
      try {
        callback()
      } catch (error) {
        console.error('Error in logout callback:', error)
      }
    })

    // Navigate to login page
    if (routerService) {
      routerService.push('/login')
    } else if (typeof window !== 'undefined') {
      // Fallback to window.location if router is not available
      const currentPath = window.location.pathname
      if (currentPath !== '/login') {
        window.location.href = '/login'
      }
    }
  } finally {
    isLoggingOut = false
  }
}

// Attempt to refresh token without circular dependencies
export async function attemptTokenRefresh(): Promise<boolean> {
  const refreshToken = localStorage.getItem(config.refreshTokenKey)
  
  if (!refreshToken) {
    return false
  }

  try {
    const response = await axios.post<{ token: string }>(`${config.apiBaseUrl}/refresh`, null, {
      headers: {
        Authorization: `Bearer ${refreshToken}`,
        'Content-Type': 'application/json'
      },
      timeout: config.apiTimeout
    })

    const newToken = response.data.token
    localStorage.setItem(config.authTokenKey, newToken)
    
    return true
  } catch (error) {
    console.error('Token refresh failed:', error)
    return false
  }
}

// Check if user is authenticated
export function isAuthenticated(): boolean {
  const token = localStorage.getItem(config.authTokenKey)
  return !!token
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
  localStorage.removeItem(config.authTokenKey)
  localStorage.removeItem(config.refreshTokenKey)
}