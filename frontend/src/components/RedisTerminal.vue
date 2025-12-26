<script setup lang="ts">
import { ref, nextTick, watch, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useConnectionStore } from '@/stores/connection'
import { useTabsStore } from '@/stores/tabs'
import type { CommandHistoryItem } from '@/types/tab'

const props = defineProps<{
  tabId?: string
}>()

const connectionStore = useConnectionStore()
const tabsStore = useTabsStore()

// Get current tab
const currentTab = computed(() => {
  if (props.tabId) {
    return tabsStore.getTab(props.tabId)
  }
  return tabsStore.activeTab
})

// Tab-specific command history (synced with tab state)
const history = computed({
  get: () => currentTab.value?.terminalState.history ?? [],
  set: (value: CommandHistoryItem[]) => {
    if (currentTab.value) {
      tabsStore.updateTerminalState(currentTab.value.id, { history: value })
    }
  }
})

// Tab-specific input value
const currentCommand = computed({
  get: () => currentTab.value?.terminalState.inputValue ?? '',
  set: (value: string) => {
    if (currentTab.value) {
      tabsStore.updateTerminalState(currentTab.value.id, { inputValue: value })
    }
  }
})

// Local command history for arrow key navigation (not persisted per tab)
const commandHistory = ref<string[]>([])
const historyIndex = ref(-1)
const terminalRef = ref<HTMLDivElement>()
const inputRef = ref<HTMLInputElement>()
const executing = ref(false)

// 当前连接信息
const connectionId = (): string | null => connectionStore.activeConnectionId
const currentDb = (): number => connectionStore.activeConnection?.currentDb ?? 0
const connectionName = (): string => connectionStore.activeConnection?.config.name ?? 'redis'

// 执行命令
async function executeCommand() {
  const cmd = currentCommand.value.trim()
  if (!cmd) return
  
  const connId = connectionId()
  if (!connId) {
    ElMessage.warning('请先连接 Redis')
    return
  }

  // 添加到本地历史（用于箭头键导航）
  if (commandHistory.value[0] !== cmd) {
    commandHistory.value.unshift(cmd)
    if (commandHistory.value.length > 100) {
      commandHistory.value.pop()
    }
  }
  historyIndex.value = -1

  executing.value = true
  const startTime = Date.now()

  try {
    const response = await fetch(`/api/connections/${connId}/command`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ db: currentDb(), command: cmd })
    })
    
    const data = await response.json()
    const duration = Date.now() - startTime

    let resultText = data.success ? (data.result || '(nil)') : `(error) ${data.error}`
    const isErrorResult = !data.success || resultText.startsWith('(error)')

    // 显示执行时间
    if (data.success && !isErrorResult) {
      resultText += `\n(${duration}ms)`
    }

    // Add to tab-specific history
    if (currentTab.value) {
      tabsStore.addCommandToHistory(currentTab.value.id, {
        command: cmd,
        result: resultText,
        isError: isErrorResult,
        timestamp: new Date()
      })
    }
  } catch (err: any) {
    // Add error to tab-specific history
    if (currentTab.value) {
      tabsStore.addCommandToHistory(currentTab.value.id, {
        command: cmd,
        result: `(error) ${err.message || '执行失败'}`,
        isError: true,
        timestamp: new Date()
      })
    }
  } finally {
    executing.value = false
    currentCommand.value = ''
    scrollToBottom()
    // 重新聚焦输入框
    nextTick(() => {
      focusInput()
    })
  }
}

// 处理键盘事件
function handleKeyDown(e: KeyboardEvent) {
  if (e.key === 'ArrowUp') {
    e.preventDefault()
    if (historyIndex.value < commandHistory.value.length - 1) {
      historyIndex.value++
      currentCommand.value = commandHistory.value[historyIndex.value] || ''
    }
  } else if (e.key === 'ArrowDown') {
    e.preventDefault()
    if (historyIndex.value > 0) {
      historyIndex.value--
      currentCommand.value = commandHistory.value[historyIndex.value] || ''
    } else if (historyIndex.value === 0) {
      historyIndex.value = -1
      currentCommand.value = ''
    }
  } else if (e.key === 'Tab') {
    e.preventDefault()
    // 简单的命令补全
    autoComplete()
  } else if (e.ctrlKey && e.key === 'l') {
    e.preventDefault()
    clearHistory()
  }
}

// 命令补全
const REDIS_COMMANDS = [
  'GET', 'SET', 'DEL', 'EXISTS', 'EXPIRE', 'TTL', 'PTTL', 'PERSIST',
  'KEYS', 'SCAN', 'TYPE', 'RENAME', 'RENAMENX',
  'APPEND', 'STRLEN', 'GETRANGE', 'SETRANGE', 'INCR', 'DECR', 'INCRBY', 'DECRBY',
  'LPUSH', 'RPUSH', 'LPOP', 'RPOP', 'LRANGE', 'LLEN', 'LINDEX', 'LSET', 'LREM',
  'HGET', 'HSET', 'HDEL', 'HGETALL', 'HKEYS', 'HVALS', 'HLEN', 'HEXISTS',
  'SADD', 'SREM', 'SMEMBERS', 'SISMEMBER', 'SCARD', 'SUNION', 'SINTER', 'SDIFF',
  'ZADD', 'ZREM', 'ZRANGE', 'ZREVRANGE', 'ZSCORE', 'ZRANK', 'ZCARD', 'ZCOUNT',
  'PING', 'ECHO', 'INFO', 'DBSIZE', 'FLUSHDB', 'FLUSHALL', 'SELECT',
  'MULTI', 'EXEC', 'DISCARD', 'WATCH', 'UNWATCH',
  'PUBLISH', 'SUBSCRIBE', 'UNSUBSCRIBE', 'PSUBSCRIBE', 'PUNSUBSCRIBE',
  'CONFIG', 'CLIENT', 'DEBUG', 'MEMORY', 'SLOWLOG'
]

function autoComplete() {
  const input = currentCommand.value.toUpperCase()
  if (!input) return

  const parts = input.split(' ')
  const lastPart = parts[parts.length - 1] || ''
  
  const matches = REDIS_COMMANDS.filter(cmd => cmd.startsWith(lastPart))
  if (matches.length === 1 && matches[0]) {
    parts[parts.length - 1] = matches[0]
    currentCommand.value = parts.join(' ') + ' '
  } else if (matches.length > 1) {
    // 显示可能的补全 - add to tab history
    if (currentTab.value) {
      tabsStore.addCommandToHistory(currentTab.value.id, {
        command: '',
        result: matches.join('  '),
        isError: false,
        timestamp: new Date()
      })
    }
    scrollToBottom()
  }
}

// 清空历史
function clearHistory() {
  if (currentTab.value) {
    tabsStore.clearTerminalHistory(currentTab.value.id)
  }
}

// 滚动到底部
function scrollToBottom() {
  nextTick(() => {
    if (terminalRef.value) {
      terminalRef.value.scrollTop = terminalRef.value.scrollHeight
    }
  })
}

// 聚焦输入框
function focusInput() {
  inputRef.value?.focus()
}

// Watch for tab changes to scroll to bottom
watch(() => currentTab.value?.id, () => {
  nextTick(() => {
    scrollToBottom()
    focusInput()
  })
})

// Watch history changes to scroll to bottom
watch(() => history.value.length, () => {
  scrollToBottom()
})

onMounted(() => {
  focusInput()
  scrollToBottom()
})
</script>

<template>
  <div class="redis-terminal" @click="focusInput">
    <div class="terminal-header">
      <span class="terminal-title">Redis 终端</span>
      <span v-if="connectionName()" class="connection-info">
        {{ connectionName() }} - db{{ currentDb() }}
      </span>
      <div class="terminal-actions">
        <el-button size="small" text @click="clearHistory">清空</el-button>
      </div>
    </div>
    
    <div ref="terminalRef" class="terminal-content">
      <!-- 欢迎信息 -->
      <div v-if="history.length === 0" class="welcome-message">
        <div>欢迎使用 Redis 终端</div>
        <div class="tips">
          <div>• 输入 Redis 命令并按 Enter 执行</div>
          <div>• 使用 ↑/↓ 浏览历史命令</div>
          <div>• 使用 Tab 自动补全命令</div>
          <div>• 使用 Ctrl+L 清空屏幕</div>
        </div>
      </div>
      
      <!-- 命令历史 -->
      <div v-for="(item, index) in history" :key="index" class="history-item">
        <div v-if="item.command" class="command-line">
          <span class="prompt">{{ connectionName() }}:{{ currentDb() }}></span>
          <span class="command">{{ item.command }}</span>
        </div>
        <pre class="result" :class="{ error: item.isError }">{{ item.result }}</pre>
      </div>
      
      <!-- 当前输入 -->
      <div class="input-line">
        <span class="prompt">{{ connectionName() }}:{{ currentDb() }}></span>
        <input
          ref="inputRef"
          v-model="currentCommand"
          type="text"
          class="command-input"
          :disabled="executing || !connectionId()"
          placeholder="输入 Redis 命令..."
          @keydown="handleKeyDown"
          @keyup.enter="executeCommand"
        />
        <span v-if="executing" class="executing">执行中...</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.redis-terminal {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 13px;
  border-radius: 4px;
  overflow: hidden;
}

.terminal-header {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: #2d2d2d;
  border-bottom: 1px solid #3c3c3c;
}

.terminal-title {
  font-weight: 600;
  color: #e0e0e0;
}

.connection-info {
  margin-left: 12px;
  color: #6a9955;
  font-size: 12px;
}

.terminal-actions {
  margin-left: auto;
}

.terminal-actions .el-button {
  color: #909399;
}

.terminal-content {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

.welcome-message {
  color: #6a9955;
  margin-bottom: 16px;
}

.tips {
  margin-top: 8px;
  color: #808080;
  font-size: 12px;
}

.tips div {
  margin: 4px 0;
}

.history-item {
  margin-bottom: 8px;
}

.command-line {
  display: flex;
  align-items: center;
  gap: 8px;
}

.prompt {
  color: #569cd6;
  white-space: nowrap;
}

.command {
  color: #ce9178;
}

.result {
  margin: 4px 0 0 0;
  padding: 0;
  white-space: pre-wrap;
  word-break: break-all;
  color: #d4d4d4;
  font-family: inherit;
  font-size: inherit;
  line-height: 1.5;
}

.result.error {
  color: #f14c4c;
}

.input-line {
  display: flex;
  align-items: center;
  gap: 8px;
}

.command-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: #ce9178;
  font-family: inherit;
  font-size: inherit;
  caret-color: #d4d4d4;
}

.command-input::placeholder {
  color: #5a5a5a;
}

.command-input:disabled {
  opacity: 0.5;
}

.executing {
  color: #dcdcaa;
  font-size: 12px;
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0.3; }
}

/* 滚动条样式 */
.terminal-content::-webkit-scrollbar {
  width: 8px;
}

.terminal-content::-webkit-scrollbar-track {
  background: #1e1e1e;
}

.terminal-content::-webkit-scrollbar-thumb {
  background: #424242;
  border-radius: 4px;
}

.terminal-content::-webkit-scrollbar-thumb:hover {
  background: #4f4f4f;
}
</style>
