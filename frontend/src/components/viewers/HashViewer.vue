<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Plus, Delete } from '@element-plus/icons-vue'

const props = defineProps<{
  value: Record<string, string>
  editing: boolean
}>()

const emit = defineEmits<{
  (e: 'change', value: Record<string, string>): void
}>()

// 转换为数组形式便于展示
interface HashItem {
  field: string
  value: string
}

const hashList = computed<HashItem[]>(() => {
  if (!props.value) return []
  return Object.entries(props.value).map(([field, value]) => ({ field, value }))
})

// 编辑数据
const editHash = ref<Record<string, string>>({})

watch(() => props.value, (val: Record<string, string>) => {
  editHash.value = { ...(val || {}) }
}, { immediate: true })

watch(() => props.editing, (editing) => {
  if (editing) {
    editHash.value = { ...(props.value || {}) }
  }
})

// 编辑列表
const editList = computed<HashItem[]>(() => {
  return Object.entries(editHash.value).map(([field, value]) => ({ field, value }))
})

// 新增字段
const newField = ref('')
const newValue = ref('')

function handleAdd() {
  if (!newField.value.trim()) return
  
  editHash.value[newField.value] = newValue.value
  newField.value = ''
  newValue.value = ''
  emit('change', editHash.value)
}

// 删除字段
function handleDelete(field: string) {
  delete editHash.value[field]
  emit('change', { ...editHash.value })
}

// 修改值
function handleUpdateValue(field: string, value: string) {
  editHash.value[field] = value
  emit('change', editHash.value)
}

// 搜索
const searchText = ref('')
const filteredList = computed(() => {
  const list = props.editing ? editList.value : hashList.value
  if (!searchText.value) return list
  
  const search = searchText.value.toLowerCase()
  return list.filter(item => 
    item.field.toLowerCase().includes(search) ||
    item.value.toLowerCase().includes(search)
  )
})
</script>

<template>
  <div class="hash-viewer">
    <!-- 搜索和添加 -->
    <div class="toolbar">
      <el-input
        v-model="searchText"
        placeholder="搜索字段..."
        size="small"
        clearable
        style="width: 200px"
      />
      
      <template v-if="editing">
        <el-input
          v-model="newField"
          placeholder="字段名"
          size="small"
          style="width: 150px"
        />
        <el-input
          v-model="newValue"
          placeholder="值"
          size="small"
          style="flex: 1"
          @keyup.enter="handleAdd"
        />
        <el-button type="primary" :icon="Plus" size="small" @click="handleAdd">
          添加
        </el-button>
      </template>
    </div>

    <!-- Hash 内容 -->
    <el-table :data="filteredList" height="100%">
      <el-table-column label="字段" width="200">
        <template #default="{ row }">
          <span class="field-name">{{ row.field }}</span>
        </template>
      </el-table-column>
      
      <el-table-column label="值">
        <template #default="{ row }">
          <el-input
            v-if="editing"
            :model-value="row.value"
            @update:model-value="(val: string) => handleUpdateValue(row.field, val)"
            size="small"
          />
          <span v-else class="value-text">{{ row.value }}</span>
        </template>
      </el-table-column>
      
      <el-table-column v-if="editing" label="操作" width="80">
        <template #default="{ row }">
          <el-button
            type="danger"
            :icon="Delete"
            size="small"
            link
            @click="handleDelete(row.field)"
          />
        </template>
      </el-table-column>
    </el-table>

    <!-- 统计 -->
    <div class="hash-footer">
      共 {{ (editing ? editList : hashList).length }} 个字段
    </div>
  </div>
</template>

<style scoped>
.hash-viewer {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.field-name {
  font-family: monospace;
  color: #409eff;
}

.value-text {
  font-family: monospace;
  word-break: break-all;
}

.hash-footer {
  padding: 8px 0;
  text-align: right;
  font-size: 12px;
  color: #909399;
}

:deep(.el-table) {
  flex: 1;
}
</style>
