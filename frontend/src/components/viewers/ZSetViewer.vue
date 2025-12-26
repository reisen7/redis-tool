<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Plus, Delete, SortUp, SortDown } from '@element-plus/icons-vue'

interface ZSetItem {
  member: string
  score: number
}

interface ZSetEditItem {
  id: number
  member: string
  score: number
}

const props = defineProps<{
  value: ZSetItem[]
  editing: boolean
}>()

const emit = defineEmits<{
  (e: 'change', value: ZSetItem[]): void
}>()

// 用于生成唯一 ID
let idCounter = 0

// 编辑数据 - 使用带 ID 的对象避免排序导致的问题
const editList = ref<ZSetEditItem[]>([])

// 转换函数
function toEditItems(arr: ZSetItem[]): ZSetEditItem[] {
  return (arr || []).map(item => ({ id: ++idCounter, member: item.member, score: item.score }))
}

function toZSetItems(items: ZSetEditItem[]): ZSetItem[] {
  return items.map(item => ({ member: item.member, score: item.score }))
}

watch(() => props.value, (val: ZSetItem[]) => {
  if (!props.editing) {
    editList.value = toEditItems(val)
  }
}, { immediate: true })

watch(() => props.editing, (editing) => {
  if (editing) {
    editList.value = toEditItems(props.value)
  }
})

// 排序方向（仅用于显示）
const sortOrder = ref<'asc' | 'desc'>('asc')

// 显示用的排序列表（不影响编辑数据的顺序）
const displayList = computed(() => {
  const list = props.editing ? editList.value : toEditItems(props.value || [])
  // 创建副本进行排序，不修改原数组
  return [...list].sort((a, b) => {
    return sortOrder.value === 'asc' ? a.score - b.score : b.score - a.score
  })
})

// 新增成员
const newMember = ref('')
const newScore = ref(0)

function handleAdd() {
  if (!newMember.value.trim()) return
  
  // 检查是否已存在
  const existItem = editList.value.find(item => item.member === newMember.value)
  if (existItem) {
    // 更新分数
    existItem.score = newScore.value
  } else {
    editList.value.push({ id: ++idCounter, member: newMember.value, score: newScore.value })
  }
  
  newMember.value = ''
  newScore.value = 0
  emit('change', toZSetItems(editList.value))
}

// 删除成员
function handleDelete(id: number) {
  const index = editList.value.findIndex(item => item.id === id)
  if (index >= 0) {
    editList.value.splice(index, 1)
    emit('change', toZSetItems(editList.value))
  }
}

// 修改分数 - 通过 ID 查找，避免排序问题
function handleUpdateScore(id: number, score: number) {
  const item = editList.value.find(item => item.id === id)
  if (item) {
    item.score = score
    emit('change', toZSetItems(editList.value))
  }
}

// 切换排序
function toggleSort() {
  sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
}
</script>

<template>
  <div class="zset-viewer">
    <!-- 工具栏 -->
    <div class="toolbar">
      <el-button size="small" @click="toggleSort">
        <el-icon>
          <SortUp v-if="sortOrder === 'asc'" />
          <SortDown v-else />
        </el-icon>
        {{ sortOrder === 'asc' ? '升序' : '降序' }}
      </el-button>
      
      <template v-if="editing">
        <el-input
          v-model="newMember"
          placeholder="成员"
          size="small"
          style="width: 200px"
        />
        <el-input-number
          v-model="newScore"
          placeholder="分数"
          size="small"
          style="width: 120px"
          :controls="false"
        />
        <el-button type="primary" :icon="Plus" size="small" @click="handleAdd">
          添加
        </el-button>
      </template>
    </div>

    <!-- ZSet 内容 -->
    <el-table :data="displayList" height="100%" row-key="id">
      <el-table-column label="排名" width="80">
        <template #default="{ $index }">
          <span class="rank">{{ $index + 1 }}</span>
        </template>
      </el-table-column>
      
      <el-table-column label="成员">
        <template #default="{ row }">
          <span class="member-text">{{ row.member }}</span>
        </template>
      </el-table-column>
      
      <el-table-column label="分数" width="150">
        <template #default="{ row }">
          <el-input-number
            v-if="editing"
            :model-value="row.score"
            @update:model-value="(val: number | undefined) => val !== undefined && handleUpdateScore(row.id, val)"
            size="small"
            :controls="false"
          />
          <span v-else class="score">{{ row.score }}</span>
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
    <div class="zset-footer">
      共 {{ displayList.length }} 个成员
    </div>
  </div>
</template>

<style scoped>
.zset-viewer {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.rank {
  color: #909399;
  font-family: monospace;
}

.member-text {
  font-family: monospace;
  word-break: break-all;
}

.score {
  font-family: monospace;
  color: #e6a23c;
}

.zset-footer {
  padding: 8px 0;
  text-align: right;
  font-size: 12px;
  color: #909399;
}

:deep(.el-table) {
  flex: 1;
}
</style>
