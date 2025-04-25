import axios, { AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios'
import {
  Account,
  AccountCreateRequest,
  AccountUpdateRequest,
  ApiResponse
} from '@/types/api'
import config from '@/config'

// Create axios instance
const api = axios.create({
  baseURL: config.apiBaseUrl,
  timeout: config.apiTimeout,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor
api.interceptors.request.use(
  (reqConfig: AxiosRequestConfig) => {
    const token = localStorage.getItem(config.authTokenKey)
    if (token && reqConfig.headers) {
      reqConfig.headers.Authorization = `Bearer ${token}`
    }
    return reqConfig
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  (response: AxiosResponse) => {
    return response
  },
  (error: AxiosError) => {
    // Handle common error cases here
    if (error.response?.status === 401) {
      // Unauthorized - token expired or invalid
      localStorage.removeItem(config.authTokenKey)
      localStorage.removeItem(config.refreshTokenKey)
      // Redirect to login page
      window.location.href = '/login'
    }

    return Promise.reject(error)
  }
)

// Account APIs
export const accountApi = {
  // Get all accounts for an association
  getAccounts: (associationId: number) =>
    api.get<Account[]>(`/associations/${associationId}/accounts`),

  // Get a specific account
  getAccount: (associationId: number, accountId: number) =>
    api.get<Account>(`/associations/${associationId}/accounts/${accountId}`),

  // Create a new account
  createAccount: (associationId: number, accountData: AccountCreateRequest) =>
    api.post<Account>(`/associations/${associationId}/accounts`, accountData),

  // Update an existing account
  updateAccount: (associationId: number, accountId: number, accountData: AccountUpdateRequest) =>
    api.put<Account>(`/associations/${associationId}/accounts/${accountId}`, accountData),

  // Disable an account
  disableAccount: (associationId: number, accountId: number) =>
    api.put<ApiResponse<null>>(`/associations/${associationId}/accounts/${accountId}/disable`)
}

// Export the api instance to allow direct access if needed
export default api
