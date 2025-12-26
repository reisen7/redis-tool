// Tab Types for Multi-Tab Workspace

// Tab type enum
export type TabType = 'workspace' | 'pubsub'

// Command history item for terminal
export interface CommandHistoryItem {
  command: string
  result: string
  timestamp: Date
  isError?: boolean
}

// Keys state for a tab
export interface KeysState {
  pattern: string
  cursor: number
  selectedKey: string | null
  scrollPosition: number
}

// Terminal state for a tab
export interface TerminalState {
  history: CommandHistoryItem[]
  inputValue: string
}

// Tab interface
export interface Tab {
  id: string
  connectionId: string
  title: string
  type: TabType
  keysState: KeysState
  terminalState: TerminalState
  dividerPosition: number // percentage (0-100)
  createdAt: Date
}

// Create default keys state
export function createDefaultKeysState(): KeysState {
  return {
    pattern: '*',
    cursor: 0,
    selectedKey: null,
    scrollPosition: 0
  }
}

// Create default terminal state
export function createDefaultTerminalState(): TerminalState {
  return {
    history: [],
    inputValue: ''
  }
}

// Create a new tab
export function createTab(connectionId: string, title: string, type: TabType = 'workspace'): Tab {
  return {
    id: generateTabId(),
    connectionId,
    title,
    type,
    keysState: createDefaultKeysState(),
    terminalState: createDefaultTerminalState(),
    dividerPosition: 50, // 50% split by default
    createdAt: new Date()
  }
}

// Generate unique tab ID
export function generateTabId(): string {
  return `tab_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`
}

// Serializable tab for localStorage (dates as strings)
export interface SerializableTab {
  id: string
  connectionId: string
  title: string
  type: TabType
  keysState: KeysState
  terminalState: {
    history: Array<{
      command: string
      result: string
      timestamp: string
      isError?: boolean
    }>
    inputValue: string
  }
  dividerPosition: number
  createdAt: string
}

// Convert Tab to serializable format
export function serializeTab(tab: Tab): SerializableTab {
  return {
    ...tab,
    terminalState: {
      ...tab.terminalState,
      history: tab.terminalState.history.map(item => ({
        ...item,
        timestamp: item.timestamp.toISOString()
      }))
    },
    createdAt: tab.createdAt.toISOString()
  }
}

// Convert serializable format back to Tab
export function deserializeTab(data: SerializableTab): Tab {
  return {
    ...data,
    type: data.type || 'workspace', // Default to workspace for old data
    terminalState: {
      ...data.terminalState,
      history: data.terminalState.history.map(item => ({
        ...item,
        timestamp: new Date(item.timestamp)
      }))
    },
    createdAt: new Date(data.createdAt)
  }
}
