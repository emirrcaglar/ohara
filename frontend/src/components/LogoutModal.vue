<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: [] }>()

const authStore = useAuthStore()
const router = useRouter()

const currentTime = ref('')
let clockInterval: ReturnType<typeof setInterval> | null = null

function updateTime() {
  const now = new Date()
  const hh = String(now.getUTCHours()).padStart(2, '0')
  const mm = String(now.getUTCMinutes()).padStart(2, '0')
  const ss = String(now.getUTCSeconds()).padStart(2, '0')
  currentTime.value = `${hh}:${mm}:${ss} UTC`
}

async function handleLogout() {
  await authStore.logout()
  emit('close')
  router.push('/login')
}

onMounted(() => {
  updateTime()
  clockInterval = setInterval(updateTime, 1000)
})

onUnmounted(() => {
  if (clockInterval) clearInterval(clockInterval)
})
</script>

<template>
  <Teleport to="body">
    <div
      v-if="props.open"
      class="fixed inset-0 bg-background/80 backdrop-blur-sm z-60 flex items-end justify-start p-6"
      @click.self="emit('close')"
    >
      <div
        class="w-72 bg-surface-container-lowest border-2 border-primary-container p-6 shadow-[0px_0px_40px_-10px_rgba(255,140,0,0.4)] relative"
      >
        <div class="flex items-center gap-4 mb-8">
          <div
            class="relative w-14 h-14 bg-surface-container-high border border-outline-variant flex items-center justify-center shrink-0"
          >
            <span class="material-symbols-outlined text-4xl" style="color: #ff8c00">person</span>
            <div class="absolute -bottom-1 -right-1 w-3 h-3 bg-primary-container"></div>
          </div>
          <div>
            <div class="text-[10px] text-secondary font-bold tracking-widest mb-1">
              TERMINATING_SESSION
            </div>
            <div class="text-lg font-black leading-none text-on-surface tracking-tighter">
              {{ authStore.user?.username || 'GUEST' }}
            </div>
          </div>
        </div>

        <div class="space-y-3 mb-8">
          <div class="flex justify-between text-[10px] border-b border-surface-container-high pb-1">
            <span class="opacity-60 uppercase">System Time</span>
            <span class="font-mono">{{ currentTime }}</span>
          </div>
          <div class="flex justify-between text-[10px] border-b border-surface-container-high pb-1">
            <span class="opacity-60 uppercase">IP Address</span>
            <span class="font-mono">127.0.0.1</span>
          </div>
        </div>

        <div class="flex flex-col gap-2">
          <button
            type="button"
            class="w-full bg-primary-container text-on-primary-container font-black py-4 flex items-center justify-center gap-3 active:scale-95 transition-transform"
            @click="handleLogout"
          >
            <span class="material-symbols-outlined">logout</span>
            <span class="tracking-widest">LOGOUT</span>
          </button>
        </div>

        <div
          class="absolute -top-1 -left-1 w-4 h-4 border-t-2 border-l-2 border-primary-container"
        ></div>
        <div class="absolute -top-1 -right-1 w-2 h-2 bg-primary-container"></div>
        <div class="absolute bottom-0 right-0 p-1">
          <div class="text-[8px] opacity-20 font-mono">OHARA_SYS_EXIT</div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
