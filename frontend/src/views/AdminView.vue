<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ArrowLeft, User, SwitchButton, UserFilled, Setting } from '@element-plus/icons-vue'
import UserList from '@/components/admin/UserList.vue'
import SettingsPanel from '@/components/admin/SettingsPanel.vue'

const router = useRouter()
const authStore = useAuthStore()

// Current active section
const activeSection = ref<'users' | 'settings'>('users')

function goBack() {
  router.push({ name: 'app' })
}

function handleLogout() {
  authStore.clearAuth()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="admin-container">
    <!-- Header -->
    <header class="admin-header">
      <div class="header-left">
        <el-button text @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          返回应用
        </el-button>
        <h1 class="admin-title">管理面板</h1>
      </div>
      
      <div class="header-right">
        <span class="user-info">
          <el-icon><User /></el-icon>
          {{ authStore.user?.username }}
        </span>
        <el-button text @click="handleLogout">
          <el-icon><SwitchButton /></el-icon>
          登出
        </el-button>
      </div>
    </header>

    <!-- Main content -->
    <div class="admin-body">
      <!-- Sidebar navigation -->
      <aside class="admin-sidebar">
        <el-menu
          :default-active="activeSection"
          @select="(key: string) => activeSection = key as 'users' | 'settings'"
        >
          <el-menu-item index="users">
            <el-icon><UserFilled /></el-icon>
            <span>用户管理</span>
          </el-menu-item>
          <el-menu-item index="settings">
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </el-menu-item>
        </el-menu>
      </aside>

      <!-- Content area -->
      <main class="admin-main">
        <UserList v-if="activeSection === 'users'" />
        <SettingsPanel v-else-if="activeSection === 'settings'" />
      </main>
    </div>
  </div>
</template>

<style scoped>
.admin-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f0f2f5;
}

.admin-header {
  height: 50px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.admin-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: #606266;
}

.admin-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.admin-sidebar {
  width: 200px;
  background: #fff;
  border-right: 1px solid #e4e7ed;
  flex-shrink: 0;
}

.admin-sidebar .el-menu {
  border-right: none;
  height: 100%;
}

.admin-main {
  flex: 1;
  padding: 20px;
  overflow: auto;
}
</style>
