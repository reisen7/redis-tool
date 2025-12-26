<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Connection, Refresh } from '@element-plus/icons-vue'
import { useConnectionStore } from '@/stores/connection'
import { usePubSubStore } from '@/stores/pubsub'
import { useWebSocket } from '@/composables/useWebSocket'
import SubscriptionList from './SubscriptionList.vue'
import MessageHistory from './MessageHistory.vue'
import PublishForm from './PublishForm.vue'

const props = defineProps<{
  connectionId?: string
}>()

const connectionStore = useConnectionStore()
const pubsubStore = usePubSubStore()

// WebSocket instance
const wsInstance = ref<ReturnType<typeof useWebSocket> | null>(null)

// Current connection ID - use prop if provided, otherwise use active connection
const currentConnectionId = computed(() => props.connectionId || connectionStore.activeConnectionId)
const connectionName = computed(() => {
  if (props.connectionId) {
    const conn = connectionStore.openConnections.get(props.connectionId)
    return conn?.config.name ?? ''
  }
  return connectionStore.activeConnection?.config.name ?? ''
})

// Connection state
const isConnected = computed(() => pubsubStore.isConnected)
const isConnecting = computed(() => pubsubStore.isConnecting)
const connectionState = computed(() => pubsubStore.connectionState)
const lastError = computed(() => pubsubStore.lastError)

// Initialize pubsub store
onMounted(() => {
  pubsubStore.initialize()
})

// Watch for connection changes
watch(currentConnectionId, (newId, oldId) => {
  if (oldId && wsInstance.value) {
    wsInstance.value.disconnect()
    wsInstance.value = null
  }
  
  if (newId) {
    connectWebSocket()
  }
}, { immediate: true })

// Connect to WebSocket
function connectWebSocket() {
  if (!currentConnectionId.value) {
    ElMessage.warning('请先连接 Redis')
    return
  }

  wsInstance.value = useWebSocket({
    connectionId: currentConnectionId.value,
    onMessage: (msg) => {
      // Message already added to store by the composable
    },
    onError: (error) => {
      ElMessage.error(error)
    },
    onStatusChange: (connected) => {
      if (connected) {
        ElMessage.success('Pub/Sub 连接成功')
      }
    }
  })

  wsInstance.value.connect()
}

// Disconnect WebSocket
function disconnectWebSocket() {
  if (wsInstance.value) {
    wsInstance.value.disconnect()
  }
}

// Toggle connection
function toggleConnection() {
  if (isConnected.value) {
    disconnectWebSocket()
  } else {
    connectWebSocket()
  }
}

// Subscribe to channel
function handleSubscribe(channel: string, isPattern: boolean) {
  if (!wsInstance.value) {
    ElMessage.warning('请先连接 Pub/Sub')
    return
  }
  
  wsInstance.value.subscribe(channel, isPattern)
}

// Unsubscribe from channel
function handleUnsubscribe(channel: string, isPattern: boolean) {
  if (!wsInstance.value) {
    pubsubStore.removeSubscription(channel)
    return
  }
  
  wsInstance.value.unsubscribe(channel, isPattern)
}

// Clear messages
function handleClearMessages() {
  pubsubStore.clearMessages()
}

// Cleanup on unmount
onUnmounted(() => {
  // Don't disconnect on unmount - keep connection alive
  // User can manually disconnect if needed
})

// Get connection status text
const connectionStatusText = computed(() => {
  switch (connectionState.value) {
    case 'connected':
      return '已连接'
    case 'connecting':
      return '连接中...'
    case 'reconnecting':
      return `重连中 (${pubsubStore.reconnectAttempts})`
    case 'error':
      return '连接错误'
    default:
      return '未连接'
  }
})

// Get connection status type for el-tag
const connectionStatusType = computed(() => {
  switch (connectionState.value) {
    case 'connected':
      return 'success'
    case 'connecting':
    case 'reconnecting':
      return 'warning'
    case 'error':
      return 'danger'
    default:
      return 'info'
  }
})
</script>

<template>
  <div class="pubsub-panel">
    <!-- Header -->
    <div class="panel-header">
      <div class="header-left">
        <span class="panel-title">Pub/Sub</span>
        <el-tag 
          :type="connectionStatusType" 
          size="small"
          class="status-tag"
        >
          {{ connectionStatusText }}
        </el-tag>
        <span v-if="connectionName" class="connection-name">
          {{ connectionName }}
        </span>
      </div>
      <div class="header-actions">
        <el-button
          :type="isConnected ? 'danger' : 'primary'"
          :icon="Connection"
          size="small"
          :loading="isConnecting"
          :disabled="!currentConnectionId"
          @click="toggleConnection"
        >
          {{ isConnected ? '断开' : '连接' }}
        </el-button>
      </div>
    </div>

    <!-- Error message -->
    <div v-if="lastError" class="error-banner">
      <el-alert
        :title="lastError"
        type="error"
        :closable="true"
        @close="pubsubStore.setError(null)"
      />
    </div>

    <!-- Main content -->
    <div class="panel-content">
      <!-- Left: Subscriptions and Publish -->
      <div class="left-section">
        <SubscriptionList
          :disabled="!isConnected"
          @subscribe="handleSubscribe"
          @unsubscribe="handleUnsubscribe"
        />
        <PublishForm
          :disabled="!isConnected"
          :connection-id="currentConnectionId"
        />
      </div>

      <!-- Right: Message History -->
      <div class="right-section">
        <MessageHistory @clear="handleClearMessages" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.pubsub-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: #fafafa;
  border-bottom: 1px solid #e4e7ed;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.panel-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.status-tag {
  font-size: 12px;
}

.connection-name {
  font-size: 12px;
  color: #909399;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.error-banner {
  padding: 8px 16px;
}

.panel-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.left-section {
  width: 320px;
  min-width: 280px;
  display: flex;
  flex-direction: column;
  border-right: 1px solid #e4e7ed;
}

.right-section {
  flex: 1;
  overflow: hidden;
}
</style>
