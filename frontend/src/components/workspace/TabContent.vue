<script setup lang="ts">
import { computed } from 'vue'
import { useTabsStore } from '@/stores/tabs'
import { useConnectionStore } from '@/stores/connection'
import WorkspacePanel from './WorkspacePanel.vue'
import PubSubPanel from '@/components/pubsub/PubSubPanel.vue'

const tabsStore = useTabsStore()
const connectionStore = useConnectionStore()

// Check if there's an active connection
const hasActiveConnection = computed(() => !!connectionStore.activeConnectionId)

// Check if there are any tabs for the active connection
const hasTabs = computed(() => {
  const connId = connectionStore.activeConnectionId
  if (!connId) return false
  return tabsStore.getTabsByConnection(connId).length > 0
})

// Get the active tab
const activeTab = computed(() => tabsStore.activeTab)

// Check if active tab is a pubsub tab
const isPubSubTab = computed(() => activeTab.value?.type === 'pubsub')

// Get connection info for welcome message
const connectionInfo = computed(() => {
  const conn = connectionStore.activeConnection
  if (!conn) return null
  return {
    name: conn.config.name,
    host: conn.config.host,
    port: conn.config.port,
    db: conn.currentDb
  }
})
</script>

<template>
  <div class="tab-content">
    <!-- No connection selected -->
    <div v-if="!hasActiveConnection" class="welcome-content">
      <div class="welcome-icon">
        <svg viewBox="0 0 64 64" width="80" height="80">
          <defs>
            <linearGradient id="redis-grad" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" style="stop-color:#DC382D"/>
              <stop offset="100%" style="stop-color:#A41E11"/>
            </linearGradient>
          </defs>
          <ellipse cx="32" cy="32" rx="28" ry="20" fill="url(#redis-grad)"/>
          <ellipse cx="32" cy="24" rx="28" ry="20" fill="url(#redis-grad)"/>
          <ellipse cx="32" cy="16" rx="28" ry="20" fill="#DC382D"/>
          <ellipse cx="32" cy="16" rx="20" ry="12" fill="#fff" opacity="0.2"/>
        </svg>
      </div>
      <h2 class="welcome-title">欢迎使用 Redis Manager</h2>
      <p class="welcome-desc">请从左侧选择一个连接开始使用</p>
    </div>

    <!-- Connection selected but no tabs -->
    <div v-else-if="!hasTabs" class="welcome-content">
      <div class="welcome-icon">
        <el-icon :size="64" color="#409eff">
          <svg viewBox="0 0 24 24" fill="currentColor">
            <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/>
          </svg>
        </el-icon>
      </div>
      <h2 class="welcome-title">{{ connectionInfo?.name }}</h2>
      <p class="welcome-desc">
        已连接到 {{ connectionInfo?.host }}:{{ connectionInfo?.port }} (db{{ connectionInfo?.db }})
      </p>
      <p class="welcome-hint">点击上方的 + 按钮创建新标签页</p>
    </div>

    <!-- Pub/Sub tab content -->
    <PubSubPanel v-else-if="activeTab && isPubSubTab" :connection-id="activeTab.connectionId" />

    <!-- Workspace tab content -->
    <WorkspacePanel v-else-if="activeTab" :tab="activeTab" />
  </div>
</template>

<style scoped>
.tab-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #fff;
}

.welcome-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  text-align: center;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4e7ed 100%);
}

.welcome-icon {
  margin-bottom: 24px;
  opacity: 0.8;
}

.welcome-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 12px 0;
}

.welcome-desc {
  font-size: 14px;
  color: #606266;
  margin: 0 0 8px 0;
}

.welcome-hint {
  font-size: 13px;
  color: #909399;
  margin: 16px 0 0 0;
}
</style>
