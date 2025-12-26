<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { Tab } from '@/types/tab'
import { useTabsStore } from '@/stores/tabs'
import { useConnectionStore } from '@/stores/connection'
import ResizableDivider from './ResizableDivider.vue'
import KeyList from '@/components/KeyList.vue'
import KeyDetail from '@/components/KeyDetail.vue'
import RedisTerminal from '@/components/RedisTerminal.vue'

const props = defineProps<{
  tab: Tab
}>()

const tabsStore = useTabsStore()
const connectionStore = useConnectionStore()

// Local divider position that syncs with tab state
const dividerPosition = ref(props.tab.dividerPosition)

// Watch for tab changes and update local position
watch(() => props.tab.id, () => {
  dividerPosition.value = props.tab.dividerPosition
}, { immediate: true })

// Watch for external changes to the tab's divider position
watch(() => props.tab.dividerPosition, (newPos) => {
  if (newPos !== dividerPosition.value) {
    dividerPosition.value = newPos
  }
})

// Computed styles for the panels
const topPanelStyle = computed(() => ({
  height: `${dividerPosition.value}%`
}))

const bottomPanelStyle = computed(() => ({
  height: `${100 - dividerPosition.value}%`
}))

// Handle divider position change
function onDividerPositionChange(newPosition: number) {
  dividerPosition.value = newPosition
}

// Persist position when drag ends
function onDragEnd() {
  tabsStore.updateDividerPosition(props.tab.id, dividerPosition.value)
}

// Right panel tab for key detail vs terminal
const activeRightTab = ref('detail')
</script>

<template>
  <div class="workspace-panel">
    <!-- Top section: Keys list -->
    <div class="panel-section top-section" :style="topPanelStyle">
      <div class="section-content">
        <KeyList :tab-id="tab.id" @create-key="() => {}" />
      </div>
    </div>

    <!-- Resizable divider -->
    <ResizableDivider
      :position="dividerPosition"
      :min-position="20"
      :max-position="80"
      direction="horizontal"
      @update:position="onDividerPositionChange"
      @drag-end="onDragEnd"
    />

    <!-- Bottom section: Terminal and Key Detail -->
    <div class="panel-section bottom-section" :style="bottomPanelStyle">
      <el-tabs v-model="activeRightTab" class="bottom-tabs">
        <el-tab-pane label="Key 详情" name="detail">
          <div class="tab-pane-content">
            <KeyDetail />
          </div>
        </el-tab-pane>
        <el-tab-pane label="终端" name="terminal">
          <div class="tab-pane-content">
            <RedisTerminal :tab-id="tab.id" />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<style scoped>
.workspace-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.panel-section {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 100px;
}

.section-header {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: #fafafa;
  border-bottom: 1px solid #e4e7ed;
  flex-shrink: 0;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #303133;
}

.section-content {
  flex: 1;
  overflow: hidden;
}

.top-section {
  border-bottom: none;
}

.bottom-section {
  display: flex;
  flex-direction: column;
}

.bottom-tabs {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.bottom-tabs :deep(.el-tabs__header) {
  margin: 0;
  padding: 0 12px;
  background: #fafafa;
  border-bottom: 1px solid #e4e7ed;
  flex-shrink: 0;
}

.bottom-tabs :deep(.el-tabs__content) {
  flex: 1;
  overflow: hidden;
  padding: 0;
}

.bottom-tabs :deep(.el-tab-pane) {
  height: 100%;
}

.tab-pane-content {
  height: 100%;
  overflow: hidden;
}
</style>
