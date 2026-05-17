<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { Library, Network, Terminal, ShieldCheck } from 'lucide-vue-next'
import { useAuthStore } from '../stores/auth'
import StatusBar from './StatusBar.vue'
import LogoutModal from './LogoutModal.vue'

defineOptions({
  name: 'AppSidebar',
})

const props = defineProps<{ open?: boolean }>()
const emit = defineEmits(['close'])

const route = useRoute()
const authStore = useAuthStore()

const navItems = [
  { name: 'Library', icon: Library, path: '/library' },
  { name: 'Network', icon: Network, path: '/network', adminOnly: true },
  { name: 'Logs', icon: Terminal, path: '/logs', adminOnly: true },
]

const showLogoutModal = ref(false)
</script>

<template>
  <aside
    class="fixed left-0 top-0 h-full flex flex-col bg-surface-container-low z-40 w-64 border-0 rounded-0 transition-transform duration-200"
    :class="props.open ? 'translate-x-0' : '-translate-x-full md:translate-x-0'"
  >
    <RouterLink
      to="/"
      class="p-8 flex flex-col gap-1 group cursor-pointer hover:opacity-80 transition-opacity"
    >
      <span class="text-2xl font-bold tracking-tighter text-primary-container"> OHARA </span>
    </RouterLink>

    <nav class="flex-1 mt-4">
      <RouterLink
        v-for="item in navItems"
        v-show="!item.adminOnly || authStore.user?.role === 'admin'"
        :key="item.name"
        :to="item.path"
        class="flex items-center gap-4 px-6 py-4 transition-none"
        :class="
          route.path === item.path
            ? 'text-primary-container font-black border-l-4 border-primary-container bg-surface-container-high'
            : 'text-white/60 hover:bg-surface-container-high hover:text-primary'
        "
      >
        <component :is="item.icon" class="w-5 h-5" />
        <span class="uppercase tracking-tight text-sm">{{ item.name }}</span>
      </RouterLink>

      <div v-if="authStore.user?.role === 'admin'" class="mt-8 px-4">
        <RouterLink
          to="/admin/approvals"
          class="w-full py-4 px-4 bg-secondary-container text-on-secondary-container font-bold uppercase tracking-tighter text-sm text-left flex justify-between items-center group transition-transform active:translate-x-1"
        >
          <div class="flex items-center gap-3">
            <ShieldCheck class="w-4 h-4" />
            USER_APPROVALS
          </div>
          <span class="material-symbols-outlined group-hover:translate-x-1 transition-transform">
            chevron_right
          </span>
        </RouterLink>
      </div>
    </nav>

    <button
      type="button"
      class="p-4 bg-surface-container border-t border-surface-container-high flex items-center gap-3 w-full text-left hover:bg-surface-container-high transition-colors"
      @click="showLogoutModal = true"
    >
      <div class="w-10 h-10 bg-surface-container-highest flex items-center justify-center shrink-0">
        <span class="material-symbols-outlined" style="color: #ff8c00">person</span>
      </div>
      <div class="flex-1 overflow-hidden">
        <div class="truncate font-bold text-[10px] text-on-surface uppercase tracking-widest">
          {{ authStore.user?.username || 'GUEST_USER' }}
        </div>
        <div class="text-[9px] text-secondary opacity-80 uppercase">
          {{ authStore.user?.role || 'CONNECTED' }}
        </div>
      </div>
    </button>

    <StatusBar />

    <LogoutModal :open="showLogoutModal" @close="showLogoutModal = false" />
  </aside>
</template>
