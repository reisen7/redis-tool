<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import type { ConnectionConfig } from '@/types/connection'
import { useConnectionStore } from '@/stores/connection'
import { useAuthStore } from '@/stores/auth'
import { useTabsStore } from '@/stores/tabs'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { User, Setting, SwitchButton, Key, ChatDotRound, ArrowDown } from '@element-plus/icons-vue'
import ConnectionSidebar from '@/components/ConnectionSidebar.vue'
import ConnectionDialog from '@/components/ConnectionDialog.vue'
import CreateKeyDialog from '@/components/CreateKeyDialog.vue'
import ThemePanel from '@/components/ThemePanel.vue'
import { TabBar, TabContent } from '@/components/workspace'
import { PubSubPanel } from '@/components/pubsub'
import TotpSetup from '@/components/auth/TotpSetup.vue'

const connectionStore = useConnectionStore()
const authStore = useAuthStore()
const tabsStore = useTabsStore()
const router = useRouter()

// 连接对话框
const showConnectionDialog = ref(false)
const editingConnection = ref<ConnectionConfig | null>(null)
const defaultGroupId = ref<string>('')

// 新增 Key 对话框
const showCreateKeyDialog = ref(false)

// 2FA 设置对话框
const show2FADialog = ref(false)

// 右侧面板模式: 'workspace' | 'pubsub'
const rightPanelMode = ref<'workspace' | 'pubsub'>('workspace')

// 是否有活动连接
const hasActiveConnection = computed(() => !!connectionStore.activeConnectionId)

// 当前用户信息
const currentUser = computed(() => authStore.user)
const isAdmin = computed(() => authStore.isAdmin)
const has2FAEnabled = computed(() => authStore.user?.totpEnabled ?? false)

// 打开新建连接
function handleNewConnection(groupId?: string) {
  editingConnection.value = null
  defaultGroupId.value = groupId || ''
  showConnectionDialog.value = true
}

// 打开编辑连接
function handleEditConnection(config: ConnectionConfig) {
  editingConnection.value = config
  defaultGroupId.value = ''
  showConnectionDialog.value = true
}

// 连接保存后
function handleConnectionSaved() {
  connectionStore.loadConnections()
}

// 打开新增 Key
function handleCreateKey() {
  showCreateKeyDialog.value = true
}

// 切换右侧面板模式
function switchPanelMode(mode: 'workspace' | 'pubsub') {
  if (!hasActiveConnection.value && mode !== 'workspace') {
    ElMessage.warning('请先连接 Redis')
    return
  }
  rightPanelMode.value = mode
}

// 登出
async function handleLogout() {
  try {
    await ElMessageBox.confirm(
      '确定要退出登录吗？',
      '退出确认',
      { type: 'warning' }
    )
    authStore.clearAuth()
    tabsStore.reset()
    router.push({ name: 'login' })
  } catch {
    // 用户取消
  }
}

// 跳转到管理面板
function goToAdmin() {
  router.push({ name: 'admin' })
}

// 打开 2FA 设置
function open2FASettings() {
  show2FADialog.value = true
}

// 2FA 设置完成
function handle2FASetupComplete() {
  show2FADialog.value = false
  ElMessage.success('双重认证已启用')
}

// 禁用 2FA
async function handleDisable2FA() {
  // 这个功能需要在 TotpSetup 组件中实现禁用流程
  // 或者创建一个单独的禁用对话框
  ElMessage.info('请在设置中禁用双重认证')
}

// 处理用户下拉菜单命令
function handleUserCommand(command: string) {
  switch (command) {
    case '2fa':
      open2FASettings()
      break
    case 'admin':
      goToAdmin()
      break
    case 'logout':
      handleLogout()
      break
  }
}

// 初始化
onMounted(() => {
  connectionStore.loadGroups()
  connectionStore.loadConnections()
  // Initialize tabs with current user ID for user-specific storage
  tabsStore.initialize(authStore.user?.id)
})
</script>

<template>
  <div class="app-container">
    <!-- 头部 -->
    <header class="app-header">
      <div class="logo">
        <svg class="logo-icon" viewBox="0 0 32 32" width="28" height="28">
          <defs>
            <linearGradient id="redis-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" style="stop-color:#DC382D"/>
              <stop offset="100%" style="stop-color:#A41E11"/>
            </linearGradient>
          </defs>
          <ellipse cx="16" cy="16" rx="14" ry="10" fill="url(#redis-gradient)"/>
          <ellipse cx="16" cy="12" rx="14" ry="10" fill="url(#redis-gradient)"/>
          <ellipse cx="16" cy="8" rx="14" ry="10" fill="#DC382D"/>
          <ellipse cx="16" cy="8" rx="10" ry="6" fill="#fff" opacity="0.2"/>
        </svg>
        <span class="logo-text">Redis Manager</span>
      </div>

      <!-- 面板切换按钮 -->
      <div class="panel-switcher">
        <el-button-group>
          <el-button
            :type="rightPanelMode === 'workspace' ? 'primary' : 'default'"
            @click="switchPanelMode('workspace')"
          >
            <el-icon><Key /></el-icon>
            工作区
          </el-button>
          <el-button
            :type="rightPanelMode === 'pubsub' ? 'primary' : 'default'"
            @click="switchPanelMode('pubsub')"
          >
            <el-icon><ChatDotRound /></el-icon>
            Pub/Sub
          </el-button>
        </el-button-group>
      </div>
      
      <!-- 用户信息和操作 -->
      <div class="header-actions">
        <!-- 用户下拉菜单 -->
        <el-dropdown trigger="click" @command="handleUserCommand">
          <span class="user-dropdown-trigger">
            <el-avatar :size="28" class="user-avatar">
              <el-icon><User /></el-icon>
            </el-avatar>
            <span class="user-name">{{ currentUser?.username }}</span>
            <el-tag v-if="isAdmin" size="small" type="warning" class="admin-tag">管理员</el-tag>
            <el-icon class="dropdown-arrow"><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item disabled>
                <div class="dropdown-user-info">
                  <span class="dropdown-username">{{ currentUser?.username }}</span>
                  <span class="dropdown-role">{{ isAdmin ? '管理员' : '普通用户' }}</span>
                </div>
              </el-dropdown-item>
              <el-dropdown-item divided command="2fa">
                <el-icon><Key /></el-icon>
                {{ has2FAEnabled ? '双重认证已启用' : '启用双重认证' }}
                <el-tag v-if="has2FAEnabled" size="small" type="success" class="ml-2">已启用</el-tag>
              </el-dropdown-item>
              <el-dropdown-item v-if="isAdmin" command="admin">
                <el-icon><Setting /></el-icon>
                管理面板
              </el-dropdown-item>
              <el-dropdown-item divided command="logout">
                <el-icon><SwitchButton /></el-icon>
                退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <!-- 主体 -->
    <main class="app-main">
      <!-- 左侧连接列表 -->
      <aside class="sidebar">
        <ConnectionSidebar
          @new-connection="handleNewConnection"
          @edit-connection="handleEditConnection"
        />
      </aside>

      <!-- 右侧工作区 -->
      <section class="workspace-area">
        <!-- 工作区模式 -->
        <template v-if="rightPanelMode === 'workspace'">
          <!-- Tab 栏 -->
          <TabBar v-if="hasActiveConnection" />
          
          <!-- Tab 内容 -->
          <TabContent />
        </template>

        <!-- Pub/Sub 模式 -->
        <template v-else>
          <PubSubPanel />
        </template>
      </section>
    </main>

    <!-- 连接对话框 -->
    <ConnectionDialog
      v-model:visible="showConnectionDialog"
      :edit-config="editingConnection"
      :default-group-id="defaultGroupId"
      @saved="handleConnectionSaved"
    />

    <!-- 新增 Key 对话框 -->
    <CreateKeyDialog v-model:visible="showCreateKeyDialog" />

    <!-- 2FA 设置对话框 -->
    <el-dialog
      v-model="show2FADialog"
      title="双重认证设置"
      width="600px"
      :close-on-click-modal="false"
    >
      <TotpSetup
        v-if="!has2FAEnabled"
        @setup-complete="handle2FASetupComplete"
        @cancel="show2FADialog = false"
      />
      <div v-else class="totp-enabled-info">
        <el-result
          icon="success"
          title="双重认证已启用"
          sub-title="您的账户已受到双重认证保护"
        >
          <template #extra>
            <el-button @click="show2FADialog = false">关闭</el-button>
          </template>
        </el-result>
      </div>
    </el-dialog>

    <!-- 主题配置面板 -->
    <ThemePanel />
  </div>
</template>



<style scoped>
.app-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f0f2f5;
}

.app-header {
  height: 50px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  flex-shrink: 0;
}

.logo {
  display: flex;
  align-items: center;
  gap: 8px;
}

.logo-icon {
  flex-shrink: 0;
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.panel-switcher {
  display: flex;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-dropdown-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.user-dropdown-trigger:hover {
  background: #f5f7fa;
}

.user-avatar {
  background: #409eff;
}

.user-name {
  font-size: 14px;
  color: #303133;
}

.admin-tag {
  margin-left: 4px;
}

.dropdown-arrow {
  font-size: 12px;
  color: #909399;
  transition: transform 0.2s;
}

.dropdown-user-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.dropdown-username {
  font-weight: 600;
  color: #303133;
}

.dropdown-role {
  font-size: 12px;
  color: #909399;
}

.ml-2 {
  margin-left: 8px;
}

.app-main {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.sidebar {
  width: 260px;
  flex-shrink: 0;
  background: #fff;
}

.workspace-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-left: 1px solid #e4e7ed;
  overflow: hidden;
}

.totp-enabled-info {
  padding: 20px;
}
</style>
