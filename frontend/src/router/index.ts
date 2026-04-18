import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    component: () => import('../views/HomeView.vue')
  },
  {
    path: '/media',
    component: () => import('../views/MediaView.vue')
  },
  {
    path: '/library',
    component: () => import('../views/LibraryView.vue')
  },
  {
    path: '/uploads',
    component: () => import('../views/UploadsView.vue')
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

export default router
