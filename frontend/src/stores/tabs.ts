import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import type { Tab, KeysState, TerminalState, CommandHistoryItem, SerializableTab, TabType } from '@/types/tab'
import { createTab, serializeTab, deserializeTab } from '@/types/tab'

const TABS_STORAGE_KEY_PREFIX = 'redis_manager_tabs_'
const ACTIVE_TAB_STORAGE_KEY_PREFIX = 'redis_manager_active_tab_'

// Helper to get user-specific storage keys
function getStorageKeys(userId?: number) {
  const suffix = userId ? String(userId) : 'anonymous'
  return {
    tabs: `${TABS_STORAGE_KEY_PREFIX}${suffix}`,
    activeTab: `${ACTIVE_TAB_STORAGE_KEY_PREFIX}${suffix}`
  }
}

export const useTabsStore = defineStore('tabs', () => {
  // State
  const tabs = ref<Tab[]>([])
  const activeTabId = ref<string | null>(null)
  const currentUserId = ref<number | undefined>(undefined)
  
  // Flag to prevent persisting during reset
  let isResetting = false

  // Computed
  const activeTab = computed(() => {
    if (!activeTabId.value) return null
    return tabs.value.find((tab: Tab) => tab.id === activeTabId.value) || null
  })

  const tabCount = computed(() => tabs.value.length)

  const hasActiveTabs = computed(() => tabs.value.length > 0)

  // Get tabs for a specific connection
  const getTabsByConnection = computed(() => {
    return (connectionId: string) => tabs.value.filter((tab: Tab) => tab.connectionId === connectionId)
  })

  // Initialize from localStorage for a specific user
  function initialize(userId?: number) {
    currentUserId.value = userId
    const keys = getStorageKeys(userId)
    
    try {
      const storedTabs = localStorage.getItem(keys.tabs)
      const storedActiveTab = localStorage.getItem(keys.activeTab)

      if (storedTabs) {
        const parsedTabs: SerializableTab[] = JSON.parse(storedTabs)
        tabs.value = parsedTabs.map(deserializeTab)
      } else {
        tabs.value = []
      }

      if (storedActiveTab && tabs.value.some(tab => tab.id === storedActiveTab)) {
        activeTabId.value = storedActiveTab
      } else if (tabs.value.length > 0) {
        const firstTab = tabs.value[0]
        if (firstTab) {
          activeTabId.value = firstTab.id
        }
      } else {
        activeTabId.value = null
      }
    } catch (error) {
      console.error('Failed to initialize tabs from localStorage:', error)
      tabs.value = []
      activeTabId.value = null
    }
  }

  // Persist to localStorage
  function persistTabs() {
    // Skip persisting during reset to avoid race condition
    if (isResetting) return
    
    const keys = getStorageKeys(currentUserId.value)
    
    try {
      const serializedTabs = tabs.value.map(serializeTab)
      localStorage.setItem(keys.tabs, JSON.stringify(serializedTabs))
      
      if (activeTabId.value) {
        localStorage.setItem(keys.activeTab, activeTabId.value)
      } else {
        localStorage.removeItem(keys.activeTab)
      }
    } catch (error) {
      console.error('Failed to persist tabs to localStorage:', error)
    }
  }

  // Watch for changes and persist
  watch([tabs, activeTabId], () => {
    persistTabs()
  }, { deep: true })

  // Create a new tab for a connection
  function addTab(connectionId: string, title: string, type: TabType = 'workspace'): Tab {
    const newTab = createTab(connectionId, title, type)
    tabs.value.push(newTab)
    activeTabId.value = newTab.id
    return newTab
  }

  // Remove a tab
  function removeTab(tabId: string): boolean {
    const index = tabs.value.findIndex(tab => tab.id === tabId)
    if (index === -1) return false

    tabs.value.splice(index, 1)

    // If we removed the active tab, activate another one
    if (activeTabId.value === tabId) {
      if (tabs.value.length > 0) {
        // Prefer the tab at the same index, or the previous one
        const newIndex = Math.min(index, tabs.value.length - 1)
        const newActiveTab = tabs.value[newIndex]
        if (newActiveTab) {
          activeTabId.value = newActiveTab.id
        } else {
          activeTabId.value = null
        }
      } else {
        activeTabId.value = null
      }
    }

    return true
  }

  // Remove all tabs for a connection
  function removeTabsByConnection(connectionId: string): number {
    const tabsToRemove = tabs.value.filter((tab: Tab) => tab.connectionId === connectionId)
    let removedCount = 0

    for (const tab of tabsToRemove) {
      if (removeTab(tab.id)) {
        removedCount++
      }
    }

    return removedCount
  }

  // Set active tab
  function setActiveTab(tabId: string): boolean {
    const tab = tabs.value.find((t: Tab) => t.id === tabId)
    if (!tab) return false
    
    activeTabId.value = tabId
    return true
  }

  // Update tab title
  function updateTabTitle(tabId: string, title: string): boolean {
    const tab = tabs.value.find((t: Tab) => t.id === tabId)
    if (!tab) return false
    
    tab.title = title
    return true
  }

  // Update keys state for a tab
  function updateKeysState(tabId: string, keysState: Partial<KeysState>): boolean {
    const tab = tabs.value.find((t: Tab) => t.id === tabId)
    if (!tab) return false
    
    tab.keysState = { ...tab.keysState, ...keysState }
    return true
  }

  // Update terminal state for a tab
  function updateTerminalState(tabId: string, terminalState: Partial<TerminalState>): boolean {
    const tab = tabs.value.find((t: Tab) => t.id === tabId)
    if (!tab) return false
    
    tab.terminalState = { ...tab.terminalState, ...terminalState }
    return true
  }

  // Add command to terminal history
  function addCommandToHistory(tabId: string, item: CommandHistoryItem): boolean {
    const tab = tabs.value.find((t: Tab) => t.id === tabId)
    if (!tab) return false
    
    tab.terminalState.history.push(item)
    return true
  }

  // Clear terminal history for a tab
  function clearTerminalHistory(tabId: string): boolean {
    const tab = tabs.value.find((t: Tab) => t.id === tabId)
    if (!tab) return false
    
    tab.terminalState.history = []
    return true
  }

  // Update divider position for a tab
  function updateDividerPosition(tabId: string, position: number): boolean {
    const tab = tabs.value.find((t: Tab) => t.id === tabId)
    if (!tab) return false
    
    // Clamp position between 10 and 90
    tab.dividerPosition = Math.max(10, Math.min(90, position))
    return true
  }

  // Get tab by ID
  function getTab(tabId: string): Tab | null {
    return tabs.value.find((t: Tab) => t.id === tabId) || null
  }

  // Check if a tab exists
  function hasTab(tabId: string): boolean {
    return tabs.value.some((t: Tab) => t.id === tabId)
  }

  // Clear all tabs
  function clearAllTabs(): void {
    tabs.value = []
    activeTabId.value = null
  }

  // Clear tabs and storage (for logout)
  function reset(): void {
    isResetting = true
    // First clear localStorage to prevent watch from re-persisting
    const keys = getStorageKeys(currentUserId.value)
    localStorage.removeItem(keys.tabs)
    localStorage.removeItem(keys.activeTab)
    // Then clear state
    tabs.value = []
    activeTabId.value = null
    currentUserId.value = undefined
    isResetting = false
  }

  return {
    // State
    tabs,
    activeTabId,
    // Computed
    activeTab,
    tabCount,
    hasActiveTabs,
    getTabsByConnection,
    // Actions
    initialize,
    addTab,
    removeTab,
    removeTabsByConnection,
    setActiveTab,
    updateTabTitle,
    updateKeysState,
    updateTerminalState,
    addCommandToHistory,
    clearTerminalHistory,
    updateDividerPosition,
    getTab,
    hasTab,
    clearAllTabs,
    reset
  }
})
