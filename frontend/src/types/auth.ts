// User role types
export type UserRole = 'admin' | 'user'

// User interface
export interface User {
  id: number
  username: string
  role: UserRole
  enabled: boolean
  totpEnabled: boolean
  mustChangePassword: boolean
  createdAt: string
  updatedAt?: string
}

// Login request
export interface LoginRequest {
  username: string
  password: string
}

// Login response
export interface LoginResponse {
  success: boolean
  token?: string
  user?: User
  requires2FA?: boolean
  tempToken?: string
  error?: string
}

// Register request
export interface RegisterRequest {
  username: string
  password: string
}

// Register response
export interface RegisterResponse {
  success: boolean
  message?: string
  error?: string
}

// 2FA verification request
export interface Verify2FARequest {
  tempToken: string
  code: string
}

// 2FA verification response
export interface Verify2FAResponse {
  success: boolean
  token?: string
  user?: User
  error?: string
}

// 2FA setup response
export interface Setup2FAResponse {
  success: boolean
  secret?: string
  qrCode?: string
  recoveryCodes?: string[]
  error?: string
}

// 2FA activation request
export interface Activate2FARequest {
  code: string
}

// 2FA disable request
export interface Disable2FARequest {
  password: string
  code: string
}

// Generic API response
export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: string
  message?: string
}

// Auth state for store
export interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
}
