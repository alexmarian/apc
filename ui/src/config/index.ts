// Environment configuration
// In a real app, this would be injected via environment variables
// and processed by Vite's import.meta.env feature

interface AppConfig {
  // API Configuration
  apiBaseUrl: string
  apiTimeout: number

  // Authentication
  authTokenKey: string
  refreshTokenKey: string

  // App info
  appName: string
  appVersion: string
}

// Development configuration
const devConfig: AppConfig = {
  apiBaseUrl: 'http://localhost:8080/v1/api',
  apiTimeout: 10000,
  authTokenKey: 'apc_auth_token',
  refreshTokenKey: 'apc_refresh_token',
  appName: 'APC Management Portal',
  appVersion: '1.0.0-dev'
}

// Production configuration
const prodConfig: AppConfig = {
  apiBaseUrl: import.meta.env.MODE === 'production'
    ? (import.meta.env.VITE_API_BASE_URL || 'https://api.example.com/v1/api')
    : devConfig.apiBaseUrl,
  apiTimeout: 30000,
  authTokenKey: 'apc_auth_token',
  refreshTokenKey: 'apc_refresh_token',
  appName: 'APC Management Portal',
  appVersion: '1.0.0'
}

// Export the appropriate configuration based on environment
const config: AppConfig = import.meta.env.MODE === 'production' ? prodConfig : devConfig

export default config
