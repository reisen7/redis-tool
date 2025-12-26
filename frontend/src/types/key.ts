// Redis 数据类型
export type RedisDataType = 'string' | 'list' | 'hash' | 'set' | 'zset' | 'none'

// Key 信息
export interface KeyInfo {
  key: string
  type: RedisDataType
  ttl: number
  size?: number
}

// Key 值（通用）
export interface KeyValue {
  key: string
  type: RedisDataType
  value: any
  ttl: number
  size?: number
  format?: StringFormat
}

// String 格式类型
export type StringFormat = 'text' | 'json' | 'xml' | 'yaml' | 'toml' | 'properties' | 'number' | 'binary'

// 格式选项
export const STRING_FORMAT_OPTIONS: { label: string; value: StringFormat }[] = [
  { label: '原始文本', value: 'text' },
  { label: 'JSON', value: 'json' },
  { label: 'XML', value: 'xml' },
  { label: 'YAML', value: 'yaml' },
  { label: 'TOML', value: 'toml' },
  { label: 'Properties', value: 'properties' }
]

// Hash 字段
export interface HashField {
  field: string
  value: string
}

// ZSet 成员
export interface ZSetMember {
  member: string
  score: number
}

// 搜索类型
export type SearchType = 'keys' | 'scan' | 'regex'

// 搜索配置
export interface SearchConfig {
  searchType: SearchType
  pattern: string
  scanCount: number
  regexPattern: string
}

// Key 列表响应
export interface KeysResponse {
  success: boolean
  keys: KeyInfo[]
  cursor: number
  total?: number
  error?: string
}

// Key 值响应
export interface KeyValueResponse {
  success: boolean
  data?: KeyValue
  error?: string
}

// 创建 Key 请求
export interface CreateKeyRequest {
  key: string
  type: RedisDataType
  value: any
  ttl: number
}

// TTL 快捷选项
export const TTL_PRESETS = [
  { label: '1 分钟', value: 60 },
  { label: '5 分钟', value: 300 },
  { label: '30 分钟', value: 1800 },
  { label: '1 小时', value: 3600 },
  { label: '6 小时', value: 21600 },
  { label: '12 小时', value: 43200 },
  { label: '1 天', value: 86400 },
  { label: '7 天', value: 604800 },
  { label: '30 天', value: 2592000 },
  { label: '永不过期', value: -1 }
]

// 检测 String 内容格式
export function detectStringFormat(value: string): StringFormat {
  if (!value || typeof value !== 'string') return 'text'
  
  const trimmed = value.trim()
  
  // 检测 JSON
  if ((trimmed.startsWith('{') && trimmed.endsWith('}')) ||
      (trimmed.startsWith('[') && trimmed.endsWith(']'))) {
    try {
      JSON.parse(trimmed)
      return 'json'
    } catch {}
  }
  
  // 检测 XML (包括 HTML)
  if (trimmed.startsWith('<?xml') || 
      trimmed.startsWith('<!DOCTYPE') ||
      trimmed.startsWith('<html') ||
      (trimmed.startsWith('<') && trimmed.endsWith('>') && /<\/\w+>/.test(trimmed))) {
    return 'xml'
  }
  
  // 检测 TOML (优先于 properties，因为 TOML 有 [section] 标记)
  if (/^\[[\w.-]+\]\s*$/m.test(trimmed) && /^\w+\s*=\s*.+$/m.test(trimmed)) {
    return 'toml'
  }
  
  // 检测 YAML (优先于 properties)
  // YAML 特征：冒号后有空格，或者有缩进的列表
  if (trimmed.includes(':')) {
    const lines = trimmed.split('\n').filter(l => l.trim() && !l.trim().startsWith('#'))
    // YAML: key: value 格式，冒号后有空格或换行
    const yamlPattern = /^[\w-]+:\s|^[\w-]+:$/
    const hasYamlStyle = lines.some(line => yamlPattern.test(line.trim()))
    // YAML 列表: - item
    const hasYamlList = lines.some(line => /^\s*-\s+/.test(line))
    
    if (hasYamlStyle || hasYamlList) {
      return 'yaml'
    }
  }
  
  // 检测 Properties (key=value 或 key:value 格式，无 section)
  if (trimmed.includes('\n') && !trimmed.includes('[')) {
    const lines = trimmed.split('\n').filter(l => l.trim() && !l.trim().startsWith('#') && !l.trim().startsWith('!'))
    if (lines.length > 0 && lines.every(l => /^[\w.]+\s*[=:]\s*/.test(l.trim()))) {
      return 'properties'
    }
  }
  
  // 检测数字
  if (/^-?\d+(\.\d+)?$/.test(trimmed)) {
    return 'number'
  }
  
  return 'text'
}

// 格式化 TTL 显示
export function formatTTL(ttl: number): string {
  if (ttl === -1) return '永不过期'
  if (ttl === -2) return '已过期'
  if (ttl <= 0) return '已过期'
  
  const days = Math.floor(ttl / 86400)
  const hours = Math.floor((ttl % 86400) / 3600)
  const minutes = Math.floor((ttl % 3600) / 60)
  const seconds = ttl % 60
  
  const parts: string[] = []
  if (days > 0) parts.push(`${days}天`)
  if (hours > 0) parts.push(`${hours}小时`)
  if (minutes > 0) parts.push(`${minutes}分`)
  if (seconds > 0 || parts.length === 0) parts.push(`${seconds}秒`)
  
  return parts.join(' ')
}

// 获取类型图标
export function getTypeIcon(type: RedisDataType): string {
  const icons: Record<RedisDataType, string> = {
    string: '🔤',
    list: '📋',
    hash: '#️⃣',
    set: '📦',
    zset: '📊',
    none: '❓'
  }
  return icons[type] || '❓'
}

// 获取类型颜色
export function getTypeColor(type: RedisDataType): string {
  const colors: Record<RedisDataType, string> = {
    string: '#67c23a',
    list: '#409eff',
    hash: '#e6a23c',
    set: '#f56c6c',
    zset: '#909399',
    none: '#c0c4cc'
  }
  return colors[type] || '#c0c4cc'
}
