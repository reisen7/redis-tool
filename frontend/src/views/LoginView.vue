<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { authApi } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const form = reactive({
  username: '',
  password: ''
})

const loading = ref(false)
const errorMessage = ref('')
const registrationEnabled = ref(true) // TODO: Fetch from settings API

// Form validation rules
const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

const formRef = ref()

async function handleLogin() {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  
  loading.value = true
  errorMessage.value = ''
  
  try {
    const response = await authApi.login({
      username: form.username,
      password: form.password
    })
    
    if (response.data.success) {
      const data = response.data.data
      if (data.requires2FA && data.tempToken) {
        // Need 2FA verification
        authStore.setTempToken(data.tempToken)
        router.push({ name: 'two-factor' })
      } else if (data.token && data.user) {
        // Login successful
        authStore.setAuth(data)
        ElMessage.success('登录成功')
        
        // Redirect to original destination or app
        const redirect = route.query.redirect as string
        router.push(redirect || { name: 'app' })
      }
    } else {
      errorMessage.value = response.data.error || '用户名或密码错误'
    }
  } catch (error: any) {
    errorMessage.value = error.response?.data?.error || '登录失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

function goToRegister() {
  router.push({ name: 'register' })
}

// Check registration status on mount
onMounted(async () => {
  // TODO: Fetch registration enabled status from settings API
  // For now, default to true
  registrationEnabled.value = true
})
</script>

<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <svg class="logo-icon" viewBox="0 0 32 32" width="48" height="48">
          <defs>
            <linearGradient id="redis-gradient-login" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" style="stop-color:#DC382D"/>
              <stop offset="100%" style="stop-color:#A41E11"/>
            </linearGradient>
          </defs>
          <ellipse cx="16" cy="16" rx="14" ry="10" fill="url(#redis-gradient-login)"/>
          <ellipse cx="16" cy="12" rx="14" ry="10" fill="url(#redis-gradient-login)"/>
          <ellipse cx="16" cy="8" rx="14" ry="10" fill="#DC382D"/>
          <ellipse cx="16" cy="8" rx="10" ry="6" fill="#fff" opacity="0.2"/>
        </svg>
        <h1 class="login-title">Redis Manager</h1>
        <p class="login-subtitle">登录到您的账户</p>
      </div>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        class="login-form"
        @submit.prevent="handleLogin"
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
            @keyup.enter="handleLogin"
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
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            class="login-button"
            @click="handleLogin"
          >
            登录
          </el-button>
          密码：Admin@123
        </el-form-item>

        <div v-if="registrationEnabled" class="register-link">
          还没有账户？
          <el-link type="primary" @click="goToRegister">立即注册</el-link>
        </div>
      </el-form>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  padding: 40px;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.logo-icon {
  margin-bottom: 16px;
}

.login-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 8px 0;
}

.login-subtitle {
  font-size: 14px;
  color: #909399;
  margin: 0;
}

.login-form {
  width: 100%;
}

.error-alert {
  margin-bottom: 20px;
}

.login-button {
  width: 100%;
}

.register-link {
  text-align: center;
  font-size: 14px;
  color: #606266;
  margin-top: 16px;
}
</style>
