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
  ExpenseDistributionResponse,
  LoginRequest,
  LoginResponse,
  Owner,
  Unit, UnitReportDetails
} from '@/types/api'
import config from '@/config'
import { useAuthStore } from '@/stores/auth'
import type { InternalAxiosRequestConfig } from 'axios';
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
  (reqConfig: InternalAxiosRequestConfig) => {
    if (
      reqConfig.url &&
      (reqConfig.url.includes('/login') || reqConfig.url.includes('/refresh'))
    ) {
      return reqConfig;
    }
    const token = localStorage.getItem(config.authTokenKey);
    if (token && reqConfig.headers) {
      reqConfig.headers.Authorization = `Bearer ${token}`;
    }
    return reqConfig;
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  }
);

// Response interceptor
api.interceptors.response.use(
  (response: AxiosResponse) => {
    return response
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as AxiosRequestConfig & { _retry?: boolean }

    if (!originalRequest) {
      return Promise.reject(error)
    }
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true // Mark that we've tried refreshing for this request

      try {
        const refreshToken = localStorage.getItem(config.refreshTokenKey)
        if (refreshToken) {
          const authStore = useAuthStore()
          const success = await authStore.refreshAccessToken()

          if (success) {
            const newToken = localStorage.getItem(config.authTokenKey)
            if (originalRequest.headers) {
              originalRequest.headers.Authorization = `Bearer ${newToken}`
            }
            return api(originalRequest)
          }
        }
        localStorage.removeItem(config.authTokenKey)
        localStorage.removeItem(config.refreshTokenKey)

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
    api.get<Unit>(`/associations/${associationId}/buildings/${buildingId}/units/${unitId}`),

  updateUnit: (associationId: number, buildingId: number, unitId: number, unitData: Partial<Unit>) =>
    api.put<Unit>(`/associations/${associationId}/buildings/${buildingId}/units/${unitId}`, unitData),

  getUnitReport: (associationId: number, buildingId: number, unitId: number) =>
    api.get<UnitReportDetails>(`/associations/${associationId}/buildings/${buildingId}/units/${unitId}/report`),

  getUnitOwners: (associationId: number, buildingId: number, unitId: number) =>
    api.get<Owner[]>(`/associations/${associationId}/buildings/${buildingId}/units/${unitId}/owners`),

  getUnitOwnerships: (associationId: number, buildingId: number, unitId: number) =>
    api.get<any[]>(`/associations/${associationId}/buildings/${buildingId}/units/${unitId}/ownerships`)
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
  getOwner: (associationId: number, ownerId: number) =>
    api.get<Owner>(`/associations/${associationId}/owners/${ownerId}`),

  updateOwner: (associationId: number, ownerId: number, ownerData: {
    name: string;
    identification_number: string;
    contact_phone: string;
    contact_email: string;
  }) =>
    api.put<Owner>(`/associations/${associationId}/owners/${ownerId}`, ownerData),
  getOwnerReport: (associationId: number, includeUnits: boolean = false, includeCoOwners: boolean = false) =>
    api.get(`/associations/${associationId}/owners/report`, {
      params: {
        units: includeUnits ? 'true' : 'false',
        co_owners: includeCoOwners ? 'true' : 'false'
      }
    }),
  getVotingOwners: (associationId: number) =>
    api.get<VotingOwner[]>(`/associations/${associationId}/owners/voters`),
  createOwner: (associationId: number, ownerData: {
    name: string;
    identification_number: string;
    contact_phone: string;
    contact_email: string;
  }) =>
    api.post<Owner>(`/associations/${associationId}/owners`, ownerData)
}
// Ownership APIs
export const ownershipApi = {
  // Disable an existing ownership
  setOwnershipVoting: (associationId: number, buildingId: number, unitId: number, ownershipId: number) =>
    api.post(`/associations/${associationId}/buildings/${buildingId}/units/${unitId}/ownerships/${ownershipId}/voting`),

  disableOwnership: (associationId: number, ownershipId: number, endDate?: Date) => {
    const payload = endDate ? { end_date: endDate.toISOString() } : {}
    return api.put(`/associations/${associationId}/ownerships/${ownershipId}/disable`, payload)
  },

  // Get ownership details by ID
  getOwnership: (associationId: number, ownershipId: number) =>
    api.get(`/associations/${associationId}/ownerships/${ownershipId}`),

  // Create a new ownership for a unit
  createUnitOwnership: (associationId: number, buildingId: number, unitId: number, ownershipData: {
    owner_id: number;
    start_date: string;
    end_date?: string | null;
    registration_document: string;
    registration_date: string;
    is_exclusive?: boolean;
  }) =>
    api.post(`/associations/${associationId}/buildings/${buildingId}/units/${unitId}/ownerships`, ownershipData)
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
    api.delete(`/associations/${associationId}/expenses/${expenseId}`),

  getExpenseDistribution: (associationId: number, params: {
    start_date?: string;
    end_date?: string;
    category_id?: number | null;
    category_type?: string | null;
    category_family?: string | null;
    distribution_method?: 'area' | 'count' | 'equal';
    unit_type?: string | null;
    include_details?: boolean;
  }) =>
    api.get<ExpenseDistributionResponse>(`/associations/${associationId}/expenses/distribution`, {
      params: {
        start_date: params.start_date,
        end_date: params.end_date,
        category_id: params.category_id || undefined,
        category_type: params.category_type || undefined,
        category_family: params.category_family || undefined,
        distribution_method: params.distribution_method || 'area',
        unit_type: params.unit_type || undefined,
        include_details: params.include_details ? 'true' : undefined
      }
    })
}

export default api;
