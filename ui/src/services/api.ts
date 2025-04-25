import axios from 'axios'
import type { AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios'
import type {
  Account,
  AccountCreateRequest,
  AccountUpdateRequest,
  ApiResponse,
  Association,
  Building,
  Category,
  CategoryCreateRequest,
  Expense,
  ExpenseCreateRequest,
  LoginRequest,
  LoginResponse,
  Owner,
  Unit
} from '@/types/api'
import config from '@/config'
import { useAuthStore } from '@/stores/auth'

import '@/types/axios'

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
    if (reqConfig.url &&
      (reqConfig.url.includes('/login') ||
        reqConfig.url.includes('/refresh'))) {
      return reqConfig
    }
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
  async (error: AxiosError) => {
    const originalRequest = error.config

    // If error is 401 and we haven't tried refreshing the token yet
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true // Mark that we've tried refreshing for this request

      try {
        const refreshToken = localStorage.getItem(config.refreshTokenKey)

        // Only try refreshing if we have a refresh token
        if (refreshToken) {
          // Get auth store outside of Vue component
          const authStore = useAuthStore()

          // Try to refresh the token
          const success = await authStore.refreshAccessToken()

          if (success) {
            // Update the Authorization header with new token
            const newToken = localStorage.getItem(config.authTokenKey)
            if (originalRequest.headers) {
              originalRequest.headers.Authorization = `Bearer ${newToken}`
            }

            // Retry the original request with new token
            return api(originalRequest)
          }
        }

        // If we couldn't refresh token or don't have refresh token,
        // proceed with logout and redirect
        localStorage.removeItem(config.authTokenKey)
        localStorage.removeItem(config.refreshTokenKey)

        // Redirect to login page
        window.location.href = '/login'
      } catch (refreshError) {
        console.error('Token refresh failed:', refreshError)

        // Clear auth data and redirect
        localStorage.removeItem(config.authTokenKey)
        localStorage.removeItem(config.refreshTokenKey)
        window.location.href = '/login'
      }
    }

    return Promise.reject(error)
  }
)

// Auth APIs
export const authApi = {
  login: (credentials: LoginRequest) =>
    api.post<LoginResponse>(`/login`, credentials),

  refreshToken: (refreshToken: string) =>
    api.post<{ token: string }>(`/refresh`, null, {
      headers: {
        Authorization: `Bearer ${refreshToken}`
      }
    })
}

// Association APIs
export const associationApi = {
  getAssociations: () =>
    api.get<Association[]>(`/associations`),

  getAssociation: (associationId: number) =>
    api.get<Association>(`/associations/${associationId}`)
}

// Building APIs
export const buildingApi = {
  getBuildings: (associationId: number) =>
    api.get<Building[]>(`/associations/${associationId}/buildings`),

  getBuilding: (associationId: number, buildingId: number) =>
    api.get<Building>(`/associations/${associationId}/buildings/${buildingId}`)
}

// Unit APIs
export const unitApi = {
  getUnits: (associationId: number, buildingId: number) =>
    api.get<Unit[]>(`/associations/${associationId}/buildings/${buildingId}/units`),

  getUnit: (associationId: number, buildingId: number, unitId: number) =>
    api.get<Unit>(`/associations/${associationId}/buildings/${buildingId}/units/${unitId}`)
}

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

// Owner APIs
export const ownerApi = {
  getOwners: (associationId: number) =>
    api.get<Owner[]>(`/associations/${associationId}/owners`),

  getOwnerReport: (associationId: number, includeUnits: boolean = false, includeCoOwners: boolean = false) =>
    api.get(`/associations/${associationId}/owners/report`, {
      params: {
        units: includeUnits ? 'true' : 'false',
        co_owners: includeCoOwners ? 'true' : 'false'
      }
    })
}

// Category APIs
export const categoryApi = {
  getCategories: (associationId: number) =>
    api.get<Category[]>(`/associations/${associationId}/categories`),

  getCategory: (associationId: number, categoryId: number) =>
    api.get<Category>(`/associations/${associationId}/categories/${categoryId}`),

  createCategory: (associationId: number, categoryData: CategoryCreateRequest) =>
    api.post<Category>(`/associations/${associationId}/categories`, categoryData),

  deactivateCategory: (associationId: number, categoryId: number) =>
    api.put<ApiResponse<null>>(`/associations/${associationId}/categories/${categoryId}/deactivate`)
}

// Expense APIs
export const expenseApi = {
  getExpenses: (associationId: number, startDate?: string, endDate?: string) =>
    api.get<Expense[]>(`/associations/${associationId}/expenses`, {
      params: {
        start_date: startDate,
        end_date: endDate
      }
    }),

  getExpense: (associationId: number, expenseId: number) =>
    api.get<Expense>(`/associations/${associationId}/expenses/${expenseId}`),

  createExpense: (associationId: number, expenseData: ExpenseCreateRequest) =>
    api.post<Expense>(`/associations/${associationId}/expenses`, expenseData),

  updateExpense: (associationId: number, expenseId: number, expenseData: Partial<ExpenseCreateRequest>) =>
    api.put<Expense>(`/associations/${associationId}/expenses/${expenseId}`, expenseData),

  deleteExpense: (associationId: number, expenseId: number) =>
    api.delete(`/associations/${associationId}/expenses/${expenseId}`)
}

// Export the api instance to allow direct access if needed
export default api
