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
  CategoryUpdateRequest,
  CategoryUsageResponse,
  Expense,
  ExpenseCreateRequest,
  ExpenseDistributionResponse,
  LoginRequest,
  LoginResponse,
  Owner,
  Unit, UnitReportDetails, VotingOwner,
  ResetPasswordResponse, ResetPasswordRequest,
  Gathering,
  GatheringCreateRequest,
  GatheringUpdateRequest,
  GatheringStatusUpdateRequest,
  VotingMatter,
  VotingMatterCreateRequest,
  VotingMatterUpdateRequest,
  GatheringParticipant,
  ParticipantCreateRequest,
  ParticipantCheckInRequest,
  BallotSubmissionRequest,
  VotingResults,
  QualifiedUnit,
  NonParticipatingOwner,
  GatheringStatistics,
  VotingAuditLog,
  NotificationCreateRequest
} from '@/types/api'
import config from '@/config'
import { attemptTokenRefresh, performLogout, getAuthToken } from '@/services/auth-service'
import type { InternalAxiosRequestConfig } from 'axios'
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
      return reqConfig
    }
    const token = getAuthToken()
    if (token && reqConfig.headers) {
      reqConfig.headers.Authorization = `Bearer ${token}`
    }
    return reqConfig
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

// Flag to prevent multiple simultaneous logout attempts
let isRefreshing = false
let refreshPromise: Promise<boolean> | null = null

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

    // Handle 401 errors (unauthorized)
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      // If we're already refreshing, wait for the existing refresh
      if (isRefreshing && refreshPromise) {
        try {
          const success = await refreshPromise
          if (success) {
            const newToken = getAuthToken()
            if (originalRequest.headers && newToken) {
              originalRequest.headers.Authorization = `Bearer ${newToken}`
            }
            return api(originalRequest)
          }
        } catch {
          // If refresh failed, continue to logout
        }
      }

      // Start refresh process
      if (!isRefreshing) {
        isRefreshing = true
        refreshPromise = attemptTokenRefresh()
        
        try {
          const success = await refreshPromise
          if (success) {
            const newToken = getAuthToken()
            if (originalRequest.headers && newToken) {
              originalRequest.headers.Authorization = `Bearer ${newToken}`
            }
            return api(originalRequest)
          }
        } catch {
          // If refresh failed, continue to logout
        } finally {
          isRefreshing = false
          refreshPromise = null
        }
      }

      // If we get here, refresh failed or no refresh token available
      performLogout()
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
    }),

  resetPassword: (payload: ResetPasswordRequest) =>
    api.post<ResetPasswordResponse>('/password-reset/reset', payload)
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
  getVotingOwners: (associationId: number, filters?: {
    unit_types?: string;     // comma-separated list: "apartment,parking"
    floor?: number;          // specific floor number
    entrance?: number;       // specific entrance number
  }) =>
    api.get<VotingOwner[]>(`/associations/${associationId}/owners/voters`, {
      params: filters || {}
    }),
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
  // Get active categories only (existing endpoint)
  getCategories: (associationId: number) =>
    api.get<Category[]>(`/associations/${associationId}/categories`),

  // Get all categories with optional inactive filter
  getAllCategories: (associationId: number, includeInactive: boolean = false) =>
    api.get<Category[]>(`/associations/${associationId}/categories/all`, {
      params: { include_inactive: includeInactive }
    }),

  // Get single category
  getCategory: (associationId: number, categoryId: number) =>
    api.get<Category>(`/associations/${associationId}/categories/${categoryId}`),

  // Create new category
  createCategory: (associationId: number, categoryData: CategoryCreateRequest) =>
    api.post<Category>(`/associations/${associationId}/categories`, categoryData),

  // Update existing category
  updateCategory: (associationId: number, categoryId: number, categoryData: CategoryUpdateRequest) =>
    api.put<Category>(`/associations/${associationId}/categories/${categoryId}`, categoryData),

  // Deactivate category (soft delete)
  deactivateCategory: (associationId: number, categoryId: number) =>
    api.put<ApiResponse<null>>(`/associations/${associationId}/categories/${categoryId}/deactivate`),

  // Reactivate category
  reactivateCategory: (associationId: number, categoryId: number) =>
    api.put<ApiResponse<null>>(`/associations/${associationId}/categories/${categoryId}/reactivate`),

  // Get category usage statistics
  getCategoryUsage: (associationId: number, categoryId: number) =>
    api.get<CategoryUsageResponse>(`/associations/${associationId}/categories/${categoryId}/usage`),

  // Bulk deactivate categories
  bulkDeactivate: (associationId: number, categoryIds: number[]) =>
    api.post<ApiResponse<null>>(`/associations/${associationId}/categories/bulk-deactivate`, {
      ids: categoryIds
    }),

  // Bulk reactivate categories
  bulkReactivate: (associationId: number, categoryIds: number[]) =>
    api.post<ApiResponse<null>>(`/associations/${associationId}/categories/bulk-reactivate`, {
      ids: categoryIds
    })
}

// Expense APIs
export const expenseApi = {
  getExpenses: (associationId: number, startDate?: string, endDate?: string, categoryId?: number) =>
    api.get<Expense[]>(`/associations/${associationId}/expenses`, {
      params: {
        start_date: startDate,
        end_date: endDate,
        category_id: categoryId
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

// Gathering APIs
export const gatheringApi = {
  // Get all gatherings for an association
  getGatherings: (associationId: number) =>
    api.get<Gathering[]>(`/associations/${associationId}/gatherings`),

  // Get a specific gathering
  getGathering: (associationId: number, gatheringId: number) =>
    api.get<Gathering>(`/associations/${associationId}/gatherings/${gatheringId}`),

  // Create a new gathering
  createGathering: (associationId: number, gatheringData: GatheringCreateRequest) =>
    api.post<Gathering>(`/associations/${associationId}/gatherings`, gatheringData),

  // Update an existing gathering
  updateGathering: (associationId: number, gatheringId: number, gatheringData: GatheringUpdateRequest) =>
    api.put<Gathering>(`/associations/${associationId}/gatherings/${gatheringId}`, gatheringData),

  // Update gathering status
  updateGatheringStatus: (associationId: number, gatheringId: number, statusData: GatheringStatusUpdateRequest) =>
    api.put<Gathering>(`/associations/${associationId}/gatherings/${gatheringId}/status`, statusData),

  // Get gathering statistics
  getGatheringStats: (associationId: number, gatheringId: number) =>
    api.get<GatheringStatistics>(`/associations/${associationId}/gatherings/${gatheringId}/stats`),

  // Get qualified units for a gathering
  getQualifiedUnits: (associationId: number, gatheringId: number) =>
    api.get<QualifiedUnit[]>(`/associations/${associationId}/gatherings/${gatheringId}/qualified-units`),

  // Get non-participating owners
  getNonParticipatingOwners: (associationId: number, gatheringId: number) =>
    api.get<NonParticipatingOwner[]>(`/associations/${associationId}/gatherings/${gatheringId}/non-participating-owners`),

  // Get eligible voters with their available units
  getEligibleVoters: (associationId: number, gatheringId: number) =>
    api.get<any[]>(`/associations/${associationId}/gatherings/${gatheringId}/eligible-voters`),

  // Get audit logs for a gathering
  getAuditLogs: (associationId: number, gatheringId: number) =>
    api.get<VotingAuditLog[]>(`/associations/${associationId}/gatherings/${gatheringId}/audit-logs`),

  // Send notifications
  sendNotification: (associationId: number, gatheringId: number, notificationData: NotificationCreateRequest) =>
    api.post<ApiResponse<null>>(`/associations/${associationId}/gatherings/${gatheringId}/notifications`, notificationData)
}

// Voting Matter APIs
export const votingMatterApi = {
  // Get all voting matters for a gathering
  getVotingMatters: (associationId: number, gatheringId: number) =>
    api.get<VotingMatter[]>(`/associations/${associationId}/gatherings/${gatheringId}/matters`),

  // Get a specific voting matter
  getVotingMatter: (associationId: number, gatheringId: number, matterId: number) =>
    api.get<VotingMatter>(`/associations/${associationId}/gatherings/${gatheringId}/matters/${matterId}`),

  // Create a new voting matter
  createVotingMatter: (associationId: number, gatheringId: number, matterData: VotingMatterCreateRequest) =>
    api.post<VotingMatter>(`/associations/${associationId}/gatherings/${gatheringId}/matters`, matterData),

  // Update an existing voting matter
  updateVotingMatter: (associationId: number, gatheringId: number, matterId: number, matterData: VotingMatterUpdateRequest) =>
    api.put<VotingMatter>(`/associations/${associationId}/gatherings/${gatheringId}/matters/${matterId}`, matterData),

  // Delete a voting matter
  deleteVotingMatter: (associationId: number, gatheringId: number, matterId: number) =>
    api.delete(`/associations/${associationId}/gatherings/${gatheringId}/matters/${matterId}`)
}

// Participant APIs
export const participantApi = {
  // Get all participants for a gathering
  getParticipants: (associationId: number, gatheringId: number) =>
    api.get<GatheringParticipant[]>(`/associations/${associationId}/gatherings/${gatheringId}/participants`),

  // Get a specific participant
  getParticipant: (associationId: number, gatheringId: number, participantId: number) =>
    api.get<GatheringParticipant>(`/associations/${associationId}/gatherings/${gatheringId}/participants/${participantId}`),

  // Add a new participant
  addParticipant: (associationId: number, gatheringId: number, participantData: ParticipantCreateRequest) =>
    api.post<GatheringParticipant>(`/associations/${associationId}/gatherings/${gatheringId}/participants`, participantData),

  // Update a participant
  updateParticipant: (associationId: number, gatheringId: number, participantId: number, participantData: Partial<ParticipantCreateRequest>) =>
    api.put<GatheringParticipant>(`/associations/${associationId}/gatherings/${gatheringId}/participants/${participantId}`, participantData),

  // Remove a participant
  removeParticipant: (associationId: number, gatheringId: number, participantId: number) =>
    api.delete(`/associations/${associationId}/gatherings/${gatheringId}/participants/${participantId}`),

  // Check in a participant
  checkInParticipant: (associationId: number, gatheringId: number, participantId: number, checkInData: ParticipantCheckInRequest) =>
    api.post<GatheringParticipant>(`/associations/${associationId}/gatherings/${gatheringId}/participants/${participantId}/checkin`, checkInData)
}

// Voting APIs
export const votingApi = {
  // Submit a ballot (new simplified endpoint)
  submitBallot: (associationId: number, gatheringId: number, ballotData: any) =>
    api.post<any>(`/associations/${associationId}/gatherings/${gatheringId}/ballot`, ballotData),

  // Get voting results
  getResults: (associationId: number, gatheringId: number) =>
    api.get<VotingResults>(`/associations/${associationId}/gatherings/${gatheringId}/results`),

  // Get ballots for a gathering
  getBallots: (associationId: number, gatheringId: number) =>
    api.get<any[]>(`/associations/${associationId}/gatherings/${gatheringId}/ballots`),

  // Download voting results as markdown
  downloadResults: (associationId: number, gatheringId: number) =>
    api.get(`/associations/${associationId}/gatherings/${gatheringId}/download/results`, {
      responseType: 'blob'
    }),

  // Download ballots as markdown
  downloadBallots: (associationId: number, gatheringId: number) =>
    api.get(`/associations/${associationId}/gatherings/${gatheringId}/download/ballots`, {
      responseType: 'blob'
    }),

  // Verify a ballot hash
  verifyBallot: (ballotHash: string) =>
    api.post<ApiResponse<{ valid: boolean; details?: any }>>(`/ballot/verify`, { hash: ballotHash })
}

export default api
