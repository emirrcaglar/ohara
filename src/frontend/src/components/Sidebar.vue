<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Library, Settings, Terminal, ShieldCheck } from 'lucide-vue-next'
import { useAuthStore } from '../stores/auth'
import { getUserPfpUrl } from '../utils/userPfp'
import UserModal from './UserModal.vue'

defineOptions({
  name: 'AppSidebar',
})

const props = defineProps<{ open?: boolean }>()
const emit = defineEmits<{ close: [] }>()

const route = useRoute()
const authStore = useAuthStore()

const navItems = [
  { name: 'Library', icon: Library, path: '/library' },
  { name: 'Settings', icon: Settings, path: '/settings', adminOnly: true },
  { name: 'Logs', icon: Terminal, path: '/logs', adminOnly: true },
]

const showUserModal = ref(false)
const userPfpUrl = computed(() => getUserPfpUrl(authStore.user?.pfp))
</script>

<template>
  <aside
    class="fixed left-0 top-[env(safe-area-inset-top)] bottom-[env(safe-area-inset-bottom)] h-auto flex flex-col bg-surface-container-low z-40 w-64 border-0 rounded-0 transition-transform duration-200 md:top-0 md:bottom-0"
    :class="props.open ? 'translate-x-0' : '-translate-x-full md:translate-x-0'"
  >
    <RouterLink
      to="/"
      class="p-8 flex flex-col gap-1 group cursor-pointer hover:opacity-80 transition-opacity"
      @click="emit('close')"
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
        @click="emit('close')"
        :class="
          route.path === item.path
            ? 'text-primary-container font-black border-l-4 border-primary-container bg-surface-container-high'
            : 'text-white/60 hover:bg-surface-container-high hover:text-primary'
        "
      >
        <component :is="item.icon" class="w-5 h-5" />
        <span class="uppercase tracking-tight text-sm">{{ item.name }}</span>
      </RouterLink>
    </nav>

    <div v-if="authStore.user?.role === 'admin'" class="px-4 pb-3">
      <RouterLink
        to="/admin/approvals"
        class="w-full py-4 px-4 bg-surface-container-high text-primary-container font-bold uppercase tracking-tighter text-sm text-left flex justify-between items-center group transition-transform active:translate-x-1 hover:bg-surface-container-highest"
        @click="emit('close')"
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

    <button
      type="button"
      class="p-4 bg-surface-container border-t border-surface-container-high flex items-center gap-3 w-full text-left hover:bg-surface-container-high transition-colors"
      @click="showUserModal = true"
    >
      <div
        class="w-10 h-10 bg-surface-container-highest flex items-center justify-center shrink-0 overflow-hidden"
      >
        <img
          v-if="userPfpUrl"
          :src="userPfpUrl"
          class="w-full h-full object-cover grayscale brightness-75 contrast-125"
          alt="User avatar"
        />
        <span v-else class="material-symbols-outlined" style="color: #ff8c00">person</span>
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

    <UserModal :open="showUserModal" @close="showUserModal = false" />
  </aside>
</template>
