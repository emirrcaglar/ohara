import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

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
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

export default router
