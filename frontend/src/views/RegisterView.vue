<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, Check, Close } from '@element-plus/icons-vue'
import { authApi } from '@/api/auth'

const router = useRouter()

const form = reactive({
  username: '',
  password: '',
  confirmPassword: ''
})

const loading = ref(false)
const errorMessage = ref('')
const formRef = ref()

// Password strength calculation
const passwordStrength = computed(() => {
  const password = form.password
  if (!password) return { score: 0, label: '', color: '' }
  
  let score = 0
  
  // Length check
  if (password.length >= 8) score += 1
  if (password.length >= 12) score += 1
  
  // Character type checks
  if (/[a-z]/.test(password)) score += 1
  if (/[A-Z]/.test(password)) score += 1
  if (/[0-9]/.test(password)) score += 1
  if (/[^a-zA-Z0-9]/.test(password)) score += 1
  
  if (score <= 2) return { score: 1, label: '弱', color: '#f56c6c' }
  if (score <= 4) return { score: 2, label: '中', color: '#e6a23c' }
  return { score: 3, label: '强', color: '#67c23a' }
})

// Password requirements check
const passwordRequirements = computed(() => {
  const password = form.password
  return {
    minLength: password.length >= 8,
    hasUppercase: /[A-Z]/.test(password),
    hasLowercase: /[a-z]/.test(password),
    hasNumber: /[0-9]/.test(password)
  }
})

const allRequirementsMet = computed(() => {
  const req = passwordRequirements.value
  return req.minLength && req.hasUppercase && req.hasLowercase && req.hasNumber
})

// Validation rules
const validatePassword = (_rule: any, value: string, callback: any) => {
  if (!value) {
    callback(new Error('请输入密码'))
  } else if (!allRequirementsMet.value) {
    callback(new Error('密码不符合要求'))
  } else {
    callback()
  }
}

const validateConfirmPassword = (_rule: any, value: string, callback: any) => {
  if (!value) {
    callback(new Error('请确认密码'))
  } else if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度为 3-50 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, validator: validatePassword, trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

async function handleRegister() {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  
  loading.value = true
  errorMessage.value = ''
  
  try {
    const response = await authApi.register({
      username: form.username,
      password: form.password
    })
    
    if (response.data.success) {
      ElMessage.success('注册成功，请登录')
      router.push({ name: 'login' })
    } else {
      errorMessage.value = response.data.error || '注册失败'
    }
  } catch (error: any) {
    const status = error.response?.status
    if (status === 403) {
      errorMessage.value = '注册功能已关闭'
    } else if (status === 409) {
      errorMessage.value = '用户名已存在'
    } else {
      errorMessage.value = error.response?.data?.error || '注册失败，请稍后重试'
    }
  } finally {
    loading.value = false
  }
}

function goToLogin() {
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="register-container">
    <div class="register-card">
      <div class="register-header">
        <svg class="logo-icon" viewBox="0 0 32 32" width="48" height="48">
          <defs>
            <linearGradient id="redis-gradient-register" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" style="stop-color:#DC382D"/>
              <stop offset="100%" style="stop-color:#A41E11"/>
            </linearGradient>
          </defs>
          <ellipse cx="16" cy="16" rx="14" ry="10" fill="url(#redis-gradient-register)"/>
          <ellipse cx="16" cy="12" rx="14" ry="10" fill="url(#redis-gradient-register)"/>
          <ellipse cx="16" cy="8" rx="14" ry="10" fill="#DC382D"/>
          <ellipse cx="16" cy="8" rx="10" ry="6" fill="#fff" opacity="0.2"/>
        </svg>
        <h1 class="register-title">Redis Manager</h1>
        <p class="register-subtitle">创建新账户</p>
      </div>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        class="register-form"
        @submit.prevent="handleRegister"
      >
        <el-alert
          v-if="errorMessage"
          :title="errorMessage"
          type="error"
          show-icon
          :closable="false"
          class="error-alert"
        />

        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="用户名"
            size="large"
            :prefix-icon="User"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            size="large"
            show-password
            :prefix-icon="Lock"
          />
        </el-form-item>

        <!-- Password strength indicator -->
        <div v-if="form.password" class="password-strength">
          <div class="strength-bar">
            <div 
              class="strength-fill" 
              :style="{ 
                width: `${(passwordStrength.score / 3) * 100}%`,
                backgroundColor: passwordStrength.color 
              }"
            />
          </div>
          <span class="strength-label" :style="{ color: passwordStrength.color }">
            {{ passwordStrength.label }}
          </span>
        </div>

        <!-- Password requirements -->
        <div class="password-requirements">
          <div class="requirement" :class="{ met: passwordRequirements.minLength }">
            <el-icon><Check v-if="passwordRequirements.minLength" /><Close v-else /></el-icon>
            <span>至少 8 个字符</span>
          </div>
          <div class="requirement" :class="{ met: passwordRequirements.hasUppercase }">
            <el-icon><Check v-if="passwordRequirements.hasUppercase" /><Close v-else /></el-icon>
            <span>包含大写字母</span>
          </div>
          <div class="requirement" :class="{ met: passwordRequirements.hasLowercase }">
            <el-icon><Check v-if="passwordRequirements.hasLowercase" /><Close v-else /></el-icon>
            <span>包含小写字母</span>
          </div>
          <div class="requirement" :class="{ met: passwordRequirements.hasNumber }">
            <el-icon><Check v-if="passwordRequirements.hasNumber" /><Close v-else /></el-icon>
            <span>包含数字</span>
          </div>
        </div>

        <el-form-item prop="confirmPassword">
          <el-input
            v-model="form.confirmPassword"
            type="password"
            placeholder="确认密码"
            size="large"
            show-password
            :prefix-icon="Lock"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            class="register-button"
            @click="handleRegister"
          >
            注册
          </el-button>
        </el-form-item>

        <div class="login-link">
          已有账户？
          <el-link type="primary" @click="goToLogin">立即登录</el-link>
        </div>
      </el-form>
    </div>
  </div>
</template>

<style scoped>
.register-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.register-card {
  width: 100%;
  max-width: 400px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  padding: 40px;
}

.register-header {
  text-align: center;
  margin-bottom: 32px;
}

.logo-icon {
  margin-bottom: 16px;
}

.register-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 8px 0;
}

.register-subtitle {
  font-size: 14px;
  color: #909399;
  margin: 0;
}

.register-form {
  width: 100%;
}

.error-alert {
  margin-bottom: 20px;
}

.password-strength {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.strength-bar {
  flex: 1;
  height: 4px;
  background: #e4e7ed;
  border-radius: 2px;
  overflow: hidden;
}

.strength-fill {
  height: 100%;
  transition: width 0.3s, background-color 0.3s;
}

.strength-label {
  font-size: 12px;
  font-weight: 500;
  min-width: 24px;
}

.password-requirements {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 16px;
}

.requirement {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.requirement:last-child {
  margin-bottom: 0;
}

.requirement.met {
  color: #67c23a;
}

.requirement .el-icon {
  font-size: 14px;
}

.register-button {
  width: 100%;
}

.login-link {
  text-align: center;
  font-size: 14px;
  color: #606266;
  margin-top: 16px;
}
</style>
