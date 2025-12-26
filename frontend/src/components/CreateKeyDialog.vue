<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Delete } from '@element-plus/icons-vue'
import type { RedisDataType, CreateKeyRequest, StringFormat } from '@/types/key'
import { TTL_PRESETS, STRING_FORMAT_OPTIONS } from '@/types/key'
import { useConnectionStore } from '@/stores/connection'
import { useKeysStore } from '@/stores/keys'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const connectionStore = useConnectionStore()
const keysStore = useKeysStore()

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})

// 表单数据
const form = reactive({
  key: '',
  type: 'string' as RedisDataType,
  ttl: -1,
  neverExpire: true
})

// String 格式
const stringFormat = ref<StringFormat>('text')

// 各类型的值
const stringValue = ref('')
const listValue = ref<string[]>([''])
const hashValue = ref<{ field: string; value: string }[]>([{ field: '', value: '' }])
const setValue = ref<string[]>([''])
const zsetValue = ref<{ member: string; score: number }[]>([{ member: '', score: 0 }])

// 类型选项
const typeOptions: { label: string; value: RedisDataType }[] = [
  { label: 'String', value: 'string' },
  { label: 'List', value: 'list' },
  { label: 'Hash', value: 'hash' },
  { label: 'Set', value: 'set' },
  { label: 'ZSet', value: 'zset' }
]

// 重置表单
function resetForm() {
  form.key = ''
  form.type = 'string'
  form.ttl = -1
  form.neverExpire = true
  stringFormat.value = 'text'
  stringValue.value = ''
  listValue.value = ['']
  hashValue.value = [{ field: '', value: '' }]
  setValue.value = ['']
  zsetValue.value = [{ member: '', score: 0 }]
}

watch(dialogVisible, (visible) => {
  if (visible) {
    resetForm()
  }
})

// TTL 处理
watch(() => form.neverExpire, (neverExpire) => {
  if (neverExpire) {
    form.ttl = -1
  } else if (form.ttl < 0) {
    form.ttl = 3600
  }
})

// List 操作
function addListItem() {
  listValue.value.push('')
}
function removeListItem(index: number) {
  listValue.value.splice(index, 1)
}

// Hash 操作
function addHashItem() {
  hashValue.value.push({ field: '', value: '' })
}
function removeHashItem(index: number) {
  hashValue.value.splice(index, 1)
}

// Set 操作
function addSetItem() {
  setValue.value.push('')
}
function removeSetItem(index: number) {
  setValue.value.splice(index, 1)
}

// ZSet 操作
function addZSetItem() {
  zsetValue.value.push({ member: '', score: 0 })
}
function removeZSetItem(index: number) {
  zsetValue.value.splice(index, 1)
}

// String 格式化相关
const canFormatString = computed(() => ['json', 'xml'].includes(stringFormat.value))

function getStringPlaceholder(): string {
  const placeholders: Record<StringFormat, string> = {
    text: '输入文本内容',
    json: '{\n  "key": "value"\n}',
    xml: '<?xml version="1.0"?>\n<root>\n  <item>value</item>\n</root>',
    yaml: 'key: value\nlist:\n  - item1\n  - item2',
    toml: '[section]\nkey = "value"',
    properties: 'key=value\nkey2=value2',
    number: '123',
    binary: ''
  }
  return placeholders[stringFormat.value] || '输入内容'
}

// 获取格式对应的样式类
function getFormatClass(): string {
  return `format-${stringFormat.value}`
}

function formatStringValue() {
  try {
    switch (stringFormat.value) {
      case 'json':
        stringValue.value = JSON.stringify(JSON.parse(stringValue.value), null, 2)
        break
      case 'xml':
        stringValue.value = formatXml(stringValue.value)
        break
    }
  } catch (e) {
    ElMessage.warning('格式化失败，请检查内容格式')
  }
}

function minifyStringValue() {
  try {
    switch (stringFormat.value) {
      case 'json':
        stringValue.value = JSON.stringify(JSON.parse(stringValue.value))
        break
      case 'xml':
        stringValue.value = stringValue.value.replace(/>\s+</g, '><').trim()
        break
    }
  } catch (e) {
    ElMessage.warning('压缩失败，请检查内容格式')
  }
}

function formatXml(xml: string): string {
  try {
    let formatted = ''
    let indent = ''
    const tab = '  '
    
    xml = xml.replace(/>\s*</g, '><')
    const nodes = xml.split(/(<[^>]+>)/g).filter(n => n.trim())
    
    for (const node of nodes) {
      if (node.startsWith('</')) {
        indent = indent.substring(tab.length)
        formatted += indent + node + '\n'
      } else if (node.startsWith('<') && !node.startsWith('<?') && !node.startsWith('<!') && !node.endsWith('/>')) {
        formatted += indent + node + '\n'
        if (!node.includes('</')) {
          indent += tab
        }
      } else if (node.startsWith('<')) {
        formatted += indent + node + '\n'
      } else {
        formatted += indent + node.trim() + '\n'
      }
    }
    
    return formatted.trim()
  } catch {
    return xml
  }
}

// 构建值
function buildValue(): any {
  switch (form.type) {
    case 'string':
      return stringValue.value
    case 'list':
      return listValue.value.filter(v => v.trim())
    case 'hash':
      const hash: Record<string, string> = {}
      hashValue.value.forEach(item => {
        if (item.field.trim()) {
          hash[item.field] = item.value
        }
      })
      return hash
    case 'set':
      return setValue.value.filter(v => v.trim())
    case 'zset':
      return zsetValue.value
        .filter(item => item.member.trim())
        .map(item => ({ member: item.member, score: item.score }))
    default:
      return null
  }
}

// 提交
const saving = ref(false)

async function handleSubmit() {
  if (!form.key.trim()) {
    ElMessage.warning('请输入 Key 名称')
    return
  }

  const connectionId = connectionStore.activeConnectionId
  const db = connectionStore.activeConnection?.currentDb ?? 0
  
  if (!connectionId) {
    ElMessage.warning('请先选择一个连接')
    return
  }

  const request: CreateKeyRequest = {
    key: form.key,
    type: form.type,
    value: buildValue(),
    ttl: form.neverExpire ? -1 : form.ttl
  }

  saving.value = true
  try {
    const result = await keysStore.createKey(connectionId, db, request)
    if (result.success) {
      ElMessage.success('创建成功')
      dialogVisible.value = false
    } else {
      ElMessage.error(result.error || '创建失败')
    }
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <el-dialog
    v-model="dialogVisible"
    title="新增 Key"
    width="650px"
    destroy-on-close
  >
    <el-form :model="form" label-width="100px">
      <el-form-item label="Key 名称" required>
        <el-input v-model="form.key" placeholder="请输入 Key 名称" />
      </el-form-item>

      <el-form-item label="数据类型">
        <el-select v-model="form.type" style="width: 100%">
          <el-option
            v-for="opt in typeOptions"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="过期时间">
        <div class="ttl-row">
          <el-checkbox v-model="form.neverExpire">永不过期</el-checkbox>
          <template v-if="!form.neverExpire">
            <el-input-number
              v-model="form.ttl"
              :min="1"
              style="width: 150px; margin-left: 16px"
            />
            <span style="margin-left: 8px">秒</span>
          </template>
        </div>
        <div v-if="!form.neverExpire" class="ttl-presets">
          <el-button
            v-for="preset in TTL_PRESETS.filter((p: { label: string; value: number }) => p.value > 0)"
            :key="preset.value"
            size="small"
            :type="form.ttl === preset.value ? 'primary' : 'default'"
            @click="form.ttl = preset.value"
          >
            {{ preset.label }}
          </el-button>
        </div>
      </el-form-item>

      <!-- String 值 -->
      <el-form-item v-if="form.type === 'string'" label="内容格式">
        <el-select v-model="stringFormat" style="width: 100%">
          <el-option
            v-for="opt in STRING_FORMAT_OPTIONS"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
      </el-form-item>
      
      <el-form-item v-if="form.type === 'string'" label="值">
        <div class="code-editor-wrapper">
          <textarea
            v-model="stringValue"
            :placeholder="getStringPlaceholder()"
            class="code-editor"
            :class="getFormatClass()"
            rows="10"
          ></textarea>
        </div>
        <div class="format-actions" v-if="canFormatString">
          <el-button size="small" @click="formatStringValue">格式化</el-button>
          <el-button size="small" @click="minifyStringValue">压缩</el-button>
        </div>
      </el-form-item>

      <!-- List 值 -->
      <el-form-item v-if="form.type === 'list'" label="列表元素">
        <div class="list-items">
          <div v-for="(_, index) in listValue" :key="index" class="list-item">
            <el-input v-model="listValue[index]" placeholder="元素值" />
            <el-button
              type="danger"
              :icon="Delete"
              link
              :disabled="listValue.length <= 1"
              @click="removeListItem(index)"
            />
          </div>
          <el-button type="primary" :icon="Plus" link @click="addListItem">
            添加元素
          </el-button>
        </div>
      </el-form-item>

      <!-- Hash 值 -->
      <el-form-item v-if="form.type === 'hash'" label="Hash 字段">
        <div class="hash-items">
          <div v-for="(item, index) in hashValue" :key="index" class="hash-item">
            <el-input v-model="item.field" placeholder="字段名" style="width: 150px" />
            <el-input v-model="item.value" placeholder="值" />
            <el-button
              type="danger"
              :icon="Delete"
              link
              :disabled="hashValue.length <= 1"
              @click="removeHashItem(index)"
            />
          </div>
          <el-button type="primary" :icon="Plus" link @click="addHashItem">
            添加字段
          </el-button>
        </div>
      </el-form-item>

      <!-- Set 值 -->
      <el-form-item v-if="form.type === 'set'" label="Set 成员">
        <div class="set-items">
          <div v-for="(_, index) in setValue" :key="index" class="set-item">
            <el-input v-model="setValue[index]" placeholder="成员值" />
            <el-button
              type="danger"
              :icon="Delete"
              link
              :disabled="setValue.length <= 1"
              @click="removeSetItem(index)"
            />
          </div>
          <el-button type="primary" :icon="Plus" link @click="addSetItem">
            添加成员
          </el-button>
        </div>
      </el-form-item>

      <!-- ZSet 值 -->
      <el-form-item v-if="form.type === 'zset'" label="ZSet 成员">
        <div class="zset-items">
          <div v-for="(item, index) in zsetValue" :key="index" class="zset-item">
            <el-input v-model="item.member" placeholder="成员" style="flex: 1" />
            <el-input-number v-model="item.score" placeholder="分数" style="width: 120px" />
            <el-button
              type="danger"
              :icon="Delete"
              link
              :disabled="zsetValue.length <= 1"
              @click="removeZSetItem(index)"
            />
          </div>
          <el-button type="primary" :icon="Plus" link @click="addZSetItem">
            添加成员
          </el-button>
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="handleSubmit">
        创建
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.ttl-row {
  display: flex;
  align-items: center;
}

.ttl-presets {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.list-items,
.hash-items,
.set-items,
.zset-items {
  width: 100%;
}

.list-item,
.hash-item,
.set-item,
.zset-item {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.list-item .el-input,
.set-item .el-input {
  flex: 1;
}

.hash-item .el-input:last-of-type {
  flex: 1;
}

/* 代码编辑器样式 */
.code-editor-wrapper {
  width: 100%;
}

.code-editor {
  width: 100%;
  min-height: 200px;
  padding: 12px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.6;
  background: #1e1e1e;
  color: #d4d4d4;
  border: 1px solid #3c3c3c;
  border-radius: 4px;
  resize: vertical;
  outline: none;
}

.code-editor:focus {
  border-color: #409eff;
}

.code-editor::placeholder {
  color: #6a6a6a;
}

/* 格式高亮颜色 */
.code-editor.format-json {
  color: #9cdcfe;
}

.code-editor.format-xml {
  color: #e06c75;
}

.code-editor.format-yaml {
  color: #98c379;
}

.code-editor.format-toml {
  color: #e5c07b;
}

.code-editor.format-properties {
  color: #c678dd;
}

.format-actions {
  margin-top: 8px;
  display: flex;
  gap: 8px;
}

.form-tip {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}
</style>
