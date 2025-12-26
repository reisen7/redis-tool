import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from './auth'
import type { User, LoginResponse, Verify2FAResponse } from '@/types/auth'

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: vi.fn((key: string) => store[key] || null),
    setItem: vi.fn((key: string, value: string) => { store[key] = value }),
    removeItem: vi.fn((key: string) => { delete store[key] }),
    clear: vi.fn(() => { store = {} })
  }
})()

Object.defineProperty(window, 'localStorage', { value: localStorageMock })

describe('Auth Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorageMock.clear()
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have null user and token initially', () => {
      const store = useAuthStore()
      expect(store.user).toBeNull()
      expect(store.token).toBeNull()
      expect(store.isAuthenticated).toBe(false)
    })

    it('should not be admin when not authenticated', () => {
      const store = useAuthStore()
      expect(store.isAdmin).toBe(false)
    })
  })

  describe('setAuth', () => {
    it('should set user and token from login response', () => {
      const store = useAuthStore()
      const mockUser: User = {
        id: 1,
        username: 'testuser',
        role: 'user',
        enabled: true,
        totpEnabled: false,
        mustChangePassword: false,
        createdAt: '2024-01-01T00:00:00Z'
      }
      const loginResponse: LoginResponse = {
        success: true,
        token: 'test-jwt-token',
        user: mockUser
      }

      store.setAuth(loginResponse)

      expect(store.user).toEqual(mockUser)
      expect(store.token).toBe('test-jwt-token')
      expect(store.isAuthenticated).toBe(true)
      expect(localStorageMock.setItem).toHaveBeenCalledWith('redis_manager_token', 'test-jwt-token')
    })

    it('should identify admin users correctly', () => {
      const store = useAuthStore()
      const adminUser: User = {
        id: 1,
        username: 'admin',
        role: 'admin',
        enabled: true,
        totpEnabled: false,
        mustChangePassword: false,
        createdAt: '2024-01-01T00:00:00Z'
      }
      const loginResponse: LoginResponse = {
        success: true,
        token: 'admin-token',
        user: adminUser
      }

      store.setAuth(loginResponse)

      expect(store.isAdmin).toBe(true)
    })
  })

  describe('clearAuth', () => {
    it('should clear user and token on logout', () => {
      const store = useAuthStore()
      // First set auth
      store.setAuth({
        success: true,
        token: 'test-token',
        user: {
          id: 1,
          username: 'test',
          role: 'user',
          enabled: true,
          totpEnabled: false,
          mustChangePassword: false,
          createdAt: '2024-01-01T00:00:00Z'
        }
      })

      // Then clear
      store.clearAuth()

      expect(store.user).toBeNull()
      expect(store.token).toBeNull()
      expect(store.isAuthenticated).toBe(false)
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('redis_manager_token')
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('redis_manager_user')
    })
  })

  describe('2FA Flow', () => {
    it('should set temp token for 2FA flow', () => {
      const store = useAuthStore()
      store.setTempToken('temp-2fa-token')
      
      expect(store.tempToken).toBe('temp-2fa-token')
      expect(store.requires2FA).toBe(true)
    })

    it('should complete 2FA and set full auth', () => {
      const store = useAuthStore()
      store.setTempToken('temp-2fa-token')
      
      const response: Verify2FAResponse = {
        success: true,
        token: 'full-jwt-token',
        user: {
          id: 1,
          username: 'test',
          role: 'user',
          enabled: true,
          totpEnabled: true,
          mustChangePassword: false,
          createdAt: '2024-01-01T00:00:00Z'
        }
      }

      store.complete2FA(response)

      expect(store.token).toBe('full-jwt-token')
      expect(store.tempToken).toBeNull()
      expect(store.requires2FA).toBe(false)
      expect(store.isAuthenticated).toBe(true)
    })
  })

  describe('initialize', () => {
    it('should restore auth from localStorage', () => {
      const mockUser: User = {
        id: 1,
        username: 'stored-user',
        role: 'user',
        enabled: true,
        totpEnabled: false,
        mustChangePassword: false,
        createdAt: '2024-01-01T00:00:00Z'
      }
      
      localStorageMock.getItem.mockImplementation((key: string) => {
        if (key === 'redis_manager_token') return 'stored-token'
        if (key === 'redis_manager_user') return JSON.stringify(mockUser)
        return null
      })

      const store = useAuthStore()
      store.initialize()

      expect(store.token).toBe('stored-token')
      expect(store.user).toEqual(mockUser)
      expect(store.isAuthenticated).toBe(true)
    })

    it('should clear auth if stored user is invalid JSON', () => {
      localStorageMock.getItem.mockImplementation((key: string) => {
        if (key === 'redis_manager_token') return 'stored-token'
        if (key === 'redis_manager_user') return 'invalid-json'
        return null
      })

      const store = useAuthStore()
      store.initialize()

      expect(store.token).toBeNull()
      expect(store.user).toBeNull()
    })
  })

  describe('getToken and hasToken', () => {
    it('should return token when authenticated', () => {
      const store = useAuthStore()
      store.setAuth({
        success: true,
        token: 'my-token',
        user: {
          id: 1,
          username: 'test',
          role: 'user',
          enabled: true,
          totpEnabled: false,
          mustChangePassword: false,
          createdAt: '2024-01-01T00:00:00Z'
        }
      })

      expect(store.getToken()).toBe('my-token')
      expect(store.hasToken()).toBe(true)
    })

    it('should return null when not authenticated', () => {
      const store = useAuthStore()
      expect(store.getToken()).toBeNull()
      expect(store.hasToken()).toBe(false)
    })
  })
})
