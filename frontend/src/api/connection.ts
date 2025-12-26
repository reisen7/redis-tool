import http from './http'
import type { ConnectionConfig, ConnectionGroup, DatabaseInfo } from '@/types/connection'

// 分组管理 API
export const groupApi = {
  // 获取所有分组
  list() {
    return http.get<{ success: boolean; data: ConnectionGroup[] }>('/groups')
  },

  // 创建分组
  create(group: ConnectionGroup) {
    return http.post<{ success: boolean; data: ConnectionGroup }>('/groups', group)
  },

  // 更新分组
  update(id: string, group: ConnectionGroup) {
    return http.put<{ success: boolean; data: ConnectionGroup }>(`/groups/${id}`, group)
  },

  // 删除分组
  delete(id: string) {
    return http.delete<{ success: boolean }>(`/groups/${id}`)
  }
}

// 连接管理 API
export const connectionApi = {
  // 获取所有连接
  list() {
    return http.get<{ success: boolean; data: ConnectionConfig[] }>('/connections')
  },

  // 创建连接
  create(config: ConnectionConfig) {
    return http.post<{ success: boolean; data: ConnectionConfig }>('/connections', config)
  },

  // 更新连接
  update(id: string, config: ConnectionConfig) {
    return http.put<{ success: boolean; data: ConnectionConfig }>(`/connections/${id}`, config)
  },

  // 删除连接
  delete(id: string) {
    return http.delete<{ success: boolean }>(`/connections/${id}`)
  },

  // 测试连接
  test(config: ConnectionConfig) {
    return http.post<{ success: boolean; error?: string }>('/connections/test', config)
  },

  // 连接到 Redis
  connect(id: string) {
    return http.post<{ success: boolean; error?: string }>(`/connections/${id}/connect`)
  },

  // 断开连接
  disconnect(id: string) {
    return http.post<{ success: boolean }>(`/connections/${id}/disconnect`)
  },

  // 获取数据库列表
  getDatabases(id: string) {
    return http.get<{ success: boolean; data: DatabaseInfo[] }>(`/connections/${id}/databases`)
  },

  // 更新数据库别名
  updateDatabaseAlias(id: string, dbIndex: number, alias: string) {
    return http.put<{ success: boolean }>(`/connections/${id}/databases/${dbIndex}`, { alias })
  },

  // 切换数据库
  selectDatabase(id: string, dbIndex: number) {
    return http.post<{ success: boolean }>(`/connections/${id}/databases/${dbIndex}/select`)
  }
}

export default connectionApi
