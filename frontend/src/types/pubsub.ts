// Pub/Sub Types for Redis Web Manager

// WebSocket message types (matching backend)
export type WSMessageType = 
  | 'subscribe' 
  | 'unsubscribe' 
  | 'message' 
  | 'error' 
  | 'status' 
  | 'subscribed' 
  | 'unsubscribed'

// WebSocket message structure
export interface WSMessage<T = unknown> {
  type: WSMessageType
  payload: T
}

// Subscribe/Unsubscribe payload
export interface SubscribePayload {
  channel: string
  isPattern: boolean
}

// Pub/Sub message received from Redis
export interface PubSubMessage {
  id: string
  channel: string
  pattern?: string
  message: string
  timestamp: Date
}

// Serializable version for storage
export interface SerializablePubSubMessage {
  id: string
  channel: string
  pattern?: string
  message: string
  timestamp: string
}

// Status payload from server
export interface StatusPayload {
  connected: boolean
  subscriptions: string[]
}

// Error payload from server
export interface ErrorPayload {
  message: string
  code?: string
}

// Subscription state
export interface Subscription {
  id: string
  channel: string
  isPattern: boolean
  active: boolean
  subscribedAt: Date
}

// WebSocket connection state
export type WebSocketState = 'disconnected' | 'connecting' | 'connected' | 'reconnecting' | 'error'

// Publish request
export interface PublishRequest {
  connectionId: number
  channel: string
  message: string
}

// Publish response
export interface PublishResponse {
  success: boolean
  data?: {
    channel: string
    message: string
  }
  error?: string
}

// Helper functions
export function generateMessageId(): string {
  return `msg_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`
}

export function generateSubscriptionId(): string {
  return `sub_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`
}

export function serializeMessage(msg: PubSubMessage): SerializablePubSubMessage {
  return {
    ...msg,
    timestamp: msg.timestamp.toISOString()
  }
}

export function deserializeMessage(data: SerializablePubSubMessage): PubSubMessage {
  return {
    ...data,
    timestamp: new Date(data.timestamp)
  }
}
