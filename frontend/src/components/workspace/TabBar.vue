<script setup lang="ts">
import { computed } from 'vue'
import { Plus, Close, Monitor, ChatDotRound } from '@element-plus/icons-vue'
import { useTabsStore } from '@/stores/tabs'
import { useConnectionStore } from '@/stores/connection'
import type { Tab, TabType } from '@/types/tab'

const tabsStore = useTabsStore()
const connectionStore = useConnectionStore()

// Get all tabs (not filtered by connection)
const allTabs = computed(() => {
  return tabsStore.tabs
})

// Check if a tab is active
function isActiveTab(tabId: string): boolean {
  return tabsStore.activeTabId === tabId
}

// Switch to a tab and activate its connection
function switchTab(tab: Tab) {
  tabsStore.setActiveTab(tab.id)
  // Also activate the connection for this tab
  connectionStore.setActiveConnection(tab.connectionId)
}

// Create a new workspace tab for active connection
function createNewTab() {
  const connId = connectionStore.activeConnectionId
  if (!connId) return
  
  const connName = connectionStore.activeConnection?.config.name || 'Redis'
  const workspaceTabs = allTabs.value.filter(t => t.connectionId === connId && t.type === 'workspace')
  const tabCount = workspaceTabs.length + 1
  tabsStore.addTab(connId, `${connName} - Tab ${tabCount}`, 'workspace')
}

// Create a new Pub/Sub tab for active connection
function createPubSubTab() {
  const connId = connectionStore.activeConnectionId
  if (!connId) return
  
  const connName = connectionStore.activeConnection?.config.name || 'Redis'
  const pubsubTabs = allTabs.value.filter(t => t.connectionId === connId && t.type === 'pubsub')
  const tabCount = pubsubTabs.length + 1
  tabsStore.addTab(connId, `${connName} - Pub/Sub ${tabCount}`, 'pubsub')
}

// Close a tab
function closeTab(tabId: string, event: MouseEvent) {
  event.stopPropagation()
  tabsStore.removeTab(tabId)
}

// Get truncated title for display
function getTruncatedTitle(title: string, maxLength: number = 20): string {
  if (title.length <= maxLength) return title
  return title.substring(0, maxLength - 3) + '...'
}

// Get icon for tab type
function getTabIcon(type: TabType) {
  return type === 'pubsub' ? ChatDotRound : Monitor
}
</script>

<template>
  <div class="tab-bar">
    <div class="tabs-container">
      <!-- Tab list - show all tabs -->
      <div
        v-for="tab in allTabs"
        :key="tab.id"
        class="tab-item"
        :class="{ active: isActiveTab(tab.id), pubsub: tab.type === 'pubsub' }"
        @click="switchTab(tab)"
        :title="tab.title"
      >
        <el-icon class="tab-icon">
          <component :is="getTabIcon(tab.type)" />
        </el-icon>
        <span class="tab-title">{{ getTruncatedTitle(tab.title) }}</span>
        <el-icon
          class="tab-close"
          @click="closeTab(tab.id, $event)"
        >
          <Close />
        </el-icon>
      </div>
      
      <!-- New tab button -->
      <div
        class="new-tab-btn"
        @click="createNewTab"
        title="新建工作区标签页"
      >
        <el-icon><Plus /></el-icon>
      </div>
      
      <!-- New Pub/Sub tab button -->
      <div
        class="new-tab-btn pubsub-btn"
        @click="createPubSubTab"
        title="新建 Pub/Sub 标签页"
      >
        <el-icon><ChatDotRound /></el-icon>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tab-bar {
  display: flex;
  align-items: center;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
  height: 36px;
  padding: 0 8px;
}

.tabs-container {
  display: flex;
  align-items: center;
  gap: 2px;
  overflow-x: auto;
  flex: 1;
}

.tabs-container::-webkit-scrollbar {
  height: 4px;
}

.tabs-container::-webkit-scrollbar-thumb {
  background: #c0c4cc;
  border-radius: 2px;
}

.tab-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-bottom: none;
  border-radius: 4px 4px 0 0;
  cursor: pointer;
  font-size: 13px;
  color: #606266;
  white-space: nowrap;
  transition: all 0.2s;
  position: relative;
  top: 1px;
}

.tab-item:hover {
  background: #ecf5ff;
  color: #409eff;
}

.tab-item.active {
  background: #fff;
  color: #409eff;
  border-color: #e4e7ed;
  border-bottom: 1px solid #fff;
  font-weight: 500;
}

.tab-item.pubsub {
  border-top: 2px solid #67c23a;
}

.tab-item.pubsub.active {
  color: #67c23a;
}

.tab-icon {
  font-size: 14px;
}

.tab-title {
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tab-close {
  font-size: 12px;
  color: #909399;
  border-radius: 50%;
  padding: 2px;
  transition: all 0.2s;
}

.tab-close:hover {
  background: #f56c6c;
  color: #fff;
}

.new-tab-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 4px;
  cursor: pointer;
  color: #909399;
  transition: all 0.2s;
  flex-shrink: 0;
}

.new-tab-btn:hover {
  background: #ecf5ff;
  color: #409eff;
}

.new-tab-btn.pubsub-btn:hover {
  background: #f0f9eb;
  color: #67c23a;
}
</style>
