import http from './http'
import type { KeyInfo, KeyValue, CreateKeyRequest, SearchConfig } from '@/types/key'

// Key 管理 API
export const keysApi = {
  // 获取 Key 列表
  list(connectionId: string, db: number, pattern: string = '*', cursor: number = 0, count: number = 100) {
    return http.get<{
      success: boolean
      keys: KeyInfo[]
      cursor: number
      error?: string
    }>(`/connections/${connectionId}/keys`, {
      params: { db, pattern, cursor, count }
    })
  },

  // 扫描 Key（支持正则）
  scan(connectionId: string, db: number, config: SearchConfig) {
    return http.post<{
      success: boolean
      keys: KeyInfo[]
      cursor: number
      error?: string
    }>(`/connections/${connectionId}/keys/scan`, {
      db,
      ...config
    })
  },

  // 获取 Key 详情
  get(connectionId: string, db: number, key: string) {
    return http.get<{
      success: boolean
      data: KeyValue
      error?: string
    }>(`/connections/${connectionId}/keys/${encodeURIComponent(key)}`, {
      params: { db }
    })
  },

  // 创建 Key
  create(connectionId: string, db: number, request: CreateKeyRequest) {
    return http.post<{ success: boolean; error?: string }>(
      `/connections/${connectionId}/keys`,
      { db, ...request }
    )
  },

  // 更新 Key 值
  update(connectionId: string, db: number, key: string, value: any) {
    return http.put<{ success: boolean; error?: string }>(
      `/connections/${connectionId}/keys/${encodeURIComponent(key)}`,
      { db, value }
    )
  },

  // 删除 Key
  delete(connectionId: string, db: number, keys: string[]) {
    return http.delete<{ success: boolean; deleted: number; error?: string }>(
      `/connections/${connectionId}/keys`,
      { data: { db, keys } }
    )
  },

  // 设置 TTL
  setTTL(connectionId: string, db: number, key: string, ttl: number) {
    return http.put<{ success: boolean; error?: string }>(
      `/connections/${connectionId}/keys/${encodeURIComponent(key)}/ttl`,
      { db, ttl }
    )
  },

  // 移除 TTL（持久化）
  removeTTL(connectionId: string, db: number, key: string) {
    return http.delete<{ success: boolean; error?: string }>(
      `/connections/${connectionId}/keys/${encodeURIComponent(key)}/ttl`,
      { data: { db } }
    )
  },

  // 重命名 Key
  rename(connectionId: string, db: number, oldKey: string, newKey: string) {
    return http.post<{ success: boolean; error?: string }>(
      `/connections/${connectionId}/keys/${encodeURIComponent(oldKey)}/rename`,
      { db, newKey }
    )
  },

  // 获取 Key TTL
  getTTL(connectionId: string, db: number, key: string) {
    return http.get<{ success: boolean; ttl: number; error?: string }>(
      `/connections/${connectionId}/keys/${encodeURIComponent(key)}/ttl`,
      { params: { db } }
    )
  },

  // List 操作
  listPush(connectionId: string, db: number, key: string, values: string[], direction: 'left' | 'right') {
    return http.post<{ success: boolean; error?: string }>(
      `/connections/${connectionId}/keys/${encodeURIComponent(key)}/list/push`,
      { db, values, direction }
    )
  },

  listRemove(connectionId: string, db: number, key: string, index: number) {
    return http.delete<{ success: boolean; error?: string }>(
      `/connections/${connectionId}/keys/${encodeURIComponent(key)}/list/${index}`,
      { data: { db } }
    )
  },

  listSet(connectionId: string, db: number, key: string, index: number, value: string) {
    return http.put<{ success: boolean; error?: string }>(
      `/connections/${connectionId}/keys/${encodeURIComponent(key)}/list/${index}`,
      { db, value }
    )
  },

  // Hash 操作
  hashSet(connectionId: string, db: number, key: string, field: string, value: string) {
    return http.put<{ success: boolean; error?: string }>(
      `/connections/${connectionId}/keys/${encodeURIComponent(key)}/hash/${encodeURIComponent(field)}`,
      { db, value }
    )
  },

  hashDelete(connectionId: string, db: number, key: string, fields: string[]) {
    return http.delete<{ success: boolean; error?: string }>(
      `/connections/${connectionId}/keys/${encodeURIComponent(key)}/hash`,
      { data: { db, fields } }
    )
  },

  // Set 操作
  setAdd(connectionId: string, db: number, key: string, members: string[]) {
    return http.post<{ success: boolean; error?: string }>(
      `/connections/${connectionId}/keys/${encodeURIComponent(key)}/set`,
      { db, members }
    )
  },

  setRemove(connectionId: string, db: number, key: string, members: string[]) {
    return http.delete<{ success: boolean; error?: string }>(
      `/connections/${connectionId}/keys/${encodeURIComponent(key)}/set`,
      { data: { db, members } }
    )
  },

  // ZSet 操作
  zsetAdd(connectionId: string, db: number, key: string, members: { member: string; score: number }[]) {
    return http.post<{ success: boolean; error?: string }>(
      `/connections/${connectionId}/keys/${encodeURIComponent(key)}/zset`,
      { db, members }
    )
  },

  zsetRemove(connectionId: string, db: number, key: string, members: string[]) {
    return http.delete<{ success: boolean; error?: string }>(
      `/connections/${connectionId}/keys/${encodeURIComponent(key)}/zset`,
      { data: { db, members } }
    )
  }
}

export default keysApi
