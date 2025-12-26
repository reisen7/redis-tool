<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import adminApi, { type SystemSettings } from '@/api/admin'

// State
const loading = ref(false)
const saving = ref(false)
const settings = ref<SystemSettings>({})

// Load settings on mount
onMounted(() => {
  loadSettings()
})

// Load all settings
async function loadSettings() {
  loading.value = true
  try {
    const response = await adminApi.getSettings()
    if (response.data.success && response.data.data) {
      settings.value = response.data.data
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '加载设置失败')
  } finally {
    loading.value = false
  }
}

// Update a setting
async function updateSetting(key: string, value: string) {
  saving.value = true
  try {
    const response = await adminApi.updateSetting({ key, value })
    if (response.data.success) {
      settings.value[key] = value
      ElMessage.success('设置已保存')
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '保存设置失败')
    // Reload settings to get the correct value
    loadSettings()
  } finally {
    saving.value = false
  }
}

// Handle registration toggle
function handleRegistrationToggle(enabled: boolean) {
  updateSetting('registration_enabled', enabled ? 'true' : 'false')
}

// Get registration enabled status
function isRegistrationEnabled(): boolean {
  return settings.value.registration_enabled === 'true'
}
</script>

<template>
  <div class="settings-panel" v-loading="loading">
    <!-- Header -->
    <div class="panel-header">
      <h2>系统设置</h2>
    </div>

    <!-- Settings list -->
    <div class="settings-list">
      <!-- Registration setting -->
      <el-card class="setting-card">
        <div class="setting-item">
          <div class="setting-info">
            <h3>用户注册</h3>
            <p>控制是否允许新用户自行注册账户。关闭后，只有管理员可以创建新用户。</p>
          </div>
          <div class="setting-control">
            <el-switch
              :model-value="isRegistrationEnabled()"
              :loading="saving"
              active-text="开启"
              inactive-text="关闭"
              @change="handleRegistrationToggle"
            />
          </div>
        </div>
      </el-card>

      <!-- Placeholder for future settings -->
      <el-card class="setting-card placeholder-card">
        <div class="setting-item">
          <div class="setting-info">
            <h3>更多设置</h3>
            <p>更多系统设置将在后续版本中添加。</p>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<style scoped>
.settings-panel {
  background: #fff;
  border-radius: 4px;
  padding: 20px;
}

.panel-header {
  margin-bottom: 20px;
}

.panel-header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.settings-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.setting-card {
  border-radius: 8px;
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
}

.setting-info {
  flex: 1;
}

.setting-info h3 {
  margin: 0 0 8px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.setting-info p {
  margin: 0;
  font-size: 14px;
  color: #909399;
  line-height: 1.5;
}

.setting-control {
  flex-shrink: 0;
}

.placeholder-card {
  opacity: 0.6;
}

.placeholder-card .setting-info p {
  font-style: italic;
}
</style>
