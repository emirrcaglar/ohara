import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/library',
  },
  {
    path: '/library',
    component: () => import('../views/LibraryView.vue'),
  },
  {
    path: '/reader',
    component: () => import('../views/ReaderView.vue'),
  },
  {
    path: '/logs',
    component: () => import('../views/LogsView.vue'),
    meta: { requiresAdmin: true },
  },
  {
    path: '/login',
    component: () => import('../views/LoginView.vue'),
    meta: { fullscreen: true, public: true },
  },
  {
    path: '/register',
    component: () => import('../views/RegisterView.vue'),
    meta: { fullscreen: true, public: true },
  },
  {
    path: '/network',
    component: () => import('../views/LibraryView.vue'),
    meta: { requiresAdmin: true },
  },
  {
    path: '/admin/approvals',
    component: () => import('../views/AdminView.vue'),
    meta: { requiresAdmin: true },
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore()

  if (authStore.isInitializing) {
    await authStore.checkAuth()
  }

  if (!to.meta.public && !authStore.isAuthenticated) {
    return '/login'
  }

  if (to.meta.requiresAdmin && authStore.user?.role !== 'admin') {
    return '/'
  }

  if ((to.path === '/login' || to.path === '/register') && authStore.isAuthenticated) {
    return '/'
  }
})

export default router
