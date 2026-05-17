<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { fetchJson } from '../api/client'

interface PendingUser {
  id: number
  username: string
  role: string
  createdAt: string
}

const pendingUsers = ref<PendingUser[]>([])
const isLoading = ref(true)

async function fetchPendingUsers() {
  isLoading.value = true
  try {
    const users = await fetchJson<PendingUser[] | null>('/api/admin/users/pending')
    pendingUsers.value = Array.isArray(users) ? users : []
  } catch (error) {
    console.error('Failed to fetch pending users:', error)
    pendingUsers.value = []
  } finally {
    isLoading.value = false
  }
}

async function approveUser(id: number) {
  try {
    await fetchJson(`/api/admin/users/${id}/approve`, {
      method: 'POST',
    })
    await fetchPendingUsers()
  } catch (error) {
    console.error('Failed to approve user:', error)
    alert('Failed to approve user')
  }
}

onMounted(() => {
  fetchPendingUsers()
})
</script>

<template>
  <main class="p-8 flex flex-col gap-8 flex-1 bg-surface-container-lowest min-h-screen">
    <section class="flex-1 space-y-6">
      <div class="flex justify-between items-end border-b-0">
        <div>
          <h1
            class="font-display text-4xl font-black text-on-surface tracking-tighter uppercase leading-none"
          >
            REGISTRATION_QUEUE
          </h1>
          <p class="font-body text-secondary text-xs font-bold tracking-[0.2em] mt-2 uppercase">
            PENDING_AUTHORIZATIONS
          </p>
        </div>
      </div>

      <div class="bg-surface border-0 shadow-none overflow-hidden">
        <table class="w-full text-left font-body">
          <thead>
            <tr
              class="bg-surface-container-high text-on-surface-variant text-[10px] uppercase tracking-widest"
            >
              <th class="p-4 font-normal">USER_ID</th>
              <th class="p-4 font-normal">ROLE</th>
              <th class="p-4 font-normal">TIMESTAMP</th>
              <th class="p-4 font-normal text-right">ACTIONS</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-surface-container-highest/20">
            <tr v-if="isLoading" class="animate-pulse">
              <td
                colspan="4"
                class="p-8 text-center text-on-surface-variant text-xs uppercase tracking-widest"
              >
                Loading_Queue...
              </td>
            </tr>
            <tr v-else-if="pendingUsers.length === 0">
              <td
                colspan="4"
                class="p-8 text-center text-on-surface-variant text-xs uppercase tracking-widest"
              >
                No_Pending_Authorizations
              </td>
            </tr>
            <tr
              v-for="user in pendingUsers"
              :key="user.id"
              class="hover:bg-surface-container-low transition-colors group cursor-pointer"
            >
              <td class="p-4 text-primary font-bold text-sm tracking-tighter">
                {{ user.username }}
              </td>
              <td class="p-4">
                <span
                  class="text-secondary-container text-[10px] font-bold px-2 py-0.5 border border-secondary-container uppercase"
                >
                  {{ user.role }}
                </span>
              </td>
              <td class="p-4 text-on-surface-variant text-xs">
                {{ new Date(user.createdAt).toLocaleString() }}
              </td>
              <td class="p-4 text-right">
                <div class="flex justify-end gap-2">
                  <button
                    @click="approveUser(user.id)"
                    class="bg-primary-container text-on-primary-container px-3 py-1 text-[10px] font-black uppercase hover:bg-primary transition-colors"
                  >
                    APPROVE
                  </button>
                  <!-- Reject could be implemented later if needed -->
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </main>
</template>

<style scoped>
.font-display {
  font-family: 'Space Grotesk', sans-serif;
}
.font-body {
  font-family: 'Space Grotesk', sans-serif;
}
</style>
