<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Promotion } from '@element-plus/icons-vue'
import { pubsubApi } from '@/api/pubsub'
import { usePubSubStore } from '@/stores/pubsub'

const props = defineProps<{
  disabled?: boolean
  connectionId: string | null
}>()

const pubsubStore = usePubSubStore()

// Form state
const channel = ref('')
const message = ref('')
const publishing = ref(false)

// Recent channels for quick selection
const recentChannels = computed(() => {
  const channels = new Set<string>()
  pubsubStore.subscriptions.forEach(sub => channels.add(sub.channel))
  return Array.from(channels).slice(0, 5)
})

// Publish message
async function handlePublish() {
  const channelValue = channel.value.trim()
  const messageValue = message.value.trim()

  if (!channelValue) {
    ElMessage.warning('请输入频道名称')
    return
  }

  if (!messageValue) {
    ElMessage.warning('请输入消息内容')
    return
  }

  if (!props.connectionId) {
    ElMessage.warning('请先连接 Redis')
    return
  }

  publishing.value = true

  try {
    const response = await pubsubApi.publish({
      connectionId: parseInt(props.connectionId),
      channel: channelValue,
      message: messageValue
    })

    if (response.data.success) {
      ElMessage.success(`消息已发布到 ${channelValue}`)
      // Clear message but keep channel for convenience
      message.value = ''
    } else {
      ElMessage.error(response.data.error || '发布失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '发布失败')
  } finally {
    publishing.value = false
  }
}

// Select recent channel
function selectChannel(ch: string) {
  channel.value = ch
}

// Format JSON
function formatJson() {
  try {
    const parsed = JSON.parse(message.value)
    message.value = JSON.stringify(parsed, null, 2)
  } catch {
    ElMessage.warning('无效的 JSON 格式')
  }
}

// Minify JSON
function minifyJson() {
  try {
    const parsed = JSON.parse(message.value)
    message.value = JSON.stringify(parsed)
  } catch {
    ElMessage.warning('无效的 JSON 格式')
  }
}
</script>

<template>
  <div class="publish-form">
    <div class="section-header">
      <span class="section-title">发布消息</span>
    </div>

    <div class="form-content">
      <!-- Channel input -->
      <div class="form-item">
        <label class="form-label">频道</label>
        <el-input
          v-model="channel"
          placeholder="输入频道名称"
          size="small"
          :disabled="disabled"
        />
        <!-- Recent channels -->
        <div v-if="recentChannels.length > 0" class="recent-channels">
          <span class="recent-label">最近:</span>
          <el-tag
            v-for="ch in recentChannels"
            :key="ch"
            size="small"
            type="info"
            class="recent-tag"
            @click="selectChannel(ch)"
          >
            {{ ch }}
          </el-tag>
        </div>
      </div>

      <!-- Message input -->
      <div class="form-item">
        <div class="label-row">
          <label class="form-label">消息</label>
          <div class="json-actions">
            <el-button
              size="small"
              text
              :disabled="disabled || !message.trim()"
              @click="formatJson"
            >
              格式化
            </el-button>
            <el-button
              size="small"
              text
              :disabled="disabled || !message.trim()"
              @click="minifyJson"
            >
              压缩
            </el-button>
          </div>
        </div>
        <el-input
          v-model="message"
          type="textarea"
          placeholder="输入消息内容 (支持 JSON)"
          :rows="4"
          :disabled="disabled"
          resize="vertical"
        />
      </div>

      <!-- Publish button -->
      <div class="form-actions">
        <el-button
          type="primary"
          :icon="Promotion"
          :loading="publishing"
          :disabled="disabled || !channel.trim() || !message.trim()"
          @click="handlePublish"
        >
          发布
        </el-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.publish-form {
  border-top: 1px solid #e4e7ed;
}

.section-header {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
}

.form-content {
  padding: 12px;
}

.form-item {
  margin-bottom: 12px;
}

.form-label {
  display: block;
  margin-bottom: 6px;
  font-size: 12px;
  font-weight: 500;
  color: #606266;
}

.label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
}

.json-actions {
  display: flex;
  gap: 4px;
}

.json-actions .el-button {
  padding: 2px 6px;
  font-size: 11px;
}

.recent-channels {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.recent-label {
  font-size: 11px;
  color: #909399;
}

.recent-tag {
  cursor: pointer;
  transition: all 0.2s;
}

.recent-tag:hover {
  background: #409eff;
  color: #fff;
  border-color: #409eff;
}

.form-actions {
  margin-top: 16px;
}

.form-actions .el-button {
  width: 100%;
}
</style>
