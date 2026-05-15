import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchJson } from '../api/client'

export interface User {
  id: number
  username: string
  role: string
  isApproved: boolean
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const isAuthenticated = ref(false)
  const isInitializing = ref(true)

  async function checkAuth() {
    try {
      const userData = await fetchJson<User>('/api/auth/me')
      user.value = userData
      isAuthenticated.value = true
    } catch (error) {
      user.value = null
      isAuthenticated.value = false
    } finally {
      isInitializing.value = false
    }
  }

  async function login(username: string, password: string) {
    await fetchJson('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
    await checkAuth()
  }

  async function logout() {
    await fetchJson('/api/auth/logout', { method: 'POST' })
    user.value = null
    isAuthenticated.value = false
  }

  return {
    user,
    isAuthenticated,
    isInitializing,
    checkAuth,
    login,
    logout,
  }
})
