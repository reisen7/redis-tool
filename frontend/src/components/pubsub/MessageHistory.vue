<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { Delete, Download, Search } from '@element-plus/icons-vue'
import { usePubSubStore } from '@/stores/pubsub'
import type { PubSubMessage } from '@/types/pubsub'

const emit = defineEmits<{
  (e: 'clear'): void
}>()

const pubsubStore = usePubSubStore()

// Filter
const filterChannel = ref('')
const autoScroll = ref(true)

// Message list ref for auto-scroll
const messageListRef = ref<HTMLDivElement>()

// Filtered messages
const filteredMessages = computed(() => {
  if (!filterChannel.value.trim()) {
    return pubsubStore.messages
  }
  
  const filter = filterChannel.value.toLowerCase()
  return pubsubStore.messages.filter(msg => 
    msg.channel.toLowerCase().includes(filter)
  )
})

// Watch for new messages and auto-scroll
watch(() => pubsubStore.messages.length, () => {
  if (autoScroll.value) {
    scrollToBottom()
  }
})

// Scroll to bottom
function scrollToBottom() {
  nextTick(() => {
    if (messageListRef.value) {
      messageListRef.value.scrollTop = messageListRef.value.scrollHeight
    }
  })
}

// Handle scroll - disable auto-scroll if user scrolls up
function handleScroll() {
  if (!messageListRef.value) return
  
  const { scrollTop, scrollHeight, clientHeight } = messageListRef.value
  const isAtBottom = scrollHeight - scrollTop - clientHeight < 50
  autoScroll.value = isAtBottom
}

// Clear messages
function handleClear() {
  emit('clear')
}

// Export messages as JSON
function handleExport() {
  const data = JSON.stringify(pubsubStore.messages, null, 2)
  const blob = new Blob([data], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `pubsub-messages-${Date.now()}.json`
  a.click()
  URL.revokeObjectURL(url)
}

// Format timestamp
function formatTimestamp(date: Date): string {
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    fractionalSecondDigits: 3
  })
}

// Format message content (try to parse JSON)
function formatMessage(content: string): { formatted: string; isJson: boolean } {
  try {
    const parsed = JSON.parse(content)
    return {
      formatted: JSON.stringify(parsed, null, 2),
      isJson: true
    }
  } catch {
    return {
      formatted: content,
      isJson: false
    }
  }
}

// Get unique channels for filter suggestions
const uniqueChannels = computed(() => {
  const channels = new Set<string>()
  pubsubStore.messages.forEach(msg => channels.add(msg.channel))
  return Array.from(channels)
})
</script>

<template>
  <div class="message-history">
    <!-- Header -->
    <div class="section-header">
      <span class="section-title">消息历史</span>
      <span class="message-count">{{ filteredMessages.length }} 条消息</span>
      <div class="header-actions">
        <el-button
          :icon="Download"
          size="small"
          text
          :disabled="pubsubStore.messages.length === 0"
          title="导出消息"
          @click="handleExport"
        />
        <el-button
          :icon="Delete"
          size="small"
          text
          type="danger"
          :disabled="pubsubStore.messages.length === 0"
          title="清空消息"
          @click="handleClear"
        />
      </div>
    </div>

    <!-- Filter bar -->
    <div class="filter-bar">
      <el-input
        v-model="filterChannel"
        placeholder="按频道筛选..."
        size="small"
        clearable
        :prefix-icon="Search"
      />
      <el-checkbox v-model="autoScroll" size="small">
        自动滚动
      </el-checkbox>
    </div>

    <!-- Message list -->
    <div 
      ref="messageListRef" 
      class="message-list"
      @scroll="handleScroll"
    >
      <div v-if="filteredMessages.length === 0" class="empty-state">
        <p v-if="pubsubStore.messages.length === 0">暂无消息</p>
        <p v-else>没有匹配的消息</p>
        <p class="hint">订阅频道后，收到的消息将显示在这里</p>
      </div>

      <div
        v-for="msg in filteredMessages"
        :key="msg.id"
        class="message-item"
      >
        <div class="message-header">
          <span class="channel-badge">{{ msg.channel }}</span>
          <span v-if="msg.pattern" class="pattern-badge">
            via {{ msg.pattern }}
          </span>
          <span class="timestamp">{{ formatTimestamp(msg.timestamp) }}</span>
        </div>
        <div class="message-content">
          <pre 
            :class="{ 'json-content': formatMessage(msg.message).isJson }"
          >{{ formatMessage(msg.message).formatted }}</pre>
        </div>
      </div>
    </div>

    <!-- Auto-scroll indicator -->
    <div 
      v-if="!autoScroll && pubsubStore.messages.length > 0" 
      class="scroll-indicator"
      @click="autoScroll = true; scrollToBottom()"
    >
      <span>有新消息，点击滚动到底部</span>
    </div>
  </div>
</template>

<style scoped>
.message-history {
  height: 100%;
  display: flex;
  flex-direction: column;
  position: relative;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
}

.message-count {
  font-size: 12px;
  color: #909399;
}

.header-actions {
  margin-left: auto;
  display: flex;
  gap: 4px;
}

.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  background: #fafafa;
  border-bottom: 1px solid #e4e7ed;
}

.filter-bar .el-input {
  flex: 1;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.empty-state {
  padding: 48px 16px;
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

.message-item {
  margin-bottom: 8px;
  padding: 10px 12px;
  background: #f9f9f9;
  border-radius: 6px;
  border: 1px solid #ebeef5;
}

.message-item:hover {
  background: #f5f7fa;
  border-color: #dcdfe6;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.channel-badge {
  display: inline-block;
  padding: 2px 8px;
  background: #409eff;
  color: #fff;
  font-size: 12px;
  font-weight: 500;
  border-radius: 4px;
}

.pattern-badge {
  font-size: 11px;
  color: #909399;
  font-style: italic;
}

.timestamp {
  margin-left: auto;
  font-size: 11px;
  color: #909399;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
}

.message-content {
  background: #fff;
  border-radius: 4px;
  border: 1px solid #ebeef5;
  overflow: hidden;
}

.message-content pre {
  margin: 0;
  padding: 10px 12px;
  font-size: 12px;
  font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
  color: #303133;
}

.message-content pre.json-content {
  background: #fafafa;
  color: #606266;
}

.scroll-indicator {
  position: absolute;
  bottom: 16px;
  left: 50%;
  transform: translateX(-50%);
  padding: 8px 16px;
  background: #409eff;
  color: #fff;
  font-size: 12px;
  border-radius: 20px;
  cursor: pointer;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
  transition: all 0.2s;
}

.scroll-indicator:hover {
  background: #66b1ff;
}

/* Scrollbar styles */
.message-list::-webkit-scrollbar {
  width: 6px;
}

.message-list::-webkit-scrollbar-track {
  background: #f0f0f0;
}

.message-list::-webkit-scrollbar-thumb {
  background: #c0c4cc;
  border-radius: 3px;
}

.message-list::-webkit-scrollbar-thumb:hover {
  background: #909399;
}
</style>
