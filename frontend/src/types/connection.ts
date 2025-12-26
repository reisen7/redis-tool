// Redis 连接模式
export type ConnectionMode = 'standalone' | 'cluster' | 'sentinel'

// SSH 认证类型
export type SSHAuthType = 'password' | 'privateKey'

// SSH 配置
export interface SSHConfig {
  enabled: boolean
  host: string
  port: number
  username: string
  authType: SSHAuthType
  password: string
  privateKey: string
  privateKeyPassphrase: string
}

// TLS 配置
export interface TLSConfig {
  enabled: boolean
  caCert: string
  clientCert: string
  clientKey: string
  skipVerify: boolean
}

// 连接分组
export interface ConnectionGroup {
  id: string
  name: string
  icon: string
  color: string
  sortOrder: number
  createdAt?: string
  updatedAt?: string
}

// 连接配置
export interface ConnectionConfig {
  id: string
  groupId: string
  name: string
  mode: ConnectionMode
  host: string
  port: number
  username: string
  password: string
  sentinelMasterName: string
  sentinelNodes: string[]
  clusterNodes: string[]
  ssh: SSHConfig
  tls: TLSConfig
  sortOrder?: number
  createdAt?: string
  updatedAt?: string
}

// 数据库信息
export interface DatabaseInfo {
  index: number
  alias: string
  keyCount: number
}

// 连接状态
export type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error'

// 连接实例（运行时）
export interface ConnectionInstance {
  config: ConnectionConfig
  status: ConnectionStatus
  databases: DatabaseInfo[]
  currentDb: number
  error?: string
}

// 创建默认 SSH 配置
export function createDefaultSSHConfig(): SSHConfig {
  return {
    enabled: false,
    host: '',
    port: 22,
    username: '',
    authType: 'password',
    password: '',
    privateKey: '',
    privateKeyPassphrase: ''
  }
}

// 创建默认 TLS 配置
export function createDefaultTLSConfig(): TLSConfig {
  return {
    enabled: false,
    caCert: '',
    clientCert: '',
    clientKey: '',
    skipVerify: false
  }
}

// 创建默认连接配置
export function createDefaultConnection(): ConnectionConfig {
  return {
    id: '',
    groupId: '',
    name: '',
    mode: 'standalone',
    host: '127.0.0.1',
    port: 6379,
    username: '',
    password: '',
    sentinelMasterName: '',
    sentinelNodes: [],
    clusterNodes: [],
    ssh: createDefaultSSHConfig(),
    tls: createDefaultTLSConfig()
  }
}

// 创建默认分组
export function createDefaultGroup(): ConnectionGroup {
  return {
    id: '',
    name: '',
    icon: 'folder',
    color: '#409eff',
    sortOrder: 0
  }
}
