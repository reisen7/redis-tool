<script setup lang="ts">
import { onMounted } from 'vue'
import { useThemeStore } from '@/stores/theme'
import { useAuthStore } from '@/stores/auth'

const themeStore = useThemeStore()
const authStore = useAuthStore()

onMounted(() => {
  // Initialize theme
  themeStore.loadTheme()
  themeStore.setupSystemThemeListener()
  
  // Initialize auth state from localStorage
  authStore.initialize()
})
</script>

<template>
  <router-view />
</template>

<style>
/* Global styles can go here if needed */
</style>
