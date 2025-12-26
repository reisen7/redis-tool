<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import { ElMessage } from 'element-plus'
import type { ConnectionConfig, ConnectionMode, SSHAuthType, ConnectionGroup } from '@/types/connection'
import { createDefaultConnection, createDefaultGroup } from '@/types/connection'
import { useConnectionStore } from '@/stores/connection'

const props = defineProps<{
  visible: boolean
  editConfig?: ConnectionConfig | null
  defaultGroupId?: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'saved', config: ConnectionConfig): void
}>()

const connectionStore = useConnectionStore()

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})

const form = reactive<ConnectionConfig>(createDefaultConnection())
const testing = ref(false)
const saving = ref(false)
const activeTab = ref('basic')

// 分组输入（支持选择或输入新名称）
const groupInput = ref('')

// 监听编辑配置变化
watch(() => props.editConfig, (config) => {
  if (config) {
    Object.assign(form, JSON.parse(JSON.stringify(config)))
    // 设置分组输入框的值
    const group = connectionStore.groups.find(g => g.id === config.groupId)
    groupInput.value = group?.name || ''
  } else {
    Object.assign(form, createDefaultConnection())
    groupInput.value = ''
    // 设置默认分组
    if (props.defaultGroupId) {
      form.groupId = props.defaultGroupId
      const group = connectionStore.groups.find(g => g.id === props.defaultGroupId)
      groupInput.value = group?.name || ''
    }
  }
}, { immediate: true })

// 监听默认分组变化
watch(() => props.defaultGroupId, (groupId) => {
  if (!props.editConfig && groupId) {
    form.groupId = groupId
    const group = connectionStore.groups.find(g => g.id === groupId)
    groupInput.value = group?.name || ''
  }
})

// 选择分组
function handleSelectGroup(groupId: string) {
  form.groupId = groupId
  const group = connectionStore.groups.find(g => g.id === groupId)
  groupInput.value = group?.name || ''
}

// 过滤分组选项
const filteredGroups = computed(() => {
  if (!groupInput.value) return connectionStore.groups
  const search = groupInput.value.toLowerCase()
  return connectionStore.groups.filter(g => g.name.toLowerCase().includes(search))
})

// 连接模式选项
const modeOptions: { label: string; value: ConnectionMode }[] = [
  { label: '单机模式', value: 'standalone' },
  { label: '集群模式', value: 'cluster' },
  { label: '哨兵模式', value: 'sentinel' }
]

// SSH 认证类型选项
const sshAuthOptions: { label: string; value: SSHAuthType }[] = [
  { label: '密码', value: 'password' },
  { label: '私钥', value: 'privateKey' }
]

// 添加集群节点
const newClusterNode = ref('')
function addClusterNode() {
  if (newClusterNode.value.trim()) {
    form.clusterNodes.push(newClusterNode.value.trim())
    newClusterNode.value = ''
  }
}
function removeClusterNode(index: number) {
  form.clusterNodes.splice(index, 1)
}

// 添加哨兵节点
const newSentinelNode = ref('')
function addSentinelNode() {
  if (newSentinelNode.value.trim()) {
    form.sentinelNodes.push(newSentinelNode.value.trim())
    newSentinelNode.value = ''
  }
}
function removeSentinelNode(index: number) {
  form.sentinelNodes.splice(index, 1)
}

// 测试连接
async function handleTest() {
  testing.value = true
  try {
    const result = await connectionStore.testConnection(form)
    if (result.success) {
      ElMessage.success('连接成功')
    } else {
      ElMessage.error(result.error || '连接失败')
    }
  } finally {
    testing.value = false
  }
}

// 保存连接
async function handleSave() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入连接名称')
    return
  }
  
  saving.value = true
  try {
    // 处理分组：如果输入了分组名称但没有匹配的分组，则自动创建
    if (groupInput.value.trim()) {
      const existingGroup = connectionStore.groups.find(
        g => g.name.toLowerCase() === groupInput.value.trim().toLowerCase()
      )
      
      if (existingGroup) {
        form.groupId = existingGroup.id
      } else {
        // 创建新分组
        const newGroup = createDefaultGroup()
        newGroup.name = groupInput.value.trim()
        const groupResult = await connectionStore.saveGroup(newGroup)
        if (groupResult.success && groupResult.data) {
          form.groupId = groupResult.data.id
        }
      }
    } else {
      form.groupId = ''
    }
    
    const result = await connectionStore.saveConnection(form)
    if (result.success) {
      ElMessage.success('保存成功')
      dialogVisible.value = false
      emit('saved', result.data || form)
    } else {
      ElMessage.error(result.error || '保存失败')
    }
  } finally {
    saving.value = false
  }
}

function handleClose() {
  dialogVisible.value = false
}
</script>

<template>
  <el-dialog
    v-model="dialogVisible"
    :title="editConfig?.id ? '编辑连接' : '新建连接'"
    width="700px"
    destroy-on-close
  >
    <el-tabs v-model="activeTab">
      <!-- 基本配置 -->
      <el-tab-pane label="基本配置" name="basic">
        <el-form :model="form" label-width="100px">
          <el-form-item label="连接名称" required>
            <el-input v-model="form.name" placeholder="请输入连接名称" />
          </el-form-item>
          
          <el-form-item label="所属分组">
            <el-autocomplete
              v-model="groupInput"
              :fetch-suggestions="(query: string, cb: (suggestions: { value: string; id: string }[]) => void) => {
                const results = filteredGroups.map(g => ({ value: g.name, id: g.id }))
                cb(results)
              }"
              placeholder="输入分组名称（不存在则自动创建）"
              clearable
              style="width: 100%"
              @select="(item: { value: string; id: string }) => handleSelectGroup(item.id)"
              @clear="form.groupId = ''"
            >
              <template #default="{ item }">
                <span class="group-option">
                  <span 
                    class="group-color" 
                    :style="{ backgroundColor: connectionStore.groups.find(g => g.id === item.id)?.color || '#409eff' }"
                  ></span>
                  {{ item.value }}
                </span>
              </template>
            </el-autocomplete>
            <div class="form-tip" style="margin-left: 0; margin-top: 4px;">
              输入新名称将自动创建分组
            </div>
          </el-form-item>
          
          <el-form-item label="连接模式">
            <el-radio-group v-model="form.mode">
              <el-radio-button
                v-for="opt in modeOptions"
                :key="opt.value"
                :value="opt.value"
              >
                {{ opt.label }}
              </el-radio-button>
            </el-radio-group>
          </el-form-item>

          <!-- 单机模式 -->
          <template v-if="form.mode === 'standalone'">
            <el-form-item label="主机地址">
              <el-input v-model="form.host" placeholder="127.0.0.1" />
            </el-form-item>
            <el-form-item label="端口">
              <el-input-number v-model="form.port" :min="1" :max="65535" />
            </el-form-item>
          </template>

          <!-- 集群模式 -->
          <template v-if="form.mode === 'cluster'">
            <el-form-item label="集群节点">
              <div class="node-list">
                <el-tag
                  v-for="(node, index) in form.clusterNodes"
                  :key="index"
                  closable
                  @close="removeClusterNode(index)"
                  class="node-tag"
                >
                  {{ node }}
                </el-tag>
              </div>
              <div class="node-input">
                <el-input
                  v-model="newClusterNode"
                  placeholder="host:port"
                  @keyup.enter="addClusterNode"
                />
                <el-button @click="addClusterNode">添加</el-button>
              </div>
            </el-form-item>
          </template>

          <!-- 哨兵模式 -->
          <template v-if="form.mode === 'sentinel'">
            <el-form-item label="Master 名称">
              <el-input v-model="form.sentinelMasterName" placeholder="mymaster" />
            </el-form-item>
            <el-form-item label="哨兵节点">
              <div class="node-list">
                <el-tag
                  v-for="(node, index) in form.sentinelNodes"
                  :key="index"
                  closable
                  @close="removeSentinelNode(index)"
                  class="node-tag"
                >
                  {{ node }}
                </el-tag>
              </div>
              <div class="node-input">
                <el-input
                  v-model="newSentinelNode"
                  placeholder="host:port"
                  @keyup.enter="addSentinelNode"
                />
                <el-button @click="addSentinelNode">添加</el-button>
              </div>
            </el-form-item>
          </template>

          <el-form-item label="用户名">
            <el-input v-model="form.username" placeholder="可选，Redis 6.0+ ACL" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input
              v-model="form.password"
              type="password"
              show-password
              placeholder="可选"
            />
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- SSH 隧道 -->
      <el-tab-pane label="SSH 隧道" name="ssh">
        <el-form :model="form.ssh" label-width="100px">
          <el-form-item label="启用 SSH">
            <el-switch v-model="form.ssh.enabled" />
          </el-form-item>
          
          <template v-if="form.ssh.enabled">
            <el-form-item label="SSH 主机">
              <el-input v-model="form.ssh.host" placeholder="SSH 服务器地址" />
            </el-form-item>
            <el-form-item label="SSH 端口">
              <el-input-number v-model="form.ssh.port" :min="1" :max="65535" />
            </el-form-item>
            <el-form-item label="用户名">
              <el-input v-model="form.ssh.username" />
            </el-form-item>
            <el-form-item label="认证方式">
              <el-radio-group v-model="form.ssh.authType">
                <el-radio
                  v-for="opt in sshAuthOptions"
                  :key="opt.value"
                  :value="opt.value"
                >
                  {{ opt.label }}
                </el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item v-if="form.ssh.authType === 'password'" label="密码">
              <el-input
                v-model="form.ssh.password"
                type="password"
                show-password
              />
            </el-form-item>
            <template v-else>
              <el-form-item label="私钥">
                <el-input
                  v-model="form.ssh.privateKey"
                  type="textarea"
                  :rows="4"
                  placeholder="粘贴私钥内容"
                />
              </el-form-item>
              <el-form-item label="私钥密码">
                <el-input
                  v-model="form.ssh.privateKeyPassphrase"
                  type="password"
                  show-password
                  placeholder="可选"
                />
              </el-form-item>
            </template>
          </template>
        </el-form>
      </el-tab-pane>

      <!-- TLS 配置 -->
      <el-tab-pane label="TLS/SSL" name="tls">
        <el-form :model="form.tls" label-width="100px">
          <el-form-item label="启用 TLS">
            <el-switch v-model="form.tls.enabled" />
          </el-form-item>
          
          <template v-if="form.tls.enabled">
            <el-form-item label="CA 证书">
              <el-input
                v-model="form.tls.caCert"
                type="textarea"
                :rows="3"
                placeholder="粘贴 CA 证书内容"
              />
            </el-form-item>
            <el-form-item label="客户端证书">
              <el-input
                v-model="form.tls.clientCert"
                type="textarea"
                :rows="3"
                placeholder="粘贴客户端证书内容"
              />
            </el-form-item>
            <el-form-item label="客户端私钥">
              <el-input
                v-model="form.tls.clientKey"
                type="textarea"
                :rows="3"
                placeholder="粘贴客户端私钥内容"
              />
            </el-form-item>
            <el-form-item label="跳过验证">
              <el-switch v-model="form.tls.skipVerify" />
              <span class="form-tip">仅用于测试环境</span>
            </el-form-item>
          </template>
        </el-form>
      </el-tab-pane>
    </el-tabs>

    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button :loading="testing" @click="handleTest">测试连接</el-button>
      <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.node-list {
  margin-bottom: 8px;
}
.node-tag {
  margin-right: 8px;
  margin-bottom: 4px;
}
.node-input {
  display: flex;
  gap: 8px;
}
.node-input .el-input {
  flex: 1;
}
.form-tip {
  margin-left: 12px;
  color: #909399;
  font-size: 12px;
}
.group-option {
  display: flex;
  align-items: center;
  gap: 8px;
}
.group-color {
  width: 12px;
  height: 12px;
  border-radius: 2px;
}
</style>
