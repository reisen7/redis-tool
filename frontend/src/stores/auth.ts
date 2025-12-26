import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, LoginResponse, Verify2FAResponse } from '@/types/auth'
import { useConnectionStore } from './connection'
import { useTabsStore } from './tabs'

const TOKEN_KEY = 'redis_manager_token'
const USER_KEY = 'redis_manager_user'

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const isLoading = ref(false)
  const tempToken = ref<string | null>(null) // For 2FA flow

  // Computed
  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const requires2FA = computed(() => !!tempToken.value)

  // Initialize from localStorage
  function initialize() {
    const storedToken = localStorage.getItem(TOKEN_KEY)
    const storedUser = localStorage.getItem(USER_KEY)
    
    if (storedToken && storedUser) {
      try {
        token.value = storedToken
        user.value = JSON.parse(storedUser)
      } catch {
        clearAuth()
      }
    }
  }

  // Set auth data after successful login
  function setAuth(loginResponse: LoginResponse) {
    if (loginResponse.token && loginResponse.user) {
      token.value = loginResponse.token
      user.value = loginResponse.user
      localStorage.setItem(TOKEN_KEY, loginResponse.token)
      localStorage.setItem(USER_KEY, JSON.stringify(loginResponse.user))
      tempToken.value = null
      
      // Clear previous session's connection and tabs state
      const connectionStore = useConnectionStore()
      const tabsStore = useTabsStore()
      connectionStore.reset()
      tabsStore.reset()
    }
  }

  // Set temp token for 2FA flow
  function setTempToken(tempTokenValue: string) {
    tempToken.value = tempTokenValue
  }

  // Complete 2FA verification
  function complete2FA(response: Verify2FAResponse) {
    if (response.token && response.user) {
      token.value = response.token
      user.value = response.user
      localStorage.setItem(TOKEN_KEY, response.token)
      localStorage.setItem(USER_KEY, JSON.stringify(response.user))
      tempToken.value = null
      
      // Clear previous session's connection and tabs state
      const connectionStore = useConnectionStore()
      const tabsStore = useTabsStore()
      connectionStore.reset()
      tabsStore.reset()
    }
  }

  // Clear auth data (logout)
  function clearAuth() {
    token.value = null
    user.value = null
    tempToken.value = null
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
    
    // Clear connection and tabs state
    const connectionStore = useConnectionStore()
    const tabsStore = useTabsStore()
    connectionStore.reset()
    tabsStore.reset()
  }

  // Update user data
  function updateUser(updatedUser: User) {
    user.value = updatedUser
    localStorage.setItem(USER_KEY, JSON.stringify(updatedUser))
  }

  // Set loading state
  function setLoading(loading: boolean) {
    isLoading.value = loading
  }

  // Get token for API requests
  function getToken(): string | null {
    return token.value
  }

  // Check if token exists
  function hasToken(): boolean {
    return !!token.value
  }

  return {
    // State
    user,
    token,
    isLoading,
    tempToken,
    // Computed
    isAuthenticated,
    isAdmin,
    requires2FA,
    // Actions
    initialize,
    setAuth,
    setTempToken,
    complete2FA,
    clearAuth,
    updateUser,
    setLoading,
    getToken,
    hasToken
  }
})
