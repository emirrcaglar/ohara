import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    redirect: '/library'
  },
  {
    path: '/library',
    component: () => import('../views/LibraryView.vue')
  },
  {
    path: '/media',
    component: () => import('../views/MediaView.vue')
  },
  {
    path: '/reader',
    component: () => import('../views/ReaderView.vue')
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
