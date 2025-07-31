// Test utilities for auth functionality
import { performLogout, attemptTokenRefresh, getAuthToken } from '@/services/auth-service'

// Test function to simulate expired token scenario
export function testExpiredTokenLogout() {
  console.log('Testing expired token logout scenario...')
  
  // Clear tokens to simulate expired state
  localStorage.removeItem('auth_token')
  localStorage.removeItem('refresh_token')
  
  // This should trigger logout
  performLogout()
  
  console.log('Logout test completed')
}

// Test function to simulate failed refresh token scenario
export async function testFailedRefreshTokenLogout() {
  console.log('Testing failed refresh token logout scenario...')
  
  // Set invalid refresh token
  localStorage.setItem('refresh_token', 'invalid_token')
  
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
  console.log('Current auth state:')
  console.log('- Auth token:', getAuthToken() ? 'Present' : 'Missing')
  console.log('- Refresh token:', localStorage.getItem('refresh_token') ? 'Present' : 'Missing')
  console.log('- User:', localStorage.getItem('user') || 'Not set')
}

// Export test functions for console access
if (typeof window !== 'undefined') {
  (window as any).authTests = {
    testExpiredTokenLogout,
    testFailedRefreshTokenLogout,
    logAuthState
  }
}