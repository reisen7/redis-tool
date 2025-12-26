import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

// Route definitions
const routes: RouteRecordRaw[] = [
  // Public routes
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { requiresAuth: false, isPublic: true }
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('@/views/RegisterView.vue'),
    meta: { requiresAuth: false, isPublic: true }
  },
  {
    path: '/2fa',
    name: 'two-factor',
    component: () => import('@/views/TwoFactorView.vue'),
    meta: { requiresAuth: false, isPublic: true }
  },
  
  // Protected routes
  {
    path: '/',
    name: 'app',
    component: () => import('@/views/MainView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/app',
    redirect: '/'
  },
  
  // Admin routes
  {
    path: '/admin',
    name: 'admin',
    component: () => import('@/views/AdminView.vue'),
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  
  // Catch-all redirect to app
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]

// Create router instance
const router = createRouter({
  history: createWebHistory(),
  routes
})

// Navigation guards
router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()
  
  const requiresAuth = to.meta.requiresAuth !== false
  const requiresAdmin = to.meta.requiresAdmin === true
  const isPublic = to.meta.isPublic === true
  
  // Check if user is authenticated
  const isAuthenticated = authStore.isAuthenticated
  const isAdmin = authStore.isAdmin
  
  // If route requires authentication and user is not authenticated
  if (requiresAuth && !isAuthenticated) {
    // Redirect to login page
    next({ name: 'login', query: { redirect: to.fullPath } })
    return
  }
  
  // If route requires admin role and user is not admin
  if (requiresAdmin && !isAdmin) {
    // Redirect to main app (403 Forbidden equivalent)
    next({ name: 'app' })
    return
  }
  
  // If user is authenticated and trying to access public auth pages
  if (isPublic && isAuthenticated && to.name !== 'two-factor') {
    // Redirect to main app
    next({ name: 'app' })
    return
  }
  
  // Allow navigation
  next()
})

export default router
