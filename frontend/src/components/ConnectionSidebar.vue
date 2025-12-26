<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Connection,
  Delete,
  Edit,
  FolderOpened,
  Folder,
  Monitor,
  Grid,
  Aim,
  Loading,
  FolderAdd
} from '@element-plus/icons-vue'
import type { ConnectionConfig, ConnectionGroup, DatabaseInfo, ConnectionMode } from '@/types/connection'
import { createDefaultGroup } from '@/types/connection'
import { useConnectionStore } from '@/stores/connection'
import { useKeysStore } from '@/stores/keys'
import { useTabsStore } from '@/stores/tabs'

const emit = defineEmits<{
  (e: 'new-connection', groupId?: string): void
  (e: 'edit-connection', config: ConnectionConfig): void
}>()

const connectionStore = useConnectionStore()
const keysStore = useKeysStore()
const tabsStore = useTabsStore()

// 展开的连接和分组
const expandedConnections = ref<Set<string>>(new Set())
const expandedGroups = ref<Set<string>>(new Set(['', ...connectionStore.groups.map(g => g.id)]))

// 分组对话框
const showGroupDialog = ref(false)
const editingGroup = ref<ConnectionGroup>(createDefaultGroup())

// 分组颜色选项
const groupColors = [
  '#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399', '#9c27b0'
]

onMounted(() => {
  connectionStore.loadGroups()
})

// 切换分组展开状态
function toggleGroup(groupId: string) {
  if (expandedGroups.value.has(groupId)) {
    expandedGroups.value.delete(groupId)
  } else {
    expandedGroups.value.add(groupId)
  }
}

// 切换连接展开状态
function toggleConnection(id: string) {
  if (expandedConnections.value.has(id)) {
    expandedConnections.value.delete(id)
  } else {
    expandedConnections.value.add(id)
  }
}

// 双击连接 - 自动连接并展开
async function handleDoubleClick(config: ConnectionConfig) {
  const id = config.id
  
  if (isConnected(id)) {
    // 已连接，检查是否有 Tab，没有就创建，有就激活
    const existingTabs = tabsStore.getTabsByConnection(id)
    if (existingTabs.length > 0) {
      tabsStore.setActiveTab(existingTabs[0].id)
    } else {
      tabsStore.addTab(id, config.name)
    }
    toggleConnection(id)
  } else {
    const result = await connectionStore.openConnection(config)
    if (result.success) {
      expandedConnections.value.add(id)
      // 新连接，总是创建新 Tab
      tabsStore.addTab(id, config.name)
      ElMessage.success('连接成功')
    } else {
      ElMessage.error(result.error || '连接失败')
    }
  }
}

// 打开连接
async function handleOpenConnection(config: ConnectionConfig) {
  console.log('handleOpenConnection called with config:', config)
  
  const result = await connectionStore.openConnection(config)
  console.log('openConnection result:', result)
  
  if (result.success) {
    expandedConnections.value.add(config.id)
    // 新连接，总是创建新 Tab
    tabsStore.addTab(config.id, config.name)
    console.log('Created new tab for', config.id, config.name)
    console.log('all tabs:', tabsStore.tabs)
    ElMessage.success('连接成功')
  } else {
    ElMessage.error(result.error || '连接失败')
  }
}

// 关闭连接
async function handleCloseConnection(id: string) {
  await connectionStore.closeConnection(id)
  expandedConnections.value.delete(id)
  // 关闭该连接的所有 Tab
  tabsStore.removeTabsByConnection(id)
  keysStore.clearSelection()
}

// 删除连接
async function handleDeleteConnection(config: ConnectionConfig) {
  try {
    await ElMessageBox.confirm(
      `确定要删除连接 "${config.name}" 吗？`,
      '删除确认',
      { type: 'warning' }
    )
    const result = await connectionStore.deleteConnection(config.id)
    if (result.success) {
      ElMessage.success('删除成功')
    } else {
      ElMessage.error(result.error || '删除失败')
    }
  } catch {}
}

// 选择数据库
async function handleSelectDatabase(connectionId: string, db: DatabaseInfo) {
  const result = await connectionStore.selectDatabase(connectionId, db.index)
  if (result.success) {
    connectionStore.setActiveConnection(connectionId)
    keysStore.loadKeys(connectionId, db.index, true)
  } else {
    ElMessage.error(result.error || '切换数据库失败')
  }
}

// --- 分组操作 ---

// 打开新建分组对话框
function handleNewGroup() {
  editingGroup.value = createDefaultGroup()
  showGroupDialog.value = true
}

// 编辑分组
function handleEditGroup(group: ConnectionGroup) {
  editingGroup.value = { ...group }
  showGroupDialog.value = true
}

// 保存分组
async function handleSaveGroup() {
  if (!editingGroup.value.name.trim()) {
    ElMessage.warning('请输入分组名称')
    return
  }
  
  const result = await connectionStore.saveGroup(editingGroup.value)
  if (result.success) {
    ElMessage.success('保存成功')
    showGroupDialog.value = false
    // 展开新分组
    if (result.data) {
      expandedGroups.value.add(result.data.id)
    }
  } else {
    ElMessage.error(result.error || '保存失败')
  }
}

// 删除分组
async function handleDeleteGroup(group: ConnectionGroup) {
  try {
    await ElMessageBox.confirm(
      `确定要删除分组 "${group.name}" 吗？分组内的连接将移至未分组。`,
      '删除确认',
      { type: 'warning' }
    )
    const result = await connectionStore.deleteGroup(group.id)
    if (result.success) {
      ElMessage.success('删除成功')
    } else {
      ElMessage.error(result.error || '删除失败')
    }
  } catch {}
}

// 在分组内新建连接
function handleNewConnectionInGroup(groupId: string) {
  emit('new-connection', groupId)
}

// 获取连接模式图标
function getModeIcon(mode: ConnectionMode) {
  switch (mode) {
    case 'standalone': return Monitor
    case 'cluster': return Grid
    case 'sentinel': return Aim
    default: return Monitor
  }
}

// 获取连接模式颜色
function getModeColor(mode: ConnectionMode): string {
  switch (mode) {
    case 'standalone': return '#67c23a'
    case 'cluster': return '#409eff'
    case 'sentinel': return '#e6a23c'
    default: return '#67c23a'
  }
}

// 获取连接状态图标颜色
function getStatusColor(id: string): string {
  const instance = connectionStore.openConnections.get(id)
  if (!instance) return '#909399'
  
  switch (instance.status) {
    case 'connected': return '#67c23a'
    case 'connecting': return '#e6a23c'
    case 'error': return '#f56c6c'
    default: return '#909399'
  }
}

function isConnecting(id: string): boolean {
  const instance = connectionStore.openConnections.get(id)
  return instance?.status === 'connecting'
}

function isConnected(id: string): boolean {
  const instance = connectionStore.openConnections.get(id)
  return instance?.status === 'connected'
}

function getCurrentDb(id: string): number {
  const instance = connectionStore.openConnections.get(id)
  return instance?.currentDb ?? 0
}

function getDatabases(id: string): DatabaseInfo[] {
  const instance = connectionStore.openConnections.get(id)
  return instance?.databases || []
}

// 获取分组内的连接
function getConnectionsInGroup(groupId: string): ConnectionConfig[] {
  return connectionStore.connectionsByGroup.get(groupId) || []
}
</script>

<template>
  <div class="connection-sidebar">
    <div class="sidebar-header">
      <span class="title">连接列表</span>
      <div class="header-actions">
        <el-button
          type="primary"
          :icon="FolderAdd"
          size="small"
          circle
          title="新建分组"
          @click="handleNewGroup"
        />
        <el-button
          type="primary"
          :icon="Plus"
          size="small"
          @click="emit('new-connection')"
        >
          新建
        </el-button>
      </div>
    </div>

    <div class="connection-list">
      <!-- 分组列表 -->
      <template v-for="group in connectionStore.groups" :key="group.id">
        <div class="group-item">
          <div
            class="group-header"
            @click="toggleGroup(group.id)"
          >
            <el-icon class="expand-icon">
              <FolderOpened v-if="expandedGroups.has(group.id)" />
              <Folder v-else />
            </el-icon>
            <span
              class="group-color-dot"
              :style="{ backgroundColor: group.color }"
            />
            <span class="group-name">{{ group.name }}</span>
            <span class="group-count">({{ getConnectionsInGroup(group.id).length }})</span>
            
            <div class="group-actions" @click.stop>
              <el-button
                type="primary"
                link
                size="small"
                title="在此分组新建连接"
                @click="handleNewConnectionInGroup(group.id)"
              >
                <el-icon><Plus /></el-icon>
              </el-button>
              <el-button
                type="primary"
                link
                size="small"
                @click="handleEditGroup(group)"
              >
                <el-icon><Edit /></el-icon>
              </el-button>
              <el-button
                type="danger"
                link
                size="small"
                @click="handleDeleteGroup(group)"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </div>

          <!-- 分组内的连接 -->
          <div v-if="expandedGroups.has(group.id)" class="group-connections">
            <div
              v-for="conn in getConnectionsInGroup(group.id)"
              :key="conn.id"
              class="connection-item"
            >
              <div
                class="connection-header"
                :class="{ active: connectionStore.activeConnectionId === conn.id }"
  
                @dblclick="handleDoubleClick(conn)"
              >
                <el-icon class="mode-icon" :style="{ color: getModeColor(conn.mode) }">
                  <component :is="getModeIcon(conn.mode)" />
                </el-icon>
                
                <span v-if="isConnecting(conn.id)" class="status-loading">
                  <el-icon class="is-loading"><Loading /></el-icon>
                </span>
                <span v-else class="status-dot" :style="{ backgroundColor: getStatusColor(conn.id) }" />
                
                <span class="connection-name">{{ conn.name }}</span>
                
                <div class="connection-actions" @click.stop>
                  <el-button
                    v-if="!isConnected(conn.id)"
                    type="primary"
                    link
                    size="small"
                    :loading="isConnecting(conn.id)"
                    @click="handleOpenConnection(conn)"
                  >
                    <el-icon v-if="!isConnecting(conn.id)"><Connection /></el-icon>
                  </el-button>
                  <el-button
                    v-else
                    type="danger"
                    link
                    size="small"
                    @click="handleCloseConnection(conn.id)"
                  >
                    断开
                  </el-button>
                  <el-button type="primary" link size="small" @click="emit('edit-connection', conn)">
                    <el-icon><Edit /></el-icon>
                  </el-button>
                  <el-button type="danger" link size="small" @click="handleDeleteConnection(conn)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </div>

              <!-- 数据库列表 -->
              <div
                v-if="expandedConnections.has(conn.id) && isConnected(conn.id)"
                class="database-list"
              >
                <div
                  v-for="db in getDatabases(conn.id)"
                  :key="db.index"
                  class="database-item"
                  :class="{ active: getCurrentDb(conn.id) === db.index && connectionStore.activeConnectionId === conn.id }"
                  @click="handleSelectDatabase(conn.id, db)"
                >
                  <span class="db-name">
                    db{{ db.index }}
                    <span v-if="db.alias" class="db-alias">({{ db.alias }})</span>
                  </span>
                  <span class="db-count">{{ db.keyCount }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- 未分组的连接 -->
      <div v-if="getConnectionsInGroup('').length > 0" class="group-item">
        <div class="group-header" @click="toggleGroup('')">
          <el-icon class="expand-icon">
            <FolderOpened v-if="expandedGroups.has('')" />
            <Folder v-else />
          </el-icon>
          <span class="group-name ungrouped">未分组</span>
          <span class="group-count">({{ getConnectionsInGroup('').length }})</span>
        </div>

        <div v-if="expandedGroups.has('')" class="group-connections">
          <div
            v-for="conn in getConnectionsInGroup('')"
            :key="conn.id"
            class="connection-item"
          >
            <div
              class="connection-header"
              :class="{ active: connectionStore.activeConnectionId === conn.id }"
              @dblclick="handleDoubleClick(conn)"
            >
              <el-icon class="mode-icon" :style="{ color: getModeColor(conn.mode) }">
                <component :is="getModeIcon(conn.mode)" />
              </el-icon>
              
              <span v-if="isConnecting(conn.id)" class="status-loading">
                <el-icon class="is-loading"><Loading /></el-icon>
              </span>
              <span v-else class="status-dot" :style="{ backgroundColor: getStatusColor(conn.id) }" />
              
              <span class="connection-name">{{ conn.name }}</span>
              
              <div class="connection-actions" @click.stop>
                <el-button
                  v-if="!isConnected(conn.id)"
                  type="primary"
                  link
                  size="small"
                  :loading="isConnecting(conn.id)"
                  @click="handleOpenConnection(conn)"
                >
                  <el-icon v-if="!isConnecting(conn.id)"><Connection /></el-icon>
                </el-button>
                <el-button
                  v-else
                  type="danger"
                  link
                  size="small"
                  @click="handleCloseConnection(conn.id)"
                >
                  断开
                </el-button>
                <el-button type="primary" link size="small" @click="emit('edit-connection', conn)">
                  <el-icon><Edit /></el-icon>
                </el-button>
                <el-button type="danger" link size="small" @click="handleDeleteConnection(conn)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>

            <div
              v-if="expandedConnections.has(conn.id) && isConnected(conn.id)"
              class="database-list"
            >
              <div
                v-for="db in getDatabases(conn.id)"
                :key="db.index"
                class="database-item"
                :class="{ active: getCurrentDb(conn.id) === db.index && connectionStore.activeConnectionId === conn.id }"
                @click="handleSelectDatabase(conn.id, db)"
              >
                <span class="db-name">
                  db{{ db.index }}
                  <span v-if="db.alias" class="db-alias">({{ db.alias }})</span>
                </span>
                <span class="db-count">{{ db.keyCount }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <el-empty
        v-if="connectionStore.connections.length === 0"
        description="暂无连接"
        :image-size="80"
      />
    </div>

    <!-- 分组对话框 -->
    <el-dialog
      v-model="showGroupDialog"
      :title="editingGroup.id ? '编辑分组' : '新建分组'"
      width="400px"
    >
      <el-form :model="editingGroup" label-width="80px">
        <el-form-item label="分组名称" required>
          <el-input v-model="editingGroup.name" placeholder="请输入分组名称" />
        </el-form-item>
        <el-form-item label="分组颜色">
          <div class="color-picker">
            <div
              v-for="color in groupColors"
              :key="color"
              class="color-option"
              :class="{ active: editingGroup.color === color }"
              :style="{ backgroundColor: color }"
              @click="editingGroup.color = color"
            />
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showGroupDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveGroup">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.connection-sidebar {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-color-light, #fff);
  border-right: 1px solid var(--border-color, #e4e7ed);
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color, #e4e7ed);
  background: var(--bg-color-light, #fff);
}

.title {
  font-weight: 600;
  font-size: 14px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.connection-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.group-item {
  margin-bottom: 4px;
}

.group-header {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
  user-select: none;
}

.group-header:hover {
  background: var(--bg-color-dark, #e9ecef);
}

.expand-icon {
  margin-right: 6px;
  color: #909399;
}

.group-color-dot {
  width: 10px;
  height: 10px;
  border-radius: 2px;
  margin-right: 8px;
}

.group-name {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
}

.group-name.ungrouped {
  color: #909399;
}

.group-count {
  color: #909399;
  font-size: 12px;
  margin-right: 8px;
}

.group-actions {
  display: flex;
  gap: 2px;
  opacity: 0;
  transition: opacity 0.2s;
}

.group-header:hover .group-actions {
  opacity: 1;
}

.group-connections {
  margin-left: 16px;
}

.connection-item {
  margin-bottom: 2px;
}

.connection-header {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
  user-select: none;
}

.connection-header:hover {
  background: var(--bg-color-dark, #e9ecef);
}

.connection-header.active {
  background: rgba(64, 158, 255, 0.1);
}

.mode-icon {
  margin-right: 6px;
  font-size: 14px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 8px;
  flex-shrink: 0;
}

.status-loading {
  margin-right: 8px;
  display: flex;
  align-items: center;
}

.status-loading .el-icon {
  font-size: 14px;
  color: #e6a23c;
}

.connection-name {
  flex: 1;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.connection-actions {
  display: flex;
  gap: 2px;
  opacity: 0;
  transition: opacity 0.2s;
}

.connection-header:hover .connection-actions {
  opacity: 1;
}

.database-list {
  margin-left: 24px;
  padding: 4px 0;
}

.database-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  color: var(--text-color-regular, #606266);
}

.database-item:hover {
  background: var(--bg-color-dark, #e9ecef);
}

.database-item.active {
  background: rgba(64, 158, 255, 0.1);
  color: #409eff;
}

.db-alias {
  color: #909399;
  margin-left: 4px;
}

.db-count {
  color: #909399;
  font-size: 11px;
}

.color-picker {
  display: flex;
  gap: 8px;
}

.color-option {
  width: 24px;
  height: 24px;
  border-radius: 4px;
  cursor: pointer;
  border: 2px solid transparent;
  transition: all 0.2s;
}

.color-option:hover {
  transform: scale(1.1);
}

.color-option.active {
  border-color: #303133;
}
</style>
