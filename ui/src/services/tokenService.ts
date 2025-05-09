// src/services/tokenService.ts
import api from './api'

export interface RegistrationToken {
  token: string
  created_by: string
  created_at: string
  expires_at: string
  used_at: string | null
  used_by: string | null
  revoked_at: string | null
  revoked_by: string | null
  description: string
  is_admin: boolean
  status?: string
}

export interface CreateTokenRequest {
  expiration_hours: number
  description: string
  is_admin: boolean
}

export interface ApiResponse<T> {
  data: T
  message?: string
  status: number
}

/**
 * Service for managing registration tokens
 */
const tokenService = {
  /**
   * Get all registration tokens
   */
  getAllTokens: async (): Promise<RegistrationToken[]> => {
    const response = await api.get<RegistrationToken[]>('/admin/tokens')
    return response.data
  },

  /**
   * Create a new registration token
   */
  createToken: async (tokenData: CreateTokenRequest): Promise<RegistrationToken> => {
    const response = await api.post<RegistrationToken>('/admin/tokens', tokenData)
    return response.data
  },

  /**
   * Revoke a registration token
   */
  revokeToken: async (token: string): Promise<ApiResponse<null>> => {
    const response = await api.put<ApiResponse<null>>(`/admin/tokens/${token}/revoke`)
    return response.data
  },

  /**
   * Register a new user with a token
   */
  registerWithToken: async (userData: {
    login: string
    password: string
    token: string
  }): Promise<{ login: string; qrCode: string; createdAt: string }> => {
    const response = await api.post('/users', userData)
    return response.data
  }
}

export default tokenService
