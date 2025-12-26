import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { KeyInfo, KeyValue, SearchConfig, CreateKeyRequest } from '@/types/key'
import keysApi from '@/api/keys'
import { useConnectionStore } from '@/stores/connection'

export const useKeysStore = defineStore('keys', () => {
  // 获取 connection store（延迟获取避免循环依赖）
  const getConnectionStore = () => useConnectionStore()
  // Key 列表
  const keys = ref<KeyInfo[]>([])
  
  // 当前选中的 Key
  const selectedKey = ref<string | null>(null)
  
  // 当前 Key 的值
  const currentKeyValue = ref<KeyValue | null>(null)
  
  // 加载状态
  const loading = ref(false)
  const loadingValue = ref(false)
  
  // 搜索配置
  const searchConfig = ref<SearchConfig>({
    searchType: 'scan',
    pattern: '*',
    scanCount: 100,
    regexPattern: ''
  })
  
  // 游标（用于分页）
  const cursor = ref(0)
  const hasMore = ref(true)

  // 加载 Key 列表
  async function loadKeys(connectionId: string, db: number, reset: boolean = true) {
    if (loading.value) return
    
    loading.value = true
    
    if (reset) {
      keys.value = []
      cursor.value = 0
      hasMore.value = true
    }

    try {
      const res = await keysApi.list(
        connectionId,
        db,
        searchConfig.value.pattern,
        cursor.value,
        searchConfig.value.scanCount
      )
      
      if (res.data.success) {
        if (reset) {
          keys.value = res.data.keys || []
        } else {
          keys.value = [...keys.value, ...(res.data.keys || [])]
        }
        cursor.value = res.data.cursor
        hasMore.value = res.data.cursor !== 0
      }
    } catch (error) {
      console.error('Failed to load keys:', error)
    } finally {
      loading.value = false
    }
  }

  // 搜索 Key（支持正则）
  async function searchKeys(connectionId: string, db: number) {
    loading.value = true
    keys.value = []
    cursor.value = 0

    try {
      const res = await keysApi.scan(connectionId, db, searchConfig.value)
      
      if (res.data.success) {
        keys.value = res.data.keys || []
        cursor.value = res.data.cursor
        hasMore.value = res.data.cursor !== 0
      }
    } catch (error) {
      console.error('Failed to search keys:', error)
    } finally {
      loading.value = false
    }
  }

  // 加载更多
  async function loadMore(connectionId: string, db: number) {
    if (!hasMore.value || loading.value) return
    await loadKeys(connectionId, db, false)
  }

  // 获取 Key 值
  async function getKeyValue(connectionId: string, db: number, key: string) {
    loadingValue.value = true
    selectedKey.value = key

    try {
      const res = await keysApi.get(connectionId, db, key)
      
      if (res.data.success) {
        currentKeyValue.value = res.data.data
      } else {
        currentKeyValue.value = null
      }
    } catch (error) {
      console.error('Failed to get key value:', error)
      currentKeyValue.value = null
    } finally {
      loadingValue.value = false
    }
  }

  // 刷新当前 Key 值
  async function refreshCurrentKey(connectionId: string, db: number) {
    if (selectedKey.value) {
      await getKeyValue(connectionId, db, selectedKey.value)
    }
  }

  // 创建 Key
  async function createKey(connectionId: string, db: number, request: CreateKeyRequest) {
    try {
      const res = await keysApi.create(connectionId, db, request)
      if (res.data.success) {
        // 刷新列表
        await loadKeys(connectionId, db, true)
        // 刷新数据库 Key 数量
        await getConnectionStore().refreshDatabases(connectionId)
        return { success: true }
      }
      return { success: false, error: res.data.error }
    } catch (error: any) {
      return { success: false, error: error.message }
    }
  }

  // 更新 Key 值
  async function updateKeyValue(connectionId: string, db: number, key: string, value: any) {
    try {
      const res = await keysApi.update(connectionId, db, key, value)
      if (res.data.success) {
        await refreshCurrentKey(connectionId, db)
        return { success: true }
      }
      return { success: false, error: res.data.error }
    } catch (error: any) {
      return { success: false, error: error.message }
    }
  }

  // 删除 Key
  async function deleteKeys(connectionId: string, db: number, keysToDelete: string[]) {
    try {
      const res = await keysApi.delete(connectionId, db, keysToDelete)
      if (res.data.success) {
        // 从列表中移除
        keys.value = keys.value.filter((k: KeyInfo) => !keysToDelete.includes(k.key))
        
        // 如果删除的是当前选中的 Key
        if (selectedKey.value && keysToDelete.includes(selectedKey.value)) {
          selectedKey.value = null
          currentKeyValue.value = null
        }
        
        // 刷新数据库 Key 数量
        await getConnectionStore().refreshDatabases(connectionId)
        
        return { success: true, deleted: res.data.deleted }
      }
      return { success: false, error: res.data.error }
    } catch (error: any) {
      return { success: false, error: error.message }
    }
  }

  // 设置 TTL
  async function setKeyTTL(connectionId: string, db: number, key: string, ttl: number) {
    try {
      const res = await keysApi.setTTL(connectionId, db, key, ttl)
      if (res.data.success) {
        // 更新本地数据
        const keyInfo = keys.value.find((k: KeyInfo) => k.key === key)
        if (keyInfo) {
          keyInfo.ttl = ttl
        }
        if (currentKeyValue.value && currentKeyValue.value.key === key) {
          currentKeyValue.value.ttl = ttl
        }
        return { success: true }
      }
      return { success: false, error: res.data.error }
    } catch (error: any) {
      return { success: false, error: error.message }
    }
  }

  // 移除 TTL
  async function removeKeyTTL(connectionId: string, db: number, key: string) {
    try {
      const res = await keysApi.removeTTL(connectionId, db, key)
      if (res.data.success) {
        const keyInfo = keys.value.find((k: KeyInfo) => k.key === key)
        if (keyInfo) {
          keyInfo.ttl = -1
        }
        if (currentKeyValue.value && currentKeyValue.value.key === key) {
          currentKeyValue.value.ttl = -1
        }
        return { success: true }
      }
      return { success: false, error: res.data.error }
    } catch (error: any) {
      return { success: false, error: error.message }
    }
  }

  // 重命名 Key
  async function renameKey(connectionId: string, db: number, oldKey: string, newKey: string) {
    try {
      const res = await keysApi.rename(connectionId, db, oldKey, newKey)
      if (res.data.success) {
        // 更新列表
        const keyInfo = keys.value.find((k: KeyInfo) => k.key === oldKey)
        if (keyInfo) {
          keyInfo.key = newKey
        }
        if (selectedKey.value === oldKey) {
          selectedKey.value = newKey
          if (currentKeyValue.value) {
            currentKeyValue.value.key = newKey
          }
        }
        return { success: true }
      }
      return { success: false, error: res.data.error }
    } catch (error: any) {
      return { success: false, error: error.message }
    }
  }

  // 清除选中
  function clearSelection() {
    selectedKey.value = null
    currentKeyValue.value = null
  }

  // 更新搜索配置
  function updateSearchConfig(config: Partial<SearchConfig>) {
    searchConfig.value = { ...searchConfig.value, ...config }
  }

  return {
    keys,
    selectedKey,
    currentKeyValue,
    loading,
    loadingValue,
    searchConfig,
    cursor,
    hasMore,
    loadKeys,
    searchKeys,
    loadMore,
    getKeyValue,
    refreshCurrentKey,
    createKey,
    updateKeyValue,
    deleteKeys,
    setKeyTTL,
    removeKeyTTL,
    renameKey,
    clearSelection,
    updateSearchConfig
  }
})
