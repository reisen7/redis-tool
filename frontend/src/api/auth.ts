import http from './http'
import type {
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
  Verify2FARequest,
  Verify2FAResponse,
  Setup2FAResponse,
  Activate2FARequest,
  Disable2FARequest,
  ApiResponse,
  User
} from '@/types/auth'

export const authApi = {
  // Login with username and password
  login(data: LoginRequest) {
    return http.post<LoginResponse>('/auth/login', data)
  },

  // Register new user
  register(data: RegisterRequest) {
    return http.post<RegisterResponse>('/auth/register', data)
  },

  // Verify 2FA code during login
  verify2FA(data: Verify2FARequest) {
    return http.post<Verify2FAResponse>('/auth/verify-2fa', data)
  },

  // Logout
  logout() {
    return http.post<ApiResponse>('/auth/logout')
  },

  // Get current user info
  me() {
    return http.get<ApiResponse<User>>('/auth/me')
  },

  // Setup 2FA (get QR code and secret)
  setup2FA() {
    return http.post<Setup2FAResponse>('/2fa/setup')
  },

  // Activate 2FA after scanning QR code
  activate2FA(data: Activate2FARequest) {
    return http.post<ApiResponse>('/2fa/verify', data)
  },

  // Disable 2FA
  disable2FA(data: Disable2FARequest) {
    return http.post<ApiResponse>('/2fa/disable', data)
  }
}

export default authApi
