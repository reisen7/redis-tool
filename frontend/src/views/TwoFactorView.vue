<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Key, Ticket } from '@element-plus/icons-vue'
import { authApi } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const form = reactive({
  code: ''
})

const loading = ref(false)
const errorMessage = ref('')
const useRecoveryCode = ref(false)
const formRef = ref()

// Validation rules
const rules = {
  code: [
    { required: true, message: useRecoveryCode.value ? '请输入恢复码' : '请输入验证码', trigger: 'blur' },
    { 
      pattern: useRecoveryCode.value ? /^[a-zA-Z0-9-]+$/ : /^\d{6}$/,
      message: useRecoveryCode.value ? '恢复码格式不正确' : '验证码为 6 位数字',
      trigger: 'blur'
    }
  ]
}

async function handleVerify() {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  
  // Check if we have a temp token
  if (!authStore.tempToken) {
    errorMessage.value = '会话已过期，请重新登录'
    return
  }
  
  loading.value = true
  errorMessage.value = ''
  
  try {
    const response = await authApi.verify2FA({
      tempToken: authStore.tempToken,
      code: form.code.trim()
    })
    
    if (response.data.success && response.data.token && response.data.user) {
      authStore.complete2FA(response.data)
      ElMessage.success('验证成功')
      router.push({ name: 'app' })
    } else {
      errorMessage.value = response.data.error || '验证码错误'
    }
  } catch (error: any) {
    errorMessage.value = error.response?.data?.error || '验证失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

function toggleRecoveryCode() {
  useRecoveryCode.value = !useRecoveryCode.value
  form.code = ''
  errorMessage.value = ''
}

function goBack() {
  authStore.clearAuth()
  router.push({ name: 'login' })
}

// Check if we have a temp token on mount
onMounted(() => {
  if (!authStore.tempToken) {
    // No temp token, redirect to login
    router.push({ name: 'login' })
  }
})
</script>

<template>
  <div class="two-factor-container">
    <div class="two-factor-card">
      <div class="two-factor-header">
        <div class="icon-wrapper">
          <el-icon :size="48" color="#409eff">
            <Key v-if="!useRecoveryCode" />
            <Ticket v-else />
          </el-icon>
        </div>
        <h1 class="two-factor-title">双重认证</h1>
        <p class="two-factor-subtitle">
          {{ useRecoveryCode ? '请输入恢复码' : '请输入您的验证器应用中的 6 位验证码' }}
        </p>
      </div>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        class="two-factor-form"
        @submit.prevent="handleVerify"
      >
        <el-alert
          v-if="errorMessage"
          :title="errorMessage"
          type="error"
          show-icon
          :closable="false"
          class="error-alert"
        />

        <el-form-item prop="code">
          <el-input
            v-model="form.code"
            :placeholder="useRecoveryCode ? '恢复码' : '6 位验证码'"
            size="large"
            :maxlength="useRecoveryCode ? 20 : 6"
            class="code-input"
            @keyup.enter="handleVerify"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            class="verify-button"
            @click="handleVerify"
          >
            验证
          </el-button>
        </el-form-item>

        <div class="action-links">
          <el-link type="primary" @click="toggleRecoveryCode">
            {{ useRecoveryCode ? '使用验证码' : '使用恢复码' }}
          </el-link>
          <el-link type="info" @click="goBack">
            返回登录
          </el-link>
        </div>

        <div v-if="useRecoveryCode" class="recovery-hint">
          <el-alert
            title="恢复码说明"
            type="info"
            :closable="false"
          >
            <p>恢复码是您在启用双重认证时生成的一次性代码。</p>
            <p>每个恢复码只能使用一次。</p>
          </el-alert>
        </div>
      </el-form>
    </div>
  </div>
</template>

<style scoped>
.two-factor-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.two-factor-card {
  width: 100%;
  max-width: 400px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  padding: 40px;
}

.two-factor-header {
  text-align: center;
  margin-bottom: 32px;
}

.icon-wrapper {
  margin-bottom: 16px;
}

.two-factor-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 8px 0;
}

.two-factor-subtitle {
  font-size: 14px;
  color: #909399;
  margin: 0;
  line-height: 1.5;
}

.two-factor-form {
  width: 100%;
}

.error-alert {
  margin-bottom: 20px;
}

.code-input :deep(.el-input__inner) {
  text-align: center;
  font-size: 24px;
  letter-spacing: 8px;
  font-family: monospace;
}

.verify-button {
  width: 100%;
}

.action-links {
  display: flex;
  justify-content: space-between;
  margin-top: 16px;
}

.recovery-hint {
  margin-top: 20px;
}

.recovery-hint p {
  margin: 4px 0;
  font-size: 12px;
}
</style>
