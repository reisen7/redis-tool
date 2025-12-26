<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Edit, Delete, Timer, DocumentCopy } from '@element-plus/icons-vue'
import { formatTTL, TTL_PRESETS, getTypeColor } from '@/types/key'
import { useConnectionStore } from '@/stores/connection'
import { useKeysStore } from '@/stores/keys'
import StringViewer from './viewers/StringViewer.vue'
import ListViewer from './viewers/ListViewer.vue'
import HashViewer from './viewers/HashViewer.vue'
import SetViewer from './viewers/SetViewer.vue'
import ZSetViewer from './viewers/ZSetViewer.vue'

const connectionStore = useConnectionStore()
const keysStore = useKeysStore()

// 编辑模式
const editing = ref(false)
const editValue = ref<any>(null)

// TTL 设置
const showTTLDialog = ref(false)
const ttlValue = ref(0)
const ttlPreset = ref<number | null>(null)

// 当前 Key 值
const keyValue = computed(() => keysStore.currentKeyValue)
const loading = computed(() => keysStore.loadingValue)

// 当前连接和数据库
const currentConnectionId = computed(() => connectionStore.activeConnectionId)
const currentDb = computed(() => connectionStore.activeConnection?.currentDb ?? 0)

// TTL 倒计时
const displayTTL = ref(-1)
let ttlTimer: number | null = null

// 清理定时器
function clearTTLTimer() {
  if (ttlTimer !== null) {
    clearInterval(ttlTimer)
    ttlTimer = null
  }
}

// 监听 Key 值变化，更新 TTL 倒计时
watch(() => keyValue.value?.ttl, (ttl) => {
  clearTTLTimer()
  
  if (ttl !== undefined && ttl !== null && ttl > 0) {
    displayTTL.value = ttl
    // 启动倒计时
    ttlTimer = window.setInterval(() => {
      if (displayTTL.value > 0) {
        displayTTL.value--
      } else {
        clearTTLTimer()
      }
    }, 1000)
  } else {
    displayTTL.value = ttl ?? -1
  }
}, { immediate: true })

// 组件卸载时清理定时器
onUnmounted(() => {
  clearTTLTimer()
})

// 刷新
async function handleRefresh() {
  if (!currentConnectionId.value || !keysStore.selectedKey) return
  await keysStore.refreshCurrentKey(currentConnectionId.value, currentDb.value)
}

// 开始编辑
function handleEdit() {
  if (!keyValue.value) return
  editValue.value = JSON.parse(JSON.stringify(keyValue.value.value))
  editing.value = true
}

// 取消编辑
function handleCancelEdit() {
  editing.value = false
  editValue.value = null
}

// 保存编辑
async function handleSaveEdit() {
  if (!currentConnectionId.value || !keysStore.selectedKey) return
  
  const result = await keysStore.updateKeyValue(
    currentConnectionId.value,
    currentDb.value,
    keysStore.selectedKey,
    editValue.value
  )
  
  if (result.success) {
    ElMessage.success('保存成功')
    editing.value = false
    editValue.value = null
  } else {
    ElMessage.error(result.error || '保存失败')
  }
}

// 删除 Key
async function handleDelete() {
  if (!currentConnectionId.value || !keysStore.selectedKey) return
  
  try {
    await ElMessageBox.confirm(
      `确定要删除 Key "${keysStore.selectedKey}" 吗？`,
      '删除确认',
      { type: 'warning' }
    )
    
    const result = await keysStore.deleteKeys(
      currentConnectionId.value,
      currentDb.value,
      [keysStore.selectedKey]
    )
    
    if (result.success) {
      ElMessage.success('删除成功')
    } else {
      ElMessage.error(result.error || '删除失败')
    }
  } catch {}
}

// 打开 TTL 设置
function handleOpenTTL() {
  ttlValue.value = keyValue.value?.ttl || 0
  ttlPreset.value = null
  showTTLDialog.value = true
}

// 选择 TTL 预设
function handleSelectPreset(value: number) {
  ttlPreset.value = value
  ttlValue.value = value
}

// 保存 TTL
async function handleSaveTTL() {
  if (!currentConnectionId.value || !keysStore.selectedKey) return
  
  let result
  if (ttlValue.value < 0) {
    result = await keysStore.removeKeyTTL(
      currentConnectionId.value,
      currentDb.value,
      keysStore.selectedKey
    )
  } else {
    result = await keysStore.setKeyTTL(
      currentConnectionId.value,
      currentDb.value,
      keysStore.selectedKey,
      ttlValue.value
    )
  }
  
  if (result.success) {
    ElMessage.success('TTL 设置成功')
    showTTLDialog.value = false
  } else {
    ElMessage.error(result.error || 'TTL 设置失败')
  }
}

// 复制 Key 名称
function handleCopyKey() {
  if (!keysStore.selectedKey) return
  navigator.clipboard.writeText(keysStore.selectedKey)
  ElMessage.success('已复制到剪贴板')
}

// 更新值（子组件调用）
function handleValueChange(value: any) {
  editValue.value = value
}
</script>

<template>
  <div class="key-detail" v-loading="loading">
    <!-- 空状态 -->
    <el-empty
      v-if="!keyValue"
      description="选择一个 Key 查看详情"
      :image-size="100"
    />

    <template v-else>
      <!-- 头部信息 -->
      <div class="detail-header">
        <div class="key-info">
          <div class="key-name-row">
            <span class="key-name">{{ keyValue.key }}</span>
            <el-button
              type="primary"
              link
              size="small"
              :icon="DocumentCopy"
              @click="handleCopyKey"
            />
          </div>
          
          <div class="key-meta">
            <el-tag
              size="small"
              :color="getTypeColor(keyValue.type)"
              effect="dark"
            >
              {{ keyValue.type.toUpperCase() }}
            </el-tag>
            
            <span class="ttl-info" @click="handleOpenTTL">
              <el-icon><Timer /></el-icon>
              TTL: {{ formatTTL(displayTTL) }}
            </span>
            
            <span v-if="keyValue.size" class="size-info">
              大小: {{ (keyValue.size / 1024).toFixed(2) }} KB
            </span>
          </div>
        </div>

        <div class="header-actions">
          <el-button :icon="Refresh" size="small" @click="handleRefresh">
            刷新
          </el-button>
          <el-button
            v-if="!editing"
            type="primary"
            :icon="Edit"
            size="small"
            @click="handleEdit"
          >
            编辑
          </el-button>
          <template v-else>
            <el-button size="small" @click="handleCancelEdit">取消</el-button>
            <el-button type="primary" size="small" @click="handleSaveEdit">
              保存
            </el-button>
          </template>
          <el-button
            type="danger"
            :icon="Delete"
            size="small"
            @click="handleDelete"
          >
            删除
          </el-button>
        </div>
      </div>

      <!-- 值查看器 -->
      <div class="detail-content">
        <StringViewer
          v-if="keyValue.type === 'string'"
          :value="editing ? editValue : keyValue.value"
          :editing="editing"
          @change="handleValueChange"
        />
        
        <ListViewer
          v-else-if="keyValue.type === 'list'"
          :value="editing ? editValue : keyValue.value"
          :editing="editing"
          @change="handleValueChange"
        />
        
        <HashViewer
          v-else-if="keyValue.type === 'hash'"
          :value="editing ? editValue : keyValue.value"
          :editing="editing"
          @change="handleValueChange"
        />
        
        <SetViewer
          v-else-if="keyValue.type === 'set'"
          :value="editing ? editValue : keyValue.value"
          :editing="editing"
          @change="handleValueChange"
        />
        
        <ZSetViewer
          v-else-if="keyValue.type === 'zset'"
          :value="editing ? editValue : keyValue.value"
          :editing="editing"
          @change="handleValueChange"
        />
        
        <div v-else class="unsupported">
          不支持的数据类型: {{ keyValue.type }}
        </div>
      </div>
    </template>

    <!-- TTL 设置对话框 -->
    <el-dialog v-model="showTTLDialog" title="设置过期时间" width="400px">
      <div class="ttl-presets">
        <el-button
          v-for="preset in TTL_PRESETS"
          :key="preset.value"
          size="small"
          :type="ttlPreset === preset.value ? 'primary' : 'default'"
          @click="handleSelectPreset(preset.value)"
        >
          {{ preset.label }}
        </el-button>
      </div>
      
      <el-form-item label="自定义 (秒)">
        <el-input-number
          v-model="ttlValue"
          :min="-1"
          :max="999999999"
          style="width: 100%"
        />
        <div class="ttl-hint">-1 表示永不过期</div>
      </el-form-item>

      <template #footer>
        <el-button @click="showTTLDialog = false">取消</el-button>
        <el-button type="primary" @click="handleSaveTTL">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.key-detail {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 16px;
  border-bottom: 1px solid #e4e7ed;
  background: #fafafa;
}

.key-info {
  flex: 1;
  min-width: 0;
}

.key-name-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.key-name {
  font-size: 16px;
  font-weight: 600;
  word-break: break-all;
}

.key-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 12px;
  color: #909399;
}

.ttl-info {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
}

.ttl-info:hover {
  color: #409eff;
}

.header-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.detail-content {
  flex: 1;
  overflow: auto;
  padding: 16px;
}

.unsupported {
  text-align: center;
  color: #909399;
  padding: 40px;
}

.ttl-presets {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 16px;
}

.ttl-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
</style>
