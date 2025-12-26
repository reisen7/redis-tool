import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import type { 
  PubSubMessage, 
  Subscription, 
  WebSocketState,
  SerializablePubSubMessage
} from '@/types/pubsub'
import { 
  generateMessageId, 
  generateSubscriptionId,
  serializeMessage,
  deserializeMessage
} from '@/types/pubsub'

const MESSAGES_STORAGE_KEY = 'redis_manager_pubsub_messages'
const SUBSCRIPTIONS_STORAGE_KEY = 'redis_manager_pubsub_subscriptions'
const MAX_MESSAGES = 1000 // Maximum messages to keep in history

export const usePubSubStore = defineStore('pubsub', () => {
  // State
  const subscriptions = ref<Subscription[]>([])
  const messages = ref<PubSubMessage[]>([])
  const connectionState = ref<WebSocketState>('disconnected')
  const currentConnectionId = ref<string | null>(null)
  const lastError = ref<string | null>(null)
  const reconnectAttempts = ref(0)

  // Computed
  const activeSubscriptions = computed(() => 
    subscriptions.value.filter(sub => sub.active)
  )

  const subscriptionCount = computed(() => subscriptions.value.length)

  const messageCount = computed(() => messages.value.length)

  const isConnected = computed(() => connectionState.value === 'connected')

  const isConnecting = computed(() => 
    connectionState.value === 'connecting' || connectionState.value === 'reconnecting'
  )

  const hasError = computed(() => connectionState.value === 'error')

  // Get messages for a specific channel
  const getMessagesByChannel = computed(() => {
    return (channel: string) => messages.value.filter(msg => msg.channel === channel)
  })

  // Get subscription by channel
  const getSubscriptionByChannel = computed(() => {
    return (channel: string) => subscriptions.value.find(sub => sub.channel === channel)
  })

  // Check if subscribed to a channel
  const isSubscribed = computed(() => {
    return (channel: string) => subscriptions.value.some(sub => sub.channel === channel && sub.active)
  })

  // Initialize from localStorage
  function initialize() {
    try {
      const storedMessages = localStorage.getItem(MESSAGES_STORAGE_KEY)
      const storedSubscriptions = localStorage.getItem(SUBSCRIPTIONS_STORAGE_KEY)

      if (storedMessages) {
        const parsed: SerializablePubSubMessage[] = JSON.parse(storedMessages)
        messages.value = parsed.map(deserializeMessage)
      }

      if (storedSubscriptions) {
        const parsed = JSON.parse(storedSubscriptions)
        // Mark all subscriptions as inactive on load (need to re-subscribe)
        subscriptions.value = parsed.map((sub: Subscription) => ({
          ...sub,
          active: false,
          subscribedAt: new Date(sub.subscribedAt)
        }))
      }
    } catch (error) {
      console.error('Failed to initialize pubsub from localStorage:', error)
      messages.value = []
      subscriptions.value = []
    }
  }

  // Persist to localStorage
  function persist() {
    try {
      const serializedMessages = messages.value.map(serializeMessage)
      localStorage.setItem(MESSAGES_STORAGE_KEY, JSON.stringify(serializedMessages))
      localStorage.setItem(SUBSCRIPTIONS_STORAGE_KEY, JSON.stringify(subscriptions.value))
    } catch (error) {
      console.error('Failed to persist pubsub to localStorage:', error)
    }
  }

  // Watch for changes and persist
  watch([messages, subscriptions], () => {
    persist()
  }, { deep: true })

  // Set connection state
  function setConnectionState(state: WebSocketState) {
    connectionState.value = state
    if (state === 'connected') {
      lastError.value = null
      reconnectAttempts.value = 0
    }
  }

  // Set current connection ID
  function setCurrentConnectionId(connectionId: string | null) {
    currentConnectionId.value = connectionId
  }

  // Set error
  function setError(error: string | null) {
    lastError.value = error
    if (error) {
      connectionState.value = 'error'
    }
  }

  // Increment reconnect attempts
  function incrementReconnectAttempts() {
    reconnectAttempts.value++
  }

  // Reset reconnect attempts
  function resetReconnectAttempts() {
    reconnectAttempts.value = 0
  }

  // Add a subscription
  function addSubscription(channel: string, isPattern: boolean = false): Subscription {
    // Check if already exists
    const existing = subscriptions.value.find(sub => sub.channel === channel)
    if (existing) {
      existing.active = true
      existing.subscribedAt = new Date()
      return existing
    }

    const subscription: Subscription = {
      id: generateSubscriptionId(),
      channel,
      isPattern,
      active: true,
      subscribedAt: new Date()
    }
    subscriptions.value.push(subscription)
    return subscription
  }

  // Remove a subscription
  function removeSubscription(channel: string): boolean {
    const index = subscriptions.value.findIndex(sub => sub.channel === channel)
    if (index === -1) return false
    
    subscriptions.value.splice(index, 1)
    return true
  }

  // Mark subscription as active
  function activateSubscription(channel: string): boolean {
    const sub = subscriptions.value.find(s => s.channel === channel)
    if (!sub) return false
    
    sub.active = true
    sub.subscribedAt = new Date()
    return true
  }

  // Mark subscription as inactive
  function deactivateSubscription(channel: string): boolean {
    const sub = subscriptions.value.find(s => s.channel === channel)
    if (!sub) return false
    
    sub.active = false
    return true
  }

  // Mark all subscriptions as inactive
  function deactivateAllSubscriptions() {
    for (const sub of subscriptions.value) {
      sub.active = false
    }
  }

  // Add a message to history
  function addMessage(channel: string, content: string, pattern?: string): PubSubMessage {
    const message: PubSubMessage = {
      id: generateMessageId(),
      channel,
      pattern,
      message: content,
      timestamp: new Date()
    }
    
    messages.value.push(message)
    
    // Trim messages if exceeding max
    if (messages.value.length > MAX_MESSAGES) {
      messages.value = messages.value.slice(-MAX_MESSAGES)
    }
    
    return message
  }

  // Clear all messages (preserves subscriptions)
  // Requirements: 2.8
  function clearMessages() {
    messages.value = []
  }

  // Clear messages for a specific channel
  function clearMessagesByChannel(channel: string) {
    messages.value = messages.value.filter(msg => msg.channel !== channel)
  }

  // Clear all subscriptions
  function clearSubscriptions() {
    subscriptions.value = []
  }

  // Reset store (for logout or connection change)
  function reset() {
    messages.value = []
    subscriptions.value = []
    connectionState.value = 'disconnected'
    currentConnectionId.value = null
    lastError.value = null
    reconnectAttempts.value = 0
    localStorage.removeItem(MESSAGES_STORAGE_KEY)
    localStorage.removeItem(SUBSCRIPTIONS_STORAGE_KEY)
  }

  // Sync subscriptions from server status
  function syncSubscriptionsFromServer(serverSubscriptions: string[]) {
    // Mark subscriptions as active if they're in the server list
    for (const sub of subscriptions.value) {
      sub.active = serverSubscriptions.includes(sub.channel)
    }
    
    // Add any server subscriptions we don't have locally
    for (const channel of serverSubscriptions) {
      if (!subscriptions.value.some(sub => sub.channel === channel)) {
        addSubscription(channel, false)
      }
    }
  }

  return {
    // State
    subscriptions,
    messages,
    connectionState,
    currentConnectionId,
    lastError,
    reconnectAttempts,
    // Computed
    activeSubscriptions,
    subscriptionCount,
    messageCount,
    isConnected,
    isConnecting,
    hasError,
    getMessagesByChannel,
    getSubscriptionByChannel,
    isSubscribed,
    // Actions
    initialize,
    setConnectionState,
    setCurrentConnectionId,
    setError,
    incrementReconnectAttempts,
    resetReconnectAttempts,
    addSubscription,
    removeSubscription,
    activateSubscription,
    deactivateSubscription,
    deactivateAllSubscriptions,
    addMessage,
    clearMessages,
    clearMessagesByChannel,
    clearSubscriptions,
    reset,
    syncSubscriptionsFromServer
  }
})
