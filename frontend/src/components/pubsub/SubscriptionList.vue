<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Delete, CircleCheck, CircleClose } from '@element-plus/icons-vue'
import { usePubSubStore } from '@/stores/pubsub'
import type { Subscription } from '@/types/pubsub'

const props = defineProps<{
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'subscribe', channel: string, isPattern: boolean): void
  (e: 'unsubscribe', channel: string, isPattern: boolean): void
}>()

const pubsubStore = usePubSubStore()

// New subscription form
const newChannel = ref('')
const isPattern = ref(false)
const showAddForm = ref(false)

// Subscribe to a new channel
function handleSubscribe() {
  const channel = newChannel.value.trim()
  if (!channel) {
    ElMessage.warning('请输入频道名称')
    return
  }

  emit('subscribe', channel, isPattern.value)
  
  // Reset form
  newChannel.value = ''
  isPattern.value = false
  showAddForm.value = false
}

// Unsubscribe from a channel
function handleUnsubscribe(subscription: Subscription) {
  emit('unsubscribe', subscription.channel, subscription.isPattern)
}

// Toggle add form
function toggleAddForm() {
  showAddForm.value = !showAddForm.value
  if (!showAddForm.value) {
    newChannel.value = ''
    isPattern.value = false
  }
}

// Format subscription time
function formatTime(date: Date): string {
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}
</script>

<template>
  <div class="subscription-list">
    <div class="section-header">
      <span class="section-title">订阅列表</span>
      <el-button
        type="primary"
        :icon="Plus"
        size="small"
        text
        :disabled="disabled"
        @click="toggleAddForm"
      >
        添加
      </el-button>
    </div>

    <!-- Add subscription form -->
    <div v-if="showAddForm" class="add-form">
      <el-input
        v-model="newChannel"
        placeholder="频道名称或模式"
        size="small"
        :disabled="disabled"
        @keyup.enter="handleSubscribe"
      >
        <template #prepend>
          <el-checkbox
            v-model="isPattern"
            :disabled="disabled"
            title="使用模式匹配 (PSUBSCRIBE)"
          >
            模式
          </el-checkbox>
        </template>
      </el-input>
      <div class="form-actions">
        <el-button
          type="primary"
          size="small"
          :disabled="disabled || !newChannel.trim()"
          @click="handleSubscribe"
        >
          订阅
        </el-button>
        <el-button
          size="small"
          @click="toggleAddForm"
        >
          取消
        </el-button>
      </div>
      <div class="form-hint">
        <span v-if="isPattern">模式示例: news.*, user:*:messages</span>
        <span v-else>频道示例: news, user:123:messages</span>
      </div>
    </div>

    <!-- Subscription list -->
    <div class="list-content">
      <div v-if="pubsubStore.subscriptions.length === 0" class="empty-state">
        <p>暂无订阅</p>
        <p class="hint">点击"添加"按钮订阅频道</p>
      </div>

      <div
        v-for="sub in pubsubStore.subscriptions"
        :key="sub.id"
        class="subscription-item"
        :class="{ inactive: !sub.active }"
      >
        <div class="item-info">
          <div class="item-header">
            <el-icon :class="sub.active ? 'active-icon' : 'inactive-icon'">
              <CircleCheck v-if="sub.active" />
              <CircleClose v-else />
            </el-icon>
            <span class="channel-name">{{ sub.channel }}</span>
            <el-tag v-if="sub.isPattern" size="small" type="info">模式</el-tag>
          </div>
          <div class="item-meta">
            <span v-if="sub.active" class="subscribed-at">
              订阅于 {{ formatTime(sub.subscribedAt) }}
            </span>
            <span v-else class="status-text">未激活</span>
          </div>
        </div>
        <div class="item-actions">
          <el-button
            type="danger"
            :icon="Delete"
            size="small"
            text
            @click="handleUnsubscribe(sub)"
          >
            取消
          </el-button>
        </div>
      </div>
    </div>

    <!-- Subscription count -->
    <div v-if="pubsubStore.subscriptions.length > 0" class="list-footer">
      <span class="count">
        {{ pubsubStore.activeSubscriptions.length }} / {{ pubsubStore.subscriptions.length }} 个订阅
      </span>
    </div>
  </div>
</template>

<style scoped>
.subscription-list {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
}

.add-form {
  padding: 12px;
  background: #fafafa;
  border-bottom: 1px solid #e4e7ed;
}

.add-form .el-input {
  margin-bottom: 8px;
}

.add-form :deep(.el-input-group__prepend) {
  padding: 0 8px;
}

.form-actions {
  display: flex;
  gap: 8px;
}

.form-hint {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}

.list-content {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.empty-state {
  padding: 32px 16px;
  text-align: center;
  color: #909399;
}

.empty-state p {
  margin: 0;
}

.empty-state .hint {
  font-size: 12px;
  margin-top: 8px;
}

.subscription-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-bottom: 1px solid #f0f0f0;
  transition: background-color 0.2s;
}

.subscription-item:hover {
  background: #f5f7fa;
}

.subscription-item.inactive {
  opacity: 0.6;
}

.item-info {
  flex: 1;
  min-width: 0;
}

.item-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.active-icon {
  color: #67c23a;
}

.inactive-icon {
  color: #909399;
}

.channel-name {
  font-size: 13px;
  font-weight: 500;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-meta {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
  padding-left: 22px;
}

.status-text {
  color: #f56c6c;
}

.item-actions {
  flex-shrink: 0;
}

.list-footer {
  padding: 8px 12px;
  background: #f5f7fa;
  border-top: 1px solid #e4e7ed;
}

.count {
  font-size: 12px;
  color: #909399;
}

/* Scrollbar styles */
.list-content::-webkit-scrollbar {
  width: 6px;
}

.list-content::-webkit-scrollbar-track {
  background: #f0f0f0;
}

.list-content::-webkit-scrollbar-thumb {
  background: #c0c4cc;
  border-radius: 3px;
}

.list-content::-webkit-scrollbar-thumb:hover {
  background: #909399;
}
</style>
