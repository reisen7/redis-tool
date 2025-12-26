<script setup lang="ts">
import { computed } from 'vue'
import { Setting, Close, Refresh } from '@element-plus/icons-vue'
import { useThemeStore, THEME_COLORS, FONT_OPTIONS, type ThemeMode } from '@/stores/theme'

const themeStore = useThemeStore()

const config = computed(() => themeStore.config)

function handleModeChange(mode: ThemeMode) {
  themeStore.updateConfig({ mode })
}

function handleColorChange(color: string) {
  themeStore.updateConfig({ primaryColor: color })
}

function handleFontSizeChange(size: number) {
  themeStore.updateConfig({ fontSize: size })
}

function handleFontFamilyChange(font: string) {
  themeStore.updateConfig({ fontFamily: font })
}

function handleCompactChange(compact: boolean) {
  themeStore.updateConfig({ compactMode: compact })
}

function handleBorderRadiusChange(radius: number) {
  themeStore.updateConfig({ borderRadius: radius })
}

function handleReset() {
  themeStore.resetTheme()
}

function handleClose() {
  themeStore.togglePanel()
}
</script>

<template>
  <!-- 浮动按钮 -->
  <div
    v-if="!themeStore.panelVisible"
    class="theme-fab"
    @click="themeStore.togglePanel"
  >
    <el-icon :size="20"><Setting /></el-icon>
  </div>

  <!-- 主题配置面板 -->
  <transition name="slide">
    <div v-if="themeStore.panelVisible" class="theme-panel">
      <div class="panel-header">
        <span class="panel-title">
          <el-icon><Setting /></el-icon>
          主题设置
        </span>
        <el-button type="primary" link @click="handleClose">
          <el-icon><Close /></el-icon>
        </el-button>
      </div>

      <div class="panel-content">
        <!-- 主题模式 -->
        <div class="config-item">
          <label class="config-label">主题模式</label>
          <el-radio-group
            :model-value="config.mode"
            size="small"
            @change="handleModeChange"
          >
            <el-radio-button value="light">浅色</el-radio-button>
            <el-radio-button value="dark">深色</el-radio-button>
            <el-radio-button value="system">跟随系统</el-radio-button>
          </el-radio-group>
        </div>

        <!-- 主题色 -->
        <div class="config-item">
          <label class="config-label">主题色</label>
          <div class="color-options">
            <div
              v-for="color in THEME_COLORS"
              :key="color.value"
              class="color-option"
              :class="{ active: config.primaryColor === color.value }"
              :style="{ backgroundColor: color.value }"
              :title="color.name"
              @click="handleColorChange(color.value)"
            />
            <el-color-picker
              :model-value="config.primaryColor"
              size="small"
              @change="handleColorChange"
            />
          </div>
        </div>

        <!-- 字体大小 -->
        <div class="config-item">
          <label class="config-label">字体大小</label>
          <div class="slider-row">
            <el-slider
              :model-value="config.fontSize"
              :min="12"
              :max="18"
              :step="1"
              :show-tooltip="false"
              @change="handleFontSizeChange"
            />
            <span class="slider-value">{{ config.fontSize }}px</span>
          </div>
        </div>

        <!-- 字体 -->
        <div class="config-item">
          <label class="config-label">字体</label>
          <el-select
            :model-value="config.fontFamily"
            size="small"
            style="width: 100%"
            @change="handleFontFamilyChange"
          >
            <el-option
              v-for="font in FONT_OPTIONS"
              :key="font.value"
              :label="font.label"
              :value="font.value"
            />
          </el-select>
        </div>

        <!-- 圆角大小 -->
        <div class="config-item">
          <label class="config-label">圆角大小</label>
          <div class="slider-row">
            <el-slider
              :model-value="config.borderRadius"
              :min="0"
              :max="16"
              :step="2"
              :show-tooltip="false"
              @change="handleBorderRadiusChange"
            />
            <span class="slider-value">{{ config.borderRadius }}px</span>
          </div>
        </div>

        <!-- 紧凑模式 -->
        <div class="config-item">
          <label class="config-label">紧凑模式</label>
          <el-switch
            :model-value="config.compactMode"
            @change="handleCompactChange"
          />
        </div>
      </div>

      <div class="panel-footer">
        <el-button size="small" :icon="Refresh" @click="handleReset">
          重置
        </el-button>
      </div>
    </div>
  </transition>
</template>

<style scoped>
.theme-fab {
  position: fixed;
  right: 20px;
  bottom: 20px;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: var(--el-color-primary, #409eff);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transition: all 0.3s;
  z-index: 1000;
}

.theme-fab:hover {
  transform: scale(1.1);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2);
}

.theme-panel {
  position: fixed;
  right: 20px;
  bottom: 20px;
  width: 280px;
  background: var(--el-bg-color, #fff);
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  z-index: 1001;
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-light, #e4e7ed);
  background: var(--el-fill-color-light, #f5f7fa);
}

.panel-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 14px;
}

.panel-content {
  padding: 16px;
  max-height: 400px;
  overflow-y: auto;
}

.config-item {
  margin-bottom: 16px;
}

.config-item:last-child {
  margin-bottom: 0;
}

.config-label {
  display: block;
  font-size: 12px;
  color: var(--el-text-color-secondary, #909399);
  margin-bottom: 8px;
}

.color-options {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.color-option {
  width: 24px;
  height: 24px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  border: 2px solid transparent;
}

.color-option:hover {
  transform: scale(1.1);
}

.color-option.active {
  border-color: var(--el-text-color-primary, #303133);
  box-shadow: 0 0 0 2px rgba(0, 0, 0, 0.1);
}

.slider-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.slider-row .el-slider {
  flex: 1;
}

.slider-value {
  font-size: 12px;
  color: var(--el-text-color-secondary, #909399);
  min-width: 40px;
  text-align: right;
}

.panel-footer {
  padding: 12px 16px;
  border-top: 1px solid var(--el-border-color-light, #e4e7ed);
  display: flex;
  justify-content: flex-end;
}

/* 动画 */
.slide-enter-active,
.slide-leave-active {
  transition: all 0.3s ease;
}

.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  transform: translateY(20px);
}
</style>
