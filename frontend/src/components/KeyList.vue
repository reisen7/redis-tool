<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus, Delete } from '@element-plus/icons-vue'
import type { KeyInfo, SearchType } from '@/types/key'
import { getTypeIcon, getTypeColor } from '@/types/key'
import { useConnectionStore } from '@/stores/connection'
import { useKeysStore } from '@/stores/keys'
import { useTabsStore } from '@/stores/tabs'

const props = defineProps<{
  tabId?: string
}>()

const emit = defineEmits<{
  (e: 'create-key'): void
}>()

const connectionStore = useConnectionStore()
const keysStore = useKeysStore()
const tabsStore = useTabsStore()

// 搜索类型选项
const searchTypeOptions: { label: string; value: SearchType }[] = [
  { label: 'SCAN', value: 'scan' },
  { label: 'KEYS', value: 'keys' },
  { label: '正则', value: 'regex' }
]

// 选中的 Key
const selectedKeys = ref<string[]>([])

// Table ref for scroll position
const tableRef = ref<any>(null)

// Get current tab
const currentTab = computed(() => {
  if (props.tabId) {
    return tabsStore.getTab(props.tabId)
  }
  return tabsStore.activeTab
})

// Tab-specific search pattern (synced with tab state)
const tabPattern = computed({
  get: () => currentTab.value?.keysState.pattern ?? '*',
  set: (value: string) => {
    if (currentTab.value) {
      tabsStore.updateKeysState(currentTab.value.id, { pattern: value })
    }
  }
})

// Tab-specific selected key
const tabSelectedKey = computed({
  get: () => currentTab.value?.keysState.selectedKey ?? null,
  set: (value: string | null) => {
    if (currentTab.value) {
      tabsStore.updateKeysState(currentTab.value.id, { selectedKey: value })
    }
  }
})

// 当前连接和数据库
const currentConnectionId = computed(() => connectionStore.activeConnectionId)
const currentDb = computed(() => connectionStore.activeConnection?.currentDb ?? 0)

// Sync tab pattern with keysStore search config
watch(tabPattern, (newPattern) => {
  if (newPattern !== keysStore.searchConfig.pattern) {
    keysStore.updateSearchConfig({ pattern: newPattern })
  }
}, { immediate: true })

// 监听连接变化，重新加载 Key
watch([currentConnectionId, currentDb], ([connId, db]) => {
  if (connId) {
    // Use tab-specific pattern
    keysStore.updateSearchConfig({ pattern: tabPattern.value })
    keysStore.loadKeys(connId, db, true)
  }
}, { immediate: true })

// Watch for tab changes to restore state
watch(() => currentTab.value?.id, (newTabId, oldTabId) => {
  if (newTabId && newTabId !== oldTabId) {
    // Restore tab-specific state
    const tab = currentTab.value
    if (tab) {
      // Update search config with tab's pattern
      keysStore.updateSearchConfig({ pattern: tab.keysState.pattern })
      
      // Restore selected key
      if (tab.keysState.selectedKey) {
        keysStore.selectedKey = tab.keysState.selectedKey
      }
      
      // Restore scroll position after DOM update
      nextTick(() => {
        restoreScrollPosition(tab.keysState.scrollPosition)
      })
    }
  }
}, { immediate: true })

// Save scroll position when it changes
function saveScrollPosition() {
  if (currentTab.value && tableRef.value) {
    const scrollWrapper = tableRef.value.$el?.querySelector('.el-table__body-wrapper')
    if (scrollWrapper) {
      tabsStore.updateKeysState(currentTab.value.id, { 
        scrollPosition: scrollWrapper.scrollTop 
      })
    }
  }
}

// Restore scroll position
function restoreScrollPosition(position: number) {
  if (tableRef.value) {
    const scrollWrapper = tableRef.value.$el?.querySelector('.el-table__body-wrapper')
    if (scrollWrapper) {
      scrollWrapper.scrollTop = position
    }
  }
}

// 搜索
function handleSearch() {
  if (!currentConnectionId.value) return
  
  // Save pattern to tab state
  if (currentTab.value) {
    tabsStore.updateKeysState(currentTab.value.id, { 
      pattern: keysStore.searchConfig.pattern,
      cursor: 0 
    })
  }
  
  if (keysStore.searchConfig.searchType === 'regex') {
    keysStore.searchKeys(currentConnectionId.value, currentDb.value)
  } else {
    keysStore.loadKeys(currentConnectionId.value, currentDb.value, true)
  }
}

// 刷新
function handleRefresh() {
  if (!currentConnectionId.value) return
  keysStore.loadKeys(currentConnectionId.value, currentDb.value, true)
}

// 加载更多
function handleLoadMore() {
  if (!currentConnectionId.value) return
  keysStore.loadMore(currentConnectionId.value, currentDb.value)
}

// 选择 Key
function handleSelectKey(key: KeyInfo) {
  if (!currentConnectionId.value) return
  
  // Update tab-specific selected key
  tabSelectedKey.value = key.key
  
  keysStore.getKeyValue(currentConnectionId.value, currentDb.value, key.key)
}

// 双击 Key
function handleDoubleClick(key: KeyInfo) {
  handleSelectKey(key)
}

// 删除选中的 Key
async function handleDeleteSelected() {
  if (selectedKeys.value.length === 0) {
    ElMessage.warning('请先选择要删除的 Key')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedKeys.value.length} 个 Key 吗？`,
      '删除确认',
      { type: 'warning' }
    )
    
    if (!currentConnectionId.value) return
    
    const result = await keysStore.deleteKeys(
      currentConnectionId.value,
      currentDb.value,
      selectedKeys.value
    )
    
    if (result.success) {
      ElMessage.success(`成功删除 ${result.deleted} 个 Key`)
      selectedKeys.value = []
      
      // Clear tab selected key if it was deleted
      if (tabSelectedKey.value && selectedKeys.value.includes(tabSelectedKey.value)) {
        tabSelectedKey.value = null
      }
    } else {
      ElMessage.error(result.error || '删除失败')
    }
  } catch {}
}

// 表格选择变化
function handleSelectionChange(selection: KeyInfo[]) {
  selectedKeys.value = selection.map(k => k.key)
}

// Handle table scroll
function handleTableScroll() {
  saveScrollPosition()
}

// Setup scroll listener on mount
onMounted(() => {
  nextTick(() => {
    if (tableRef.value) {
      const scrollWrapper = tableRef.value.$el?.querySelector('.el-table__body-wrapper')
      if (scrollWrapper) {
        scrollWrapper.addEventListener('scroll', handleTableScroll)
      }
    }
  })
})
</script>

<template>
  <div class="key-list">
    <!-- 搜索栏 -->
    <div class="search-bar">
      <el-select
        v-model="keysStore.searchConfig.searchType"
        style="width: 100px"
        size="small"
      >
        <el-option
          v-for="opt in searchTypeOptions"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
      
      <el-input
        v-if="keysStore.searchConfig.searchType !== 'regex'"
        v-model="keysStore.searchConfig.pattern"
        placeholder="搜索模式，如 user:*"
        size="small"
        clearable
        @keyup.enter="handleSearch"
      />
      <el-input
        v-else
        v-model="keysStore.searchConfig.regexPattern"
        placeholder="正则表达式"
        size="small"
        clearable
        @keyup.enter="handleSearch"
      />
      
      <el-button
        type="primary"
        :icon="Search"
        size="small"
        :loading="keysStore.loading"
        @click="handleSearch"
      >
        搜索
      </el-button>
    </div>

    <!-- 工具栏 -->
    <div class="toolbar">
      <el-button
        type="primary"
        :icon="Plus"
        size="small"
        @click="emit('create-key')"
      >
        新增
      </el-button>
      <el-button
        type="danger"
        :icon="Delete"
        size="small"
        :disabled="selectedKeys.length === 0"
        @click="handleDeleteSelected"
      >
        删除 ({{ selectedKeys.length }})
      </el-button>
      <el-button
        :icon="Refresh"
        size="small"
        @click="handleRefresh"
      >
        刷新
      </el-button>
      
      <span class="key-count">
        共 {{ keysStore.keys.length }} 个 Key
      </span>
    </div>

    <!-- Key 列表 -->
    <el-table
      ref="tableRef"
      :data="keysStore.keys"
      v-loading="keysStore.loading"
      height="100%"
      highlight-current-row
      @selection-change="handleSelectionChange"
      @row-dblclick="handleDoubleClick"
      @row-click="handleSelectKey"
      :row-class-name="({ row }: { row: KeyInfo }) => row.key === tabSelectedKey ? 'selected-row' : ''"
    >
      <el-table-column type="selection" width="40" />
      
      <el-table-column label="Key" min-width="200">
        <template #default="{ row }: { row: KeyInfo }">
          <div class="key-cell">
            <span class="type-icon">{{ getTypeIcon(row.type) }}</span>
            <span class="key-name">{{ row.key }}</span>
          </div>
        </template>
      </el-table-column>
      
      <el-table-column label="类型" width="100">
        <template #default="{ row }">
          <el-tag
            size="small"
            :color="getTypeColor(row.type)"
            effect="dark"
          >
            {{ row.type.toUpperCase() }}
          </el-tag>
        </template>
      </el-table-column>
    </el-table>

    <!-- 加载更多 -->
    <div v-if="keysStore.hasMore" class="load-more">
      <el-button
        type="primary"
        link
        :loading="keysStore.loading"
        @click="handleLoadMore"
      >
        加载更多...
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.key-list {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.search-bar {
  display: flex;
  gap: 8px;
  padding: 12px;
  border-bottom: 1px solid #e4e7ed;
}

.search-bar .el-input {
  flex: 1;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-bottom: 1px solid #e4e7ed;
}

.key-count {
  margin-left: auto;
  color: #909399;
  font-size: 12px;
}

.key-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.type-icon {
  font-size: 14px;
}

.key-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ttl-warning {
  color: #f56c6c;
}

.load-more {
  padding: 12px;
  text-align: center;
  border-top: 1px solid #e4e7ed;
}

:deep(.selected-row) {
  background-color: #ecf5ff !important;
}

:deep(.el-table) {
  flex: 1;
}
</style>
