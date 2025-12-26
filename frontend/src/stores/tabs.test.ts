import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useTabsStore } from './tabs'
import type { Tab, CommandHistoryItem } from '@/types/tab'

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

describe('Tabs Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorageMock.clear()
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have empty tabs and null activeTabId initially', () => {
      const store = useTabsStore()
      expect(store.tabs).toEqual([])
      expect(store.activeTabId).toBeNull()
      expect(store.activeTab).toBeNull()
      expect(store.tabCount).toBe(0)
      expect(store.hasActiveTabs).toBe(false)
    })
  })

  describe('addTab', () => {
    it('should create a new tab and set it as active', () => {
      const store = useTabsStore()
      const tab = store.addTab('conn-1', 'Test Tab')

      expect(store.tabs.length).toBe(1)
      expect(store.activeTabId).toBe(tab.id)
      expect(store.activeTab).toEqual(tab)
      expect(tab.connectionId).toBe('conn-1')
      expect(tab.title).toBe('Test Tab')
      expect(tab.dividerPosition).toBe(50)
    })

    it('should create multiple tabs and set the latest as active', () => {
      const store = useTabsStore()
      const tab1 = store.addTab('conn-1', 'Tab 1')
      const tab2 = store.addTab('conn-1', 'Tab 2')

      expect(store.tabs.length).toBe(2)
      expect(store.activeTabId).toBe(tab2.id)
      expect(store.tabCount).toBe(2)
    })

    it('should initialize tab with default keys and terminal state', () => {
      const store = useTabsStore()
      const tab = store.addTab('conn-1', 'Test Tab')

      expect(tab.keysState).toEqual({
        pattern: '*',
        cursor: 0,
        selectedKey: null,
        scrollPosition: 0
      })
      expect(tab.terminalState).toEqual({
        history: [],
        inputValue: ''
      })
    })
  })

  describe('removeTab', () => {
    it('should remove a tab and update active tab', () => {
      const store = useTabsStore()
      const tab1 = store.addTab('conn-1', 'Tab 1')
      const tab2 = store.addTab('conn-1', 'Tab 2')

      // Remove active tab (tab2)
      const result = store.removeTab(tab2.id)

      expect(result).toBe(true)
      expect(store.tabs.length).toBe(1)
      expect(store.activeTabId).toBe(tab1.id)
    })

    it('should set activeTabId to null when last tab is removed', () => {
      const store = useTabsStore()
      const tab = store.addTab('conn-1', 'Tab 1')

      store.removeTab(tab.id)

      expect(store.tabs.length).toBe(0)
      expect(store.activeTabId).toBeNull()
      expect(store.hasActiveTabs).toBe(false)
    })

    it('should return false when removing non-existent tab', () => {
      const store = useTabsStore()
      const result = store.removeTab('non-existent-id')
      expect(result).toBe(false)
    })
  })

  describe('removeTabsByConnection', () => {
    it('should remove all tabs for a specific connection', () => {
      const store = useTabsStore()
      store.addTab('conn-1', 'Tab 1')
      store.addTab('conn-1', 'Tab 2')
      store.addTab('conn-2', 'Tab 3')

      const removedCount = store.removeTabsByConnection('conn-1')

      expect(removedCount).toBe(2)
      expect(store.tabs.length).toBe(1)
      const remainingTab = store.tabs[0]
      expect(remainingTab).toBeDefined()
      expect(remainingTab?.connectionId).toBe('conn-2')
    })
  })

  describe('setActiveTab', () => {
    it('should set the active tab', () => {
      const store = useTabsStore()
      const tab1 = store.addTab('conn-1', 'Tab 1')
      store.addTab('conn-1', 'Tab 2')

      const result = store.setActiveTab(tab1.id)

      expect(result).toBe(true)
      expect(store.activeTabId).toBe(tab1.id)
    })

    it('should return false for non-existent tab', () => {
      const store = useTabsStore()
      const result = store.setActiveTab('non-existent')
      expect(result).toBe(false)
    })
  })

  describe('updateKeysState', () => {
    it('should update keys state for a tab', () => {
      const store = useTabsStore()
      const tab = store.addTab('conn-1', 'Tab 1')

      const result = store.updateKeysState(tab.id, {
        pattern: 'user:*',
        selectedKey: 'user:1'
      })

      expect(result).toBe(true)
      expect(store.getTab(tab.id)?.keysState.pattern).toBe('user:*')
      expect(store.getTab(tab.id)?.keysState.selectedKey).toBe('user:1')
      // Other fields should remain unchanged
      expect(store.getTab(tab.id)?.keysState.cursor).toBe(0)
    })
  })

  describe('updateTerminalState', () => {
    it('should update terminal state for a tab', () => {
      const store = useTabsStore()
      const tab = store.addTab('conn-1', 'Tab 1')

      const result = store.updateTerminalState(tab.id, {
        inputValue: 'GET key'
      })

      expect(result).toBe(true)
      expect(store.getTab(tab.id)?.terminalState.inputValue).toBe('GET key')
    })
  })

  describe('addCommandToHistory', () => {
    it('should add command to terminal history', () => {
      const store = useTabsStore()
      const tab = store.addTab('conn-1', 'Tab 1')

      const historyItem: CommandHistoryItem = {
        command: 'GET key',
        result: '"value"',
        timestamp: new Date()
      }

      const result = store.addCommandToHistory(tab.id, historyItem)

      expect(result).toBe(true)
      expect(store.getTab(tab.id)?.terminalState.history.length).toBe(1)
      const historyEntry = store.getTab(tab.id)?.terminalState.history[0]
      expect(historyEntry).toBeDefined()
      expect(historyEntry?.command).toBe('GET key')
    })
  })

  describe('clearTerminalHistory', () => {
    it('should clear terminal history for a tab', () => {
      const store = useTabsStore()
      const tab = store.addTab('conn-1', 'Tab 1')

      store.addCommandToHistory(tab.id, {
        command: 'GET key',
        result: '"value"',
        timestamp: new Date()
      })

      const result = store.clearTerminalHistory(tab.id)

      expect(result).toBe(true)
      expect(store.getTab(tab.id)?.terminalState.history.length).toBe(0)
    })
  })

  describe('updateDividerPosition', () => {
    it('should update divider position', () => {
      const store = useTabsStore()
      const tab = store.addTab('conn-1', 'Tab 1')

      const result = store.updateDividerPosition(tab.id, 70)

      expect(result).toBe(true)
      expect(store.getTab(tab.id)?.dividerPosition).toBe(70)
    })

    it('should clamp divider position between 10 and 90', () => {
      const store = useTabsStore()
      const tab = store.addTab('conn-1', 'Tab 1')

      store.updateDividerPosition(tab.id, 5)
      expect(store.getTab(tab.id)?.dividerPosition).toBe(10)

      store.updateDividerPosition(tab.id, 95)
      expect(store.getTab(tab.id)?.dividerPosition).toBe(90)
    })
  })

  describe('getTabsByConnection', () => {
    it('should return tabs for a specific connection', () => {
      const store = useTabsStore()
      store.addTab('conn-1', 'Tab 1')
      store.addTab('conn-1', 'Tab 2')
      store.addTab('conn-2', 'Tab 3')

      const conn1Tabs = store.getTabsByConnection('conn-1')

      expect(conn1Tabs.length).toBe(2)
      expect(conn1Tabs.every((t: Tab) => t.connectionId === 'conn-1')).toBe(true)
    })
  })

  describe('persistence', () => {
    it('should persist tabs to localStorage', async () => {
      const store = useTabsStore()
      store.addTab('conn-1', 'Tab 1')

      // Wait for the watcher to trigger
      await new Promise(resolve => setTimeout(resolve, 10))

      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        'redis_manager_tabs_anonymous',
        expect.any(String)
      )
    })

    it('should restore tabs from localStorage on initialize', () => {
      const mockTab = {
        id: 'tab_123',
        connectionId: 'conn-1',
        title: 'Stored Tab',
        type: 'workspace',
        keysState: { pattern: '*', cursor: 0, selectedKey: null, scrollPosition: 0 },
        terminalState: { history: [], inputValue: '' },
        dividerPosition: 50,
        createdAt: new Date().toISOString()
      }

      localStorageMock.getItem.mockImplementation((key: string) => {
        if (key === 'redis_manager_tabs_123') return JSON.stringify([mockTab])
        if (key === 'redis_manager_active_tab_123') return 'tab_123'
        return null
      })

      const store = useTabsStore()
      store.initialize(123) // Pass user ID

      expect(store.tabs.length).toBe(1)
      const restoredTab = store.tabs[0]
      expect(restoredTab).toBeDefined()
      expect(restoredTab?.title).toBe('Stored Tab')
      expect(store.activeTabId).toBe('tab_123')
    })
  })

  describe('reset', () => {
    it('should clear all tabs and localStorage', () => {
      const store = useTabsStore()
      store.initialize(456) // Initialize with user ID first
      store.addTab('conn-1', 'Tab 1')
      store.addTab('conn-1', 'Tab 2')

      store.reset()

      expect(store.tabs.length).toBe(0)
      expect(store.activeTabId).toBeNull()
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('redis_manager_tabs_456')
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('redis_manager_active_tab_456')
    })
  })
})
