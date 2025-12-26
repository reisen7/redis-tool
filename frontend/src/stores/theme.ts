import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

// 主题模式
export type ThemeMode = 'light' | 'dark' | 'system'

// 主题配置
export interface ThemeConfig {
  mode: ThemeMode
  primaryColor: string
  fontSize: number
  fontFamily: string
  compactMode: boolean
  borderRadius: number
}

// 预设主题色
export const THEME_COLORS = [
  { name: '蓝色', value: '#409eff' },
  { name: '绿色', value: '#67c23a' },
  { name: '橙色', value: '#e6a23c' },
  { name: '红色', value: '#f56c6c' },
  { name: '紫色', value: '#9c27b0' }
]

// 字体选项
export const FONT_OPTIONS = [
  { label: '系统默认', value: 'system-ui, -apple-system, sans-serif' },
  { label: 'Monaco', value: 'Monaco, monospace' },
  { label: 'Consolas', value: 'Consolas, monospace' },
  { label: 'Source Code Pro', value: '"Source Code Pro", monospace' }
]

// 默认配置
const defaultTheme: ThemeConfig = {
  mode: 'light',
  primaryColor: '#409eff',
  fontSize: 14,
  fontFamily: 'system-ui, -apple-system, sans-serif',
  compactMode: false,
  borderRadius: 8
}

const THEME_STORAGE_KEY = 'redis-manager-theme'

export const useThemeStore = defineStore('theme', () => {
  const config = ref<ThemeConfig>({ ...defaultTheme })
  const panelVisible = ref(false)

  // 从 localStorage 加载配置
  function loadTheme() {
    try {
      const saved = localStorage.getItem(THEME_STORAGE_KEY)
      if (saved) {
        const parsed = JSON.parse(saved)
        config.value = { ...defaultTheme, ...parsed }
      }
    } catch (e) {
      console.error('Failed to load theme:', e)
    }
    applyTheme()
  }

  // 保存配置到 localStorage
  function saveTheme() {
    try {
      localStorage.setItem(THEME_STORAGE_KEY, JSON.stringify(config.value))
    } catch (e) {
      console.error('Failed to save theme:', e)
    }
  }

  // 应用主题到 DOM
  function applyTheme() {
    const root = document.documentElement
    const { mode, primaryColor, fontSize, fontFamily, compactMode, borderRadius } = config.value

    // 确定实际主题模式
    let actualMode = mode
    if (mode === 'system') {
      actualMode = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
    }

    // 设置主题属性
    root.setAttribute('data-theme', actualMode)

    // 设置 CSS 变量
    root.style.setProperty('--el-color-primary', primaryColor)
    root.style.setProperty('--theme-primary-color', primaryColor)
    root.style.setProperty('--theme-font-size', `${fontSize}px`)
    root.style.setProperty('--theme-font-family', fontFamily)
    root.style.setProperty('--theme-border-radius', `${borderRadius}px`)

    // 紧凑模式
    if (compactMode) {
      root.classList.add('compact-mode')
    } else {
      root.classList.remove('compact-mode')
    }
  }

  // 更新配置
  function updateConfig(partial: Partial<ThemeConfig>) {
    config.value = { ...config.value, ...partial }
    applyTheme()
    saveTheme()
  }

  // 重置为默认
  function resetTheme() {
    config.value = { ...defaultTheme }
    applyTheme()
    saveTheme()
  }

  // 切换面板显示
  function togglePanel() {
    panelVisible.value = !panelVisible.value
  }

  // 监听系统主题变化
  function setupSystemThemeListener() {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    mediaQuery.addEventListener('change', () => {
      if (config.value.mode === 'system') {
        applyTheme()
      }
    })
  }

  return {
    config,
    panelVisible,
    loadTheme,
    saveTheme,
    applyTheme,
    updateConfig,
    resetTheme,
    togglePanel,
    setupSystemThemeListener
  }
})
