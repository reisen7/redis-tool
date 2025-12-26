<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CopyDocument, Check, Warning } from '@element-plus/icons-vue'
import { authApi } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const emit = defineEmits<{
  (e: 'setup-complete'): void
  (e: 'cancel'): void
}>()

const authStore = useAuthStore()

// Setup state
const step = ref<'loading' | 'scan' | 'verify' | 'recovery' | 'complete'>('loading')
const loading = ref(false)
const errorMessage = ref('')

// 2FA data
const qrCode = ref('')
const secret = ref('')
const recoveryCodes = ref<string[]>([])

// Verification form
const verifyForm = reactive({
  code: ''
})
const verifyFormRef = ref()

// Load 2FA setup data
async function loadSetupData() {
  step.value = 'loading'
  
  try {
    const response = await authApi.setup2FA()
    
    if (response.data.success) {
      // Backend wraps data in response.data.data
      const data = response.data.data || response.data
      qrCode.value = data.qrCode || ''
      secret.value = data.secret || ''
      recoveryCodes.value = data.recoveryCodes || []
      step.value = 'scan'
    } else {
      errorMessage.value = response.data.error || '获取设置信息失败'
    }
  } catch (error: any) {
    errorMessage.value = error.response?.data?.error || '获取设置信息失败'
  }
}

// Copy secret to clipboard
async function copySecret() {
  try {
    await navigator.clipboard.writeText(secret.value)
    ElMessage.success('密钥已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

// Copy recovery codes to clipboard
async function copyRecoveryCodes() {
  try {
    await navigator.clipboard.writeText(recoveryCodes.value.join('\n'))
    ElMessage.success('恢复码已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

// Go to verify step
function goToVerify() {
  step.value = 'verify'
}

// Verify TOTP code
async function handleVerify() {
  if (!verifyFormRef.value) return
  
  try {
    await verifyFormRef.value.validate()
  } catch {
    return
  }
  
  loading.value = true
  errorMessage.value = ''
  
  try {
    const response = await authApi.activate2FA({
      code: verifyForm.code.trim()
    })
    
    if (response.data.success) {
      // Update user state
      if (authStore.user) {
        authStore.updateUser({
          ...authStore.user,
          totpEnabled: true
        })
      }
      step.value = 'recovery'
    } else {
      errorMessage.value = response.data.error || '验证码错误'
    }
  } catch (error: any) {
    errorMessage.value = error.response?.data?.error || '验证失败'
  } finally {
    loading.value = false
  }
}

// Complete setup
async function completeSetup() {
  const confirmed = await ElMessageBox.confirm(
    '请确保您已安全保存恢复码。如果丢失验证器，恢复码是唯一的恢复方式。',
    '确认完成',
    {
      confirmButtonText: '我已保存',
      cancelButtonText: '返回',
      type: 'warning'
    }
  ).catch(() => false)
  
  if (confirmed) {
    step.value = 'complete'
    emit('setup-complete')
  }
}

// Cancel setup
function handleCancel() {
  emit('cancel')
}

// Validation rules
const verifyRules = {
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { pattern: /^\d{6}$/, message: '验证码为 6 位数字', trigger: 'blur' }
  ]
}

onMounted(() => {
  loadSetupData()
})
</script>

<template>
  <div class="totp-setup">
    <!-- Loading state -->
    <div v-if="step === 'loading'" class="loading-state">
      <el-icon class="is-loading" :size="48">
        <svg viewBox="0 0 1024 1024">
          <path fill="currentColor" d="M512 64a32 32 0 0 1 32 32v192a32 32 0 0 1-64 0V96a32 32 0 0 1 32-32zm0 640a32 32 0 0 1 32 32v192a32 32 0 1 1-64 0V736a32 32 0 0 1 32-32zm448-192a32 32 0 0 1-32 32H736a32 32 0 1 1 0-64h192a32 32 0 0 1 32 32zm-640 0a32 32 0 0 1-32 32H96a32 32 0 0 1 0-64h192a32 32 0 0 1 32 32z"/>
        </svg>
      </el-icon>
      <p>正在加载...</p>
    </div>

    <!-- Error state -->
    <el-alert
      v-if="errorMessage && step !== 'verify'"
      :title="errorMessage"
      type="error"
      show-icon
      :closable="false"
      class="error-alert"
    />

    <!-- Step 1: Scan QR Code -->
    <div v-if="step === 'scan'" class="setup-step">
      <h3 class="step-title">第 1 步：扫描二维码</h3>
      <p class="step-description">
        使用您的验证器应用（如 Google Authenticator、Microsoft Authenticator 或 Authy）扫描下方二维码。
      </p>
      
      <div class="qr-container">
        <img v-if="qrCode" :src="qrCode" alt="2FA QR Code" class="qr-code" />
        <div v-else class="qr-placeholder">
          <el-icon :size="48"><Warning /></el-icon>
          <p>无法加载二维码</p>
        </div>
      </div>

      <div class="secret-section">
        <p class="secret-label">无法扫描？手动输入密钥：</p>
        <div class="secret-value">
          <code>{{ secret }}</code>
          <el-button text @click="copySecret">
            <el-icon><CopyDocument /></el-icon>
          </el-button>
        </div>
      </div>

      <div class="step-actions">
        <el-button @click="handleCancel">取消</el-button>
        <el-button type="primary" @click="goToVerify">下一步</el-button>
      </div>
    </div>

    <!-- Step 2: Verify Code -->
    <div v-if="step === 'verify'" class="setup-step">
      <h3 class="step-title">第 2 步：验证设置</h3>
      <p class="step-description">
        输入验证器应用中显示的 6 位验证码，以确认设置正确。
      </p>

      <el-form
        ref="verifyFormRef"
        :model="verifyForm"
        :rules="verifyRules"
        class="verify-form"
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
            v-model="verifyForm.code"
            placeholder="6 位验证码"
            size="large"
            maxlength="6"
            class="code-input"
            @keyup.enter="handleVerify"
          />
        </el-form-item>
      </el-form>

      <div class="step-actions">
        <el-button @click="step = 'scan'">上一步</el-button>
        <el-button type="primary" :loading="loading" @click="handleVerify">
          验证
        </el-button>
      </div>
    </div>

    <!-- Step 3: Recovery Codes -->
    <div v-if="step === 'recovery'" class="setup-step">
      <h3 class="step-title">第 3 步：保存恢复码</h3>
      <p class="step-description">
        <el-icon color="#e6a23c"><Warning /></el-icon>
        请将以下恢复码保存在安全的地方。如果您丢失了验证器，可以使用恢复码登录。每个恢复码只能使用一次。
      </p>

      <div class="recovery-codes">
        <div class="codes-grid">
          <code v-for="(code, index) in recoveryCodes" :key="index" class="recovery-code">
            {{ code }}
          </code>
        </div>
        <el-button class="copy-button" @click="copyRecoveryCodes">
          <el-icon><CopyDocument /></el-icon>
          复制所有恢复码
        </el-button>
      </div>

      <el-alert
        title="重要提示"
        type="warning"
        :closable="false"
        class="warning-alert"
      >
        <ul class="warning-list">
          <li>每个恢复码只能使用一次</li>
          <li>请将恢复码保存在安全的地方</li>
          <li>不要与他人分享您的恢复码</li>
        </ul>
      </el-alert>

      <div class="step-actions">
        <el-button type="primary" @click="completeSetup">
          <el-icon><Check /></el-icon>
          我已保存恢复码
        </el-button>
      </div>
    </div>

    <!-- Complete state -->
    <div v-if="step === 'complete'" class="complete-state">
      <el-icon :size="64" color="#67c23a"><Check /></el-icon>
      <h3>双重认证已启用</h3>
      <p>您的账户现在受到双重认证保护。</p>
    </div>
  </div>
</template>

<style scoped>
.totp-setup {
  padding: 20px;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #909399;
}

.loading-state p {
  margin-top: 16px;
}

.error-alert {
  margin-bottom: 20px;
}

.setup-step {
  max-width: 500px;
  margin: 0 auto;
}

.step-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 12px 0;
}

.step-description {
  font-size: 14px;
  color: #606266;
  line-height: 1.6;
  margin-bottom: 24px;
}

.step-description .el-icon {
  vertical-align: middle;
  margin-right: 4px;
}

.qr-container {
  display: flex;
  justify-content: center;
  margin-bottom: 24px;
}

.qr-code {
  width: 200px;
  height: 200px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 8px;
  background: #fff;
}

.qr-placeholder {
  width: 200px;
  height: 200px;
  border: 1px dashed #e4e7ed;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.qr-placeholder p {
  margin-top: 8px;
  font-size: 12px;
}

.secret-section {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
}

.secret-label {
  font-size: 12px;
  color: #909399;
  margin: 0 0 8px 0;
}

.secret-value {
  display: flex;
  align-items: center;
  gap: 8px;
}

.secret-value code {
  flex: 1;
  font-family: monospace;
  font-size: 14px;
  color: #303133;
  word-break: break-all;
}

.verify-form {
  margin-bottom: 24px;
}

.code-input :deep(.el-input__inner) {
  text-align: center;
  font-size: 24px;
  letter-spacing: 8px;
  font-family: monospace;
}

.recovery-codes {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 24px;
}

.codes-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  margin-bottom: 16px;
}

.recovery-code {
  font-family: monospace;
  font-size: 14px;
  color: #303133;
  background: #fff;
  padding: 8px 12px;
  border-radius: 4px;
  text-align: center;
}

.copy-button {
  width: 100%;
}

.warning-alert {
  margin-bottom: 24px;
}

.warning-list {
  margin: 0;
  padding-left: 20px;
  font-size: 12px;
}

.warning-list li {
  margin-bottom: 4px;
}

.warning-list li:last-child {
  margin-bottom: 0;
}

.step-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.complete-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  text-align: center;
}

.complete-state h3 {
  margin: 16px 0 8px 0;
  font-size: 18px;
  color: #303133;
}

.complete-state p {
  margin: 0;
  color: #909399;
}
</style>
