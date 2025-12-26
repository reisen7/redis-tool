import { ref, onUnmounted, watch } from 'vue'
import { usePubSubStore } from '@/stores/pubsub'
import { useAuthStore } from '@/stores/auth'
import type { 
  WSMessage, 
  SubscribePayload, 
  StatusPayload, 
  ErrorPayload,
  PubSubMessage as PubSubMessageType
} from '@/types/pubsub'

// Reconnection configuration
const INITIAL_RECONNECT_DELAY = 1000 // 1 second
const MAX_RECONNECT_DELAY = 30000 // 30 seconds
const RECONNECT_MULTIPLIER = 2

export interface UseWebSocketOptions {
  connectionId: string
  onMessage?: (message: PubSubMessageType) => void
  onError?: (error: string) => void
  onStatusChange?: (connected: boolean) => void
}

export function useWebSocket(options: UseWebSocketOptions) {
  const pubsubStore = usePubSubStore()
  const authStore = useAuthStore()
  
  const ws = ref<WebSocket | null>(null)
  const reconnectTimeout = ref<ReturnType<typeof setTimeout> | null>(null)
  const reconnectDelay = ref(INITIAL_RECONNECT_DELAY)
  const isManualClose = ref(false)

  // Get WebSocket URL
  function getWebSocketUrl(): string {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const token = authStore.token
    return `${protocol}//${host}/api/pubsub/ws?connectionId=${options.connectionId}&token=${token}`
  }

  // Connect to WebSocket
  // Requirements: 2.6
  function connect() {
    if (ws.value?.readyState === WebSocket.OPEN || ws.value?.readyState === WebSocket.CONNECTING) {
      return
    }

    isManualClose.value = false
    pubsubStore.setConnectionState('connecting')
    pubsubStore.setCurrentConnectionId(options.connectionId)

    try {
      const url = getWebSocketUrl()
      ws.value = new WebSocket(url)

      ws.value.onopen = handleOpen
      ws.value.onmessage = handleMessage
      ws.value.onerror = handleError
      ws.value.onclose = handleClose
    } catch (error) {
      console.error('WebSocket connection error:', error)
      pubsubStore.setError('无法建立WebSocket连接')
      scheduleReconnect()
    }
  }

  // Disconnect from WebSocket
  function disconnect() {
    isManualClose.value = true
    clearReconnectTimeout()
    
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
    
    pubsubStore.setConnectionState('disconnected')
    pubsubStore.deactivateAllSubscriptions()
  }

  // Handle WebSocket open
  function handleOpen() {
    console.log('WebSocket connected')
    pubsubStore.setConnectionState('connected')
    reconnectDelay.value = INITIAL_RECONNECT_DELAY
    pubsubStore.resetReconnectAttempts()
    options.onStatusChange?.(true)

    // Re-subscribe to previously active subscriptions
    resubscribeAll()
  }

  // Handle WebSocket message
  function handleMessage(event: MessageEvent) {
    try {
      const data: WSMessage = JSON.parse(event.data)
      
      switch (data.type) {
        case 'message':
          handlePubSubMessage(data.payload as PubSubMessageType)
          break
        case 'status':
          handleStatusMessage(data.payload as StatusPayload)
          break
        case 'error':
          handleErrorMessage(data.payload as ErrorPayload)
          break
        case 'subscribed':
          handleSubscribedMessage(data.payload as SubscribePayload)
          break
        case 'unsubscribed':
          handleUnsubscribedMessage(data.payload as SubscribePayload)
          break
        default:
          console.warn('Unknown WebSocket message type:', data.type)
      }
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error)
    }
  }

  // Handle Pub/Sub message from Redis
  // Requirements: 2.3
  function handlePubSubMessage(payload: PubSubMessageType) {
    const message = pubsubStore.addMessage(
      payload.channel,
      payload.message,
      payload.pattern
    )
    options.onMessage?.(message)
  }

  // Handle status message
  function handleStatusMessage(payload: StatusPayload) {
    if (payload.connected) {
      pubsubStore.setConnectionState('connected')
      pubsubStore.syncSubscriptionsFromServer(payload.subscriptions)
    } else {
      pubsubStore.setConnectionState('disconnected')
    }
    options.onStatusChange?.(payload.connected)
  }

  // Handle error message
  function handleErrorMessage(payload: ErrorPayload) {
    console.error('WebSocket error:', payload.message, payload.code)
    pubsubStore.setError(payload.message)
    options.onError?.(payload.message)
  }

  // Handle subscribed confirmation
  function handleSubscribedMessage(payload: SubscribePayload) {
    pubsubStore.activateSubscription(payload.channel)
  }

  // Handle unsubscribed confirmation
  function handleUnsubscribedMessage(payload: SubscribePayload) {
    pubsubStore.deactivateSubscription(payload.channel)
  }

  // Handle WebSocket error
  function handleError(event: Event) {
    console.error('WebSocket error:', event)
    pubsubStore.setError('WebSocket连接错误')
    options.onError?.('WebSocket连接错误')
  }

  // Handle WebSocket close
  // Requirements: 2.7
  function handleClose(event: CloseEvent) {
    console.log('WebSocket closed:', event.code, event.reason)
    ws.value = null
    pubsubStore.deactivateAllSubscriptions()
    options.onStatusChange?.(false)

    if (!isManualClose.value) {
      // Automatic reconnection with backoff
      pubsubStore.setConnectionState('reconnecting')
      scheduleReconnect()
    } else {
      pubsubStore.setConnectionState('disconnected')
    }
  }

  // Schedule reconnection with exponential backoff
  // Requirements: 2.7
  function scheduleReconnect() {
    clearReconnectTimeout()
    
    pubsubStore.incrementReconnectAttempts()
    
    console.log(`Scheduling reconnect in ${reconnectDelay.value}ms (attempt ${pubsubStore.reconnectAttempts})`)
    
    reconnectTimeout.value = setTimeout(() => {
      connect()
      // Increase delay for next attempt (exponential backoff)
      reconnectDelay.value = Math.min(
        reconnectDelay.value * RECONNECT_MULTIPLIER,
        MAX_RECONNECT_DELAY
      )
    }, reconnectDelay.value)
  }

  // Clear reconnect timeout
  function clearReconnectTimeout() {
    if (reconnectTimeout.value) {
      clearTimeout(reconnectTimeout.value)
      reconnectTimeout.value = null
    }
  }

  // Subscribe to a channel
  // Requirements: 2.2
  function subscribe(channel: string, isPattern: boolean = false) {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) {
      console.warn('WebSocket not connected, cannot subscribe')
      return false
    }

    // Add to local store (will be marked active when server confirms)
    pubsubStore.addSubscription(channel, isPattern)

    const message: WSMessage<SubscribePayload> = {
      type: 'subscribe',
      payload: { channel, isPattern }
    }
    
    ws.value.send(JSON.stringify(message))
    return true
  }

  // Unsubscribe from a channel
  // Requirements: 2.5
  function unsubscribe(channel: string, isPattern: boolean = false) {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) {
      console.warn('WebSocket not connected, cannot unsubscribe')
      // Still remove from local store
      pubsubStore.removeSubscription(channel)
      return false
    }

    const message: WSMessage<SubscribePayload> = {
      type: 'unsubscribe',
      payload: { channel, isPattern }
    }
    
    ws.value.send(JSON.stringify(message))
    
    // Remove from local store
    pubsubStore.removeSubscription(channel)
    return true
  }

  // Re-subscribe to all previously active subscriptions
  function resubscribeAll() {
    const subs = pubsubStore.subscriptions
    for (const sub of subs) {
      if (ws.value?.readyState === WebSocket.OPEN) {
        const message: WSMessage<SubscribePayload> = {
          type: 'subscribe',
          payload: { channel: sub.channel, isPattern: sub.isPattern }
        }
        ws.value.send(JSON.stringify(message))
      }
    }
  }

  // Check if WebSocket is connected
  function isConnected(): boolean {
    return ws.value?.readyState === WebSocket.OPEN
  }

  // Cleanup on unmount
  onUnmounted(() => {
    disconnect()
  })

  // Watch for auth changes - disconnect if logged out
  watch(() => authStore.isAuthenticated, (isAuth) => {
    if (!isAuth && ws.value) {
      disconnect()
    }
  })

  return {
    connect,
    disconnect,
    subscribe,
    unsubscribe,
    isConnected,
    // Expose state from store for convenience
    connectionState: pubsubStore.connectionState,
    lastError: pubsubStore.lastError,
    reconnectAttempts: pubsubStore.reconnectAttempts
  }
}
