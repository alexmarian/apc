// Test utilities for auth functionality
import config from '@/config'
import { performLogout, attemptTokenRefresh, getAuthToken } from '@/services/auth-service'
import { getTokenExpiry } from '@/utils/jwt'

/**
 * Dev-only console helpers; main.ts imports this module solely when DEV.
 */
declare global {
  interface Window {
    authTests?: {
      testExpiredTokenLogout: () => void
      testFailedRefreshTokenLogout: () => Promise<void>
      logAuthState: () => void
    }
  }
}

// Test function to simulate expired token scenario
export function testExpiredTokenLogout() {
  console.log('Testing expired token logout scenario...')

  // Clear tokens to simulate expired state. These have to come from config:
  // the keys are prefixed ('apc_auth_token'), so the bare names this used
  // removed nothing and the helper never reproduced the state it claimed to.
  localStorage.removeItem(config.authTokenKey)
  localStorage.removeItem(config.refreshTokenKey)

  // This should trigger logout
  performLogout()

  console.log('Logout test completed')
}

// Test function to simulate failed refresh token scenario
export async function testFailedRefreshTokenLogout() {
  console.log('Testing failed refresh token logout scenario...')

  // Set invalid refresh token
  localStorage.setItem(config.refreshTokenKey, 'invalid_token')

  // This should fail and trigger logout
  const success = await attemptTokenRefresh()

  if (!success) {
    console.log('Refresh failed as expected, triggering logout...')
    performLogout()
  }

  console.log('Failed refresh test completed')
}

// Test function to check current auth state
export function logAuthState() {
  const expiresAt = getTokenExpiry(getAuthToken())

  console.log('Current auth state:')
  console.log('- Auth token:', getAuthToken() ? 'Present' : 'Missing')
  console.log(
    '- Expires:',
    expiresAt === null
      ? 'unknown (no readable exp)'
      : `${new Date(expiresAt).toLocaleTimeString()} (${Math.round((expiresAt - Date.now()) / 1000)}s)`,
  )
  console.log(
    '- Refresh token:',
    localStorage.getItem(config.refreshTokenKey) ? 'Present' : 'Missing',
  )
  console.log('- User:', localStorage.getItem('user') || 'Not set')
}

// Export test functions for console access
if (typeof window !== 'undefined') {
  window.authTests = {
    testExpiredTokenLogout,
    testFailedRefreshTokenLogout,
    logAuthState,
  }
}
