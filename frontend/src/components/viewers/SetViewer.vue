<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Plus, Delete } from '@element-plus/icons-vue'

interface SetItem {
  id: number
  value: string
}

const props = defineProps<{
  value: string[]
  editing: boolean
}>()

const emit = defineEmits<{
  (e: 'change', value: string[]): void
}>()

// 用于生成唯一 ID
let idCounter = 0

// 编辑数据 - 使用带 ID 的对象避免重复 key 问题
const editSet = ref<SetItem[]>([])

// 将字符串数组转换为带 ID 的对象数组
function toSetItems(arr: string[]): SetItem[] {
  return (arr || []).map(v => ({ id: ++idCounter, value: v }))
}

// 将带 ID 的对象数组转换回字符串数组
function toStringArray(items: SetItem[]): string[] {
  return items.map(item => item.value)
}

watch(() => props.value, (val) => {
  editSet.value = toSetItems(val)
}, { immediate: true })

watch(() => props.editing, (editing) => {
  if (editing) {
    editSet.value = toSetItems(props.value)
  }
})

// 新增成员
const newMember = ref('')

function handleAdd() {
  if (!newMember.value.trim()) return
  
  // Set 不允许重复值
  if (editSet.value.some(item => item.value === newMember.value)) {
    return
  }
  
  editSet.value.push({ id: ++idCounter, value: newMember.value })
  newMember.value = ''
  emit('change', toStringArray(editSet.value))
}

// 删除成员
function handleDelete(id: number) {
  const index = editSet.value.findIndex(item => item.id === id)
  if (index >= 0) {
    editSet.value.splice(index, 1)
    emit('change', toStringArray(editSet.value))
  }
}

// 搜索
const searchText = ref('')
const filteredList = computed(() => {
  const list = props.editing ? editSet.value : toSetItems(props.value || [])
  if (!searchText.value) return list
  
  const search = searchText.value.toLowerCase()
  return list.filter(item => item.value.toLowerCase().includes(search))
})

// 显示用的数据
const displayList = computed(() => {
  if (props.editing) {
    return filteredList.value
  }
  return toSetItems(props.value || []).filter(item => {
    if (!searchText.value) return true
    return item.value.toLowerCase().includes(searchText.value.toLowerCase())
  })
})
</script>

<template>
  <div class="set-viewer">
    <!-- 搜索和添加 -->
    <div class="toolbar">
      <el-input
        v-model="searchText"
        placeholder="搜索成员..."
        size="small"
        clearable
        style="width: 200px"
      />
      
      <template v-if="editing">
        <el-input
          v-model="newMember"
          placeholder="新成员"
          size="small"
          style="flex: 1"
          @keyup.enter="handleAdd"
        />
        <el-button type="primary" :icon="Plus" size="small" @click="handleAdd">
          添加
        </el-button>
      </template>
    </div>

    <!-- Set 内容 -->
    <el-table :data="displayList" height="100%" row-key="id">
      <el-table-column label="成员">
        <template #default="{ row }">
          <span class="member-text">{{ row.value }}</span>
        </template>
      </el-table-column>
      
      <el-table-column v-if="editing" label="操作" width="80">
        <template #default="{ row }">
          <el-button
            type="danger"
            :icon="Delete"
            size="small"
            link
            @click="handleDelete(row.id)"
          />
        </template>
      </el-table-column>
    </el-table>

    <!-- 统计 -->
    <div class="set-footer">
      共 {{ (editing ? editSet : toSetItems(value || []))?.length || 0 }} 个成员
    </div>
  </div>
</template>

<style scoped>
.set-viewer {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.member-text {
  font-family: monospace;
  word-break: break-all;
}

.set-footer {
  padding: 8px 0;
  text-align: right;
  font-size: 12px;
  color: #909399;
}

:deep(.el-table) {
  flex: 1;
}
</style>
