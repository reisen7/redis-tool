<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Plus, Delete } from '@element-plus/icons-vue'

const props = defineProps<{
  value: string[]
  editing: boolean
}>()

const emit = defineEmits<{
  (e: 'change', value: string[]): void
}>()

// 编辑数据
const editList = ref<string[]>([])

watch(() => props.value, (val: string[]) => {
  editList.value = [...(val || [])]
}, { immediate: true })

watch(() => props.editing, (editing: boolean) => {
  if (editing) {
    editList.value = [...(props.value || [])]
  }
})

// 新增元素
const newItem = ref('')
const insertPosition = ref<'left' | 'right'>('right')

function handleAdd() {
  if (!newItem.value.trim()) return
  
  if (insertPosition.value === 'left') {
    editList.value.unshift(newItem.value)
  } else {
    editList.value.push(newItem.value)
  }
  
  newItem.value = ''
  emit('change', editList.value)
}

// 删除元素
function handleDelete(index: number) {
  editList.value.splice(index, 1)
  emit('change', editList.value)
}

// 修改元素
function handleUpdate(index: number, value: string) {
  editList.value[index] = value
  emit('change', editList.value)
}
</script>

<template>
  <div class="list-viewer">
    <!-- 添加元素（编辑模式） -->
    <div v-if="editing" class="add-item">
      <el-radio-group v-model="insertPosition" size="small">
        <el-radio-button value="left">LPUSH</el-radio-button>
        <el-radio-button value="right">RPUSH</el-radio-button>
      </el-radio-group>
      <el-input
        v-model="newItem"
        placeholder="输入新元素"
        size="small"
        @keyup.enter="handleAdd"
      />
      <el-button type="primary" :icon="Plus" size="small" @click="handleAdd">
        添加
      </el-button>
    </div>

    <!-- 列表内容 -->
    <el-table :data="editing ? editList : value" height="100%">
      <el-table-column label="索引" width="80">
        <template #default="{ $index }">
          <span class="index">{{ $index }}</span>
        </template>
      </el-table-column>
      
      <el-table-column label="值">
        <template #default="{ row, $index }">
          <el-input
            v-if="editing"
            :model-value="row"
            @update:model-value="(val: string) => handleUpdate($index, val)"
            size="small"
          />
          <span v-else class="value-text">{{ row }}</span>
        </template>
      </el-table-column>
      
      <el-table-column v-if="editing" label="操作" width="80">
        <template #default="{ $index }">
          <el-button
            type="danger"
            :icon="Delete"
            size="small"
            link
            @click="handleDelete($index)"
          />
        </template>
      </el-table-column>
    </el-table>

    <!-- 统计 -->
    <div class="list-footer">
      共 {{ (editing ? editList : value)?.length || 0 }} 个元素
    </div>
  </div>
</template>

<style scoped>
.list-viewer {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.add-item {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.add-item .el-input {
  flex: 1;
}

.index {
  color: #909399;
  font-family: monospace;
}

.value-text {
  font-family: monospace;
  word-break: break-all;
}

.list-footer {
  padding: 8px 0;
  text-align: right;
  font-size: 12px;
  color: #909399;
}

:deep(.el-table) {
  flex: 1;
}
</style>
