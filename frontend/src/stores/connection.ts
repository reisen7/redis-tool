import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ConnectionConfig, ConnectionInstance, ConnectionGroup, DatabaseInfo } from '@/types/connection'
import { connectionApi, groupApi } from '@/api/connection'

export const useConnectionStore = defineStore('connection', () => {
  // 所有分组
  const groups = ref<ConnectionGroup[]>([])
  
  // 所有保存的连接配置
  const connections = ref<ConnectionConfig[]>([])
  
  // 已打开的连接实例
  const openConnections = ref<Map<string, ConnectionInstance>>(new Map())
  
  // 当前活动的连接 ID
  const activeConnectionId = ref<string | null>(null)
  
  // 当前活动的连接实例
  const activeConnection = computed(() => {
    if (!activeConnectionId.value) return null
    return openConnections.value.get(activeConnectionId.value) || null
  })

  // 按分组组织的连接
  const connectionsByGroup = computed(() => {
    const result: Map<string, ConnectionConfig[]> = new Map()
    
    // 初始化分组
    result.set('', []) // 未分组
    for (const group of groups.value) {
      result.set(group.id, [])
    }
    
    // 分配连接到分组
    for (const conn of connections.value) {
      const groupId = conn.groupId || ''
      if (!result.has(groupId)) {
        result.set(groupId, [])
      }
      result.get(groupId)!.push(conn)
    }
    
    return result
  })

  // --- 分组操作 ---

  // 加载分组列表
  async function loadGroups() {
    try {
      const res = await groupApi.list()
      if (res.data.success) {
        groups.value = res.data.data || []
      }
    } catch (error) {
      console.error('Failed to load groups:', error)
    }
  }

  // 保存分组
  async function saveGroup(group: ConnectionGroup) {
    try {
      if (group.id) {
        const res = await groupApi.update(group.id, group)
        if (res.data.success) {
          const index = groups.value.findIndex((g: ConnectionGroup) => g.id === group.id)
          if (index >= 0) {
            groups.value[index] = res.data.data
          }
          return { success: true }
        }
      } else {
        const res = await groupApi.create(group)
        if (res.data.success) {
          groups.value.push(res.data.data)
          return { success: true, data: res.data.data }
        }
      }
      return { success: false, error: 'Failed to save group' }
    } catch (error: any) {
      return { success: false, error: error.message }
    }
  }

  // 删除分组
  async function deleteGroup(id: string) {
    try {
      await groupApi.delete(id)
      groups.value = groups.value.filter((g: ConnectionGroup) => g.id !== id)
      // 将该分组下的连接移到未分组
      for (const conn of connections.value) {
        if (conn.groupId === id) {
          conn.groupId = ''
        }
      }
      return { success: true }
    } catch (error: any) {
      return { success: false, error: error.message }
    }
  }

  // --- 连接操作 ---

  // 加载连接列表
  async function loadConnections() {
    try {
      const res = await connectionApi.list()
      if (res.data.success) {
        connections.value = res.data.data || []
      }
    } catch (error) {
      console.error('Failed to load connections:', error)
    }
  }

  // 保存连接
  async function saveConnection(config: ConnectionConfig) {
    try {
      if (config.id) {
        const res = await connectionApi.update(config.id, config)
        if (res.data.success) {
          const index = connections.value.findIndex((c: ConnectionConfig) => c.id === config.id)
          if (index >= 0) {
            connections.value[index] = res.data.data
          }
          return { success: true }
        }
      } else {
        const res = await connectionApi.create(config)
        if (res.data.success) {
          connections.value.push(res.data.data)
          return { success: true, data: res.data.data }
        }
      }
      return { success: false, error: 'Failed to save connection' }
    } catch (error: any) {
      return { success: false, error: error.message }
    }
  }

  // 删除连接
  async function deleteConnection(id: string) {
    try {
      await connectionApi.delete(id)
      connections.value = connections.value.filter((c: ConnectionConfig) => c.id !== id)
      
      // 如果已打开，关闭它
      if (openConnections.value.has(id)) {
        openConnections.value.delete(id)
        if (activeConnectionId.value === id) {
          const firstOpen = openConnections.value.keys().next().value
          activeConnectionId.value = firstOpen || null
        }
      }
      return { success: true }
    } catch (error: any) {
      return { success: false, error: error.message }
    }
  }

  // 测试连接
  async function testConnection(config: ConnectionConfig) {
    try {
      const res = await connectionApi.test(config)
      return res.data
    } catch (error: any) {
      return { success: false, error: error.message }
    }
  }

  // 打开连接
  async function openConnection(config: ConnectionConfig) {
    const id = config.id
    
    // 如果已经打开，直接激活
    if (openConnections.value.has(id)) {
      activeConnectionId.value = id
      return { success: true, alreadyOpen: true }
    }

    // 创建连接实例
    const instance: ConnectionInstance = {
      config,
      status: 'connecting',
      databases: [],
      currentDb: 0
    }
    
    openConnections.value.set(id, instance)
    activeConnectionId.value = id

    try {
      // 连接到 Redis
      const connectRes = await connectionApi.connect(id)
      if (!connectRes.data.success) {
        instance.status = 'error'
        instance.error = connectRes.data.error
        return { success: false, error: connectRes.data.error }
      }

      // 获取数据库列表
      const dbRes = await connectionApi.getDatabases(id)
      if (dbRes.data.success) {
        instance.databases = dbRes.data.data || []
      }

      instance.status = 'connected'
      return { success: true, alreadyOpen: false }
    } catch (error: any) {
      instance.status = 'error'
      instance.error = error.message
      return { success: false, error: error.message }
    }
  }

  // 关闭连接
  async function closeConnection(id: string) {
    try {
      await connectionApi.disconnect(id)
    } catch {}
    
    openConnections.value.delete(id)
    
    if (activeConnectionId.value === id) {
      const firstOpen = openConnections.value.keys().next().value
      activeConnectionId.value = firstOpen || null
    }
  }

  // 切换活动连接
  function setActiveConnection(id: string) {
    if (openConnections.value.has(id)) {
      activeConnectionId.value = id
    }
  }

  // 切换数据库
  async function selectDatabase(connectionId: string, dbIndex: number) {
    const instance = openConnections.value.get(connectionId)
    if (!instance) return { success: false, error: 'Connection not found' }

    try {
      const res = await connectionApi.selectDatabase(connectionId, dbIndex)
      if (res.data.success) {
        instance.currentDb = dbIndex
        return { success: true }
      }
      return { success: false, error: 'Failed to select database' }
    } catch (error: any) {
      return { success: false, error: error.message }
    }
  }

  // 更新数据库别名
  async function updateDatabaseAlias(connectionId: string, dbIndex: number, alias: string) {
    const instance = openConnections.value.get(connectionId)
    if (!instance) return { success: false, error: 'Connection not found' }

    try {
      await connectionApi.updateDatabaseAlias(connectionId, dbIndex, alias)
      const db = instance.databases.find((d: DatabaseInfo) => d.index === dbIndex)
      if (db) {
        db.alias = alias
      }
      return { success: true }
    } catch (error: any) {
      return { success: false, error: error.message }
    }
  }

  // 刷新数据库信息
  async function refreshDatabases(connectionId: string) {
    const instance = openConnections.value.get(connectionId)
    if (!instance) return

    try {
      const res = await connectionApi.getDatabases(connectionId)
      if (res.data.success) {
        instance.databases = res.data.data || []
      }
    } catch (error) {
      console.error('Failed to refresh databases:', error)
    }
  }

  // Reset store state (for logout)
  function reset() {
    // Close all open connections
    openConnections.value.clear()
    activeConnectionId.value = null
    // Clear connections and groups (will be reloaded on next login)
    connections.value = []
    groups.value = []
  }

  return {
    groups,
    connections,
    openConnections,
    activeConnectionId,
    activeConnection,
    connectionsByGroup,
    loadGroups,
    saveGroup,
    deleteGroup,
    loadConnections,
    saveConnection,
    deleteConnection,
    testConnection,
    openConnection,
    closeConnection,
    setActiveConnection,
    selectDatabase,
    updateDatabaseAlias,
    refreshDatabases,
    reset
  }
})
