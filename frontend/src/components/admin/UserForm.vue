<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import adminApi from '@/api/admin'
import type { User, UserRole } from '@/types/auth'

// Props
const props = defineProps<{
  visible: boolean
  user: User | null
}>()

// Emits
const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'success'): void
}>()

// Form ref
const formRef = ref<FormInstance>()

// Form data
const formData = ref({
  username: '',
  password: '',
  role: 'user' as UserRole
})

// Loading state
const loading = ref(false)

// Computed
const isEdit = computed(() => props.user !== null)
const dialogTitle = computed(() => isEdit.value ? '编辑用户' : '新建用户')

// Dialog visibility
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
})

// Password validation
function validatePassword(_rule: any, value: string, callback: any) {
  if (!isEdit.value && !value) {
    callback(new Error('请输入密码'))
    return
  }
  if (value && value.length < 8) {
    callback(new Error('密码长度至少8位'))
    return
  }
  if (value && !/[A-Z]/.test(value)) {
    callback(new Error('密码需包含大写字母'))
    return
  }
  if (value && !/[a-z]/.test(value)) {
    callback(new Error('密码需包含小写字母'))
    return
  }
  if (value && !/[0-9]/.test(value)) {
    callback(new Error('密码需包含数字'))
    return
  }
  callback()
}

// Form rules
const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度3-50个字符', trigger: 'blur' }
  ],
  password: [
    { validator: validatePassword, trigger: 'blur' }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
}

// Watch for user changes to populate form
watch(() => props.user, (newUser) => {
  if (newUser) {
    formData.value = {
      username: newUser.username,
      password: '',
      role: newUser.role
    }
  } else {
    formData.value = {
      username: '',
      password: '',
      role: 'user'
    }
  }
}, { immediate: true })

// Reset form when dialog opens
watch(() => props.visible, (visible) => {
  if (visible) {
    formRef.value?.clearValidate()
  }
})

// Handle form submit
async function handleSubmit() {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  
  loading.value = true
  try {
    if (isEdit.value && props.user) {
      // Update existing user (only role can be changed via this form)
      const response = await adminApi.updateUser(props.user.id, {
        role: formData.value.role
      })
      if (response.data.success) {
        ElMessage.success('用户已更新')
        emit('success')
      }
    } else {
      // Create new user
      const response = await adminApi.createUser({
        username: formData.value.username,
        password: formData.value.password,
        role: formData.value.role
      })
      if (response.data.success) {
        ElMessage.success('用户已创建')
        emit('success')
      }
    }
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '操作失败')
  } finally {
    loading.value = false
  }
}

// Handle dialog close
function handleClose() {
  dialogVisible.value = false
}
</script>

<template>
  <el-dialog
    v-model="dialogVisible"
    :title="dialogTitle"
    width="450px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-width="80px"
      label-position="right"
    >
      <el-form-item label="用户名" prop="username">
        <el-input
          v-model="formData.username"
          :disabled="isEdit"
          placeholder="请输入用户名"
          maxlength="50"
        />
      </el-form-item>
      
      <el-form-item v-if="!isEdit" label="密码" prop="password">
        <el-input
          v-model="formData.password"
          type="password"
          placeholder="请输入密码"
          show-password
        />
        <div class="password-hint">
          密码需至少8位，包含大小写字母和数字
        </div>
      </el-form-item>
      
      <el-form-item label="角色" prop="role">
        <el-select v-model="formData.role" placeholder="请选择角色" style="width: 100%">
          <el-option label="普通用户" value="user" />
          <el-option label="管理员" value="admin" />
        </el-select>
      </el-form-item>
    </el-form>
    
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleSubmit">
        {{ isEdit ? '保存' : '创建' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.password-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.4;
}
</style>
