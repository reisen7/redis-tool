<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Edit, Delete, Key, Check, Close } from '@element-plus/icons-vue'
import adminApi from '@/api/admin'
import type { User, UserRole } from '@/types/auth'
import UserForm from './UserForm.vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

// State
const loading = ref(false)
const users = ref<User[]>([])
const showUserForm = ref(false)
const editingUser = ref<User | null>(null)

// Load users on mount
onMounted(() => {
  loadUsers()
})

// Load all users
async function loadUsers() {
  loading.value = true
  try {
    const response = await adminApi.listUsers()
    if (response.data.success && response.data.data) {
      users.value = response.data.data
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '加载用户列表失败')
  } finally {
    loading.value = false
  }
}

// Open create user dialog
function handleCreate() {
  editingUser.value = null
  showUserForm.value = true
}

// Open edit user dialog
function handleEdit(user: User) {
  editingUser.value = user
  showUserForm.value = true
}

// Handle form submit success
function handleFormSuccess() {
  showUserForm.value = false
  editingUser.value = null
  loadUsers()
}

// Toggle user enabled status
async function handleToggleEnabled(user: User) {
  const action = user.enabled ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(
      `确定要${action}用户 "${user.username}" 吗？`,
      '确认操作',
      { type: 'warning' }
    )
    
    const response = await adminApi.updateUser(user.id, { enabled: !user.enabled })
    if (response.data.success) {
      ElMessage.success(`用户已${action}`)
      loadUsers()
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || `${action}用户失败`)
    }
  }
}

// Reset user password
async function handleResetPassword(user: User) {
  try {
    await ElMessageBox.confirm(
      `确定要重置用户 "${user.username}" 的密码吗？`,
      '确认操作',
      { type: 'warning' }
    )
    
    const response = await adminApi.resetPassword(user.id)
    if (response.data.success && response.data.data) {
      ElMessageBox.alert(
        `临时密码: ${response.data.data.tempPassword}\n\n请将此密码告知用户，用户下次登录时需要修改密码。`,
        '密码已重置',
        { type: 'success' }
      )
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '重置密码失败')
    }
  }
}

// Delete user
async function handleDelete(user: User) {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${user.username}" 吗？此操作将删除该用户的所有数据，且不可恢复。`,
      '确认删除',
      { type: 'error' }
    )
    
    const response = await adminApi.deleteUser(user.id)
    if (response.data.success) {
      ElMessage.success('用户已删除')
      loadUsers()
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除用户失败')
    }
  }
}

// Get role display text
function getRoleText(role: UserRole): string {
  return role === 'admin' ? '管理员' : '普通用户'
}

// Get role tag type
function getRoleType(role: UserRole): 'danger' | 'info' {
  return role === 'admin' ? 'danger' : 'info'
}

// Check if user is current user
function isCurrentUser(user: User): boolean {
  return user.id === authStore.user?.id
}
</script>

<template>
  <div class="user-list">
    <!-- Header -->
    <div class="list-header">
      <h2>用户管理</h2>
      <el-button type="primary" :icon="Plus" @click="handleCreate">
        新建用户
      </el-button>
    </div>

    <!-- User table -->
    <el-table
      v-loading="loading"
      :data="users"
      stripe
      style="width: 100%"
    >
      <el-table-column prop="id" label="ID" width="80" />
      
      <el-table-column prop="username" label="用户名" min-width="120">
        <template #default="{ row }">
          <span>{{ row.username }}</span>
          <el-tag v-if="isCurrentUser(row)" size="small" type="success" class="current-tag">
            当前
          </el-tag>
        </template>
      </el-table-column>
      
      <el-table-column prop="role" label="角色" width="100">
        <template #default="{ row }">
          <el-tag :type="getRoleType(row.role)" size="small">
            {{ getRoleText(row.role) }}
          </el-tag>
        </template>
      </el-table-column>
      
      <el-table-column prop="enabled" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.enabled ? 'success' : 'danger'" size="small">
            <el-icon v-if="row.enabled"><Check /></el-icon>
            <el-icon v-else><Close /></el-icon>
            {{ row.enabled ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      
      <el-table-column prop="totpEnabled" label="2FA" width="80">
        <template #default="{ row }">
          <el-tag :type="row.totpEnabled ? 'success' : 'info'" size="small">
            {{ row.totpEnabled ? '已启用' : '未启用' }}
          </el-tag>
        </template>
      </el-table-column>
      
      <el-table-column prop="createdAt" label="创建时间" width="180">
        <template #default="{ row }">
          {{ new Date(row.createdAt).toLocaleString() }}
        </template>
      </el-table-column>
      
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button-group>
            <el-button
              size="small"
              :icon="Edit"
              @click="handleEdit(row)"
            >
              编辑
            </el-button>
            <el-button
              size="small"
              :icon="row.enabled ? Close : Check"
              :type="row.enabled ? 'warning' : 'success'"
              :disabled="isCurrentUser(row)"
              @click="handleToggleEnabled(row)"
            >
              {{ row.enabled ? '禁用' : '启用' }}
            </el-button>
            <el-button
              size="small"
              :icon="Key"
              @click="handleResetPassword(row)"
            >
              重置密码
            </el-button>
            <el-button
              size="small"
              type="danger"
              :icon="Delete"
              :disabled="isCurrentUser(row)"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </el-button-group>
        </template>
      </el-table-column>
    </el-table>

    <!-- User form dialog -->
    <UserForm
      v-model:visible="showUserForm"
      :user="editingUser"
      @success="handleFormSuccess"
    />
  </div>
</template>

<style scoped>
.user-list {
  background: #fff;
  border-radius: 4px;
  padding: 20px;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.list-header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.current-tag {
  margin-left: 8px;
}

.el-button-group {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
</style>
