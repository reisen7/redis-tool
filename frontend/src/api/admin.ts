import http from './http'
import type { User, UserRole, ApiResponse } from '@/types/auth'

// ============================================================================
// Types
// ============================================================================

// Create user request
export interface CreateUserRequest {
  username: string
  password: string
  role: UserRole
}

// Update user request
export interface UpdateUserRequest {
  enabled?: boolean
  role?: UserRole
}

// Update setting request
export interface UpdateSettingRequest {
  key: string
  value: string
}

// System settings map
export interface SystemSettings {
  registration_enabled?: string
  [key: string]: string | undefined
}

// Reset password response
export interface ResetPasswordResponse {
  tempPassword: string
}

// ============================================================================
// Admin User Management API (Requirements: 7.1, 7.2)
// ============================================================================

export const adminApi = {
  // Get all users
  // GET /api/admin/users
  // Requirements: 7.1
  listUsers() {
    return http.get<ApiResponse<User[]>>('/admin/users')
  },

  // Create a new user
  // POST /api/admin/users
  // Requirements: 7.2
  createUser(data: CreateUserRequest) {
    return http.post<ApiResponse<User>>('/admin/users', data)
  },

  // Get a specific user
  // GET /api/admin/users/:id
  getUser(id: number) {
    return http.get<ApiResponse<User>>(`/admin/users/${id}`)
  },

  // Update a user
  // PUT /api/admin/users/:id
  // Requirements: 7.3, 7.4
  updateUser(id: number, data: UpdateUserRequest) {
    return http.put<ApiResponse<User>>(`/admin/users/${id}`, data)
  },

  // Delete a user
  // DELETE /api/admin/users/:id
  // Requirements: 7.6
  deleteUser(id: number) {
    return http.delete<ApiResponse>(`/admin/users/${id}`)
  },

  // Reset user password
  // POST /api/admin/users/:id/reset-password
  // Requirements: 7.5
  resetPassword(id: number) {
    return http.post<ApiResponse<ResetPasswordResponse>>(`/admin/users/${id}/reset-password`)
  },

  // ============================================================================
  // Admin Settings API (Requirements: 8.1, 8.2)
  // ============================================================================

  // Get all system settings
  // GET /api/admin/settings
  // Requirements: 8.2
  getSettings() {
    return http.get<ApiResponse<SystemSettings>>('/admin/settings')
  },

  // Update a system setting
  // PUT /api/admin/settings
  // Requirements: 8.1
  updateSetting(data: UpdateSettingRequest) {
    return http.put<ApiResponse>('/admin/settings', data)
  }
}

export default adminApi
