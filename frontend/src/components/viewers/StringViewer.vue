<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { detectStringFormat, STRING_FORMAT_OPTIONS, type StringFormat } from '@/types/key'

const props = defineProps<{
  value: string
  editing: boolean
}>()

const emit = defineEmits<{
  (e: 'change', value: string): void
}>()

// 检测格式
const detectedFormat = computed(() => detectStringFormat(props.value))

// 当前显示格式
const displayFormat = ref<StringFormat>('text')

// 编辑时选择的格式
const editFormat = ref<StringFormat>('text')

// 格式化后的值
const formattedValue = computed(() => {
  if (!props.value) return ''
  
  try {
    switch (displayFormat.value) {
      case 'json':
        return formatJson(props.value)
      case 'xml':
        return formatXml(props.value)
      case 'yaml':
      case 'toml':
      case 'properties':
        return props.value
      default:
        return props.value
    }
  } catch {
    return props.value
  }
})

// 编辑值
const editText = ref('')

watch(() => props.value, (val) => {
  editText.value = val || ''
  const detected = detectStringFormat(val)
  displayFormat.value = detected
  editFormat.value = detected
}, { immediate: true })

watch(() => props.editing, (editing) => {
  if (editing) {
    editText.value = props.value || ''
  }
})

// JSON 格式化
function formatJson(value: string): string {
  try {
    return JSON.stringify(JSON.parse(value), null, 2)
  } catch {
    return value
  }
}

// XML 格式化
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

// 获取格式对应的语法高亮类名
function getFormatClass(format: StringFormat): string {
  return `format-${format}`
}

// 输入变化
function handleInput(value: string) {
  editText.value = value
  emit('change', value)
}

// 格式化编辑内容
function handleFormat() {
  try {
    switch (editFormat.value) {
      case 'json':
        editText.value = formatJson(editText.value)
        break
      case 'xml':
        editText.value = formatXml(editText.value)
        break
    }
    emit('change', editText.value)
  } catch (e) {
    // 格式化失败，保持原样
  }
}

// 压缩内容
function handleMinify() {
  try {
    switch (editFormat.value) {
      case 'json':
        editText.value = JSON.stringify(JSON.parse(editText.value))
        break
      case 'xml':
        editText.value = editText.value.replace(/>\s+</g, '><').trim()
        break
    }
    emit('change', editText.value)
  } catch (e) {
    // 压缩失败，保持原样
  }
}

// 可格式化的类型
const canFormat = computed(() => ['json', 'xml'].includes(editFormat.value))

// 获取占位符
function getPlaceholder(format: StringFormat): string {
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
  return placeholders[format] || '输入内容'
}
</script>

<template>
  <div class="string-viewer">
    <!-- 查看模式工具栏 -->
    <div class="viewer-toolbar" v-if="!editing">
      <span class="format-label">
        检测格式: <el-tag size="small" type="info">{{ detectedFormat.toUpperCase() }}</el-tag>
      </span>
      <el-radio-group v-model="displayFormat" size="small">
        <el-radio-button value="text">原始</el-radio-button>
        <el-radio-button 
          v-for="opt in STRING_FORMAT_OPTIONS.filter(o => o.value !== 'text')" 
          :key="opt.value"
          :value="opt.value" 
          :disabled="detectedFormat !== opt.value"
        >
          {{ opt.label }}
        </el-radio-button>
      </el-radio-group>
    </div>

    <!-- 编辑模式工具栏 -->
    <div class="viewer-toolbar" v-if="editing">
      <span class="format-label">输入格式:</span>
      <el-select v-model="editFormat" size="small" style="width: 140px">
        <el-option
          v-for="opt in STRING_FORMAT_OPTIONS"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
      <el-button 
        v-if="canFormat" 
        size="small" 
        @click="handleFormat"
      >
        格式化
      </el-button>
      <el-button 
        v-if="canFormat" 
        size="small" 
        @click="handleMinify"
      >
        压缩
      </el-button>
    </div>

    <!-- 查看模式 -->
    <div v-if="!editing" class="value-display">
      <pre class="value-content" :class="getFormatClass(displayFormat)">{{ formattedValue }}</pre>
    </div>

    <!-- 编辑模式 -->
    <div v-else class="value-edit">
      <el-input
        type="textarea"
        :model-value="editText"
        @update:model-value="handleInput"
        :rows="15"
        :placeholder="getPlaceholder(editFormat)"
        class="code-textarea"
      />
    </div>
  </div>
</template>

<style scoped>
.string-viewer {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.viewer-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.format-label {
  font-size: 12px;
  color: #909399;
}

.value-display {
  flex: 1;
  overflow: auto;
  background: #1e1e1e;
  border-radius: 4px;
  padding: 12px;
}

.value-content {
  margin: 0;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  color: #d4d4d4;
}

/* 语法高亮颜色 */
.format-json {
  color: #9cdcfe;
}

.format-xml {
  color: #e06c75;
}

.format-yaml {
  color: #98c379;
}

.format-toml {
  color: #e5c07b;
}

.format-properties {
  color: #c678dd;
}

.format-text {
  color: #d4d4d4;
}

.value-edit {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.value-edit :deep(.el-textarea) {
  flex: 1;
}

.value-edit :deep(.el-textarea__inner) {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 13px;
  line-height: 1.6;
  height: 100% !important;
  resize: none;
  background: #1e1e1e;
  color: #d4d4d4;
  border-color: #3c3c3c;
}

.value-edit :deep(.el-textarea__inner:focus) {
  border-color: #409eff;
}

.value-edit :deep(.el-textarea__inner::placeholder) {
  color: #6a6a6a;
}
</style>
