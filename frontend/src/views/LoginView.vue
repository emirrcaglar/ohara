<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const systemId = ref('')
const accessKey = ref('')
const error = ref('')
const isLoading = ref(false)

async function handleSubmit() {
  error.value = ''
  isLoading.value = true
  try {
    await authStore.login(systemId.value, accessKey.value)
    router.push('/')
  } catch (err: any) {
    error.value = err.message || 'Authentication failed'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div
    class="bg-surface-container-lowest text-on-background min-h-screen flex items-center justify-center p-6 antialiased selection:bg-secondary-container selection:text-white"
  >
    <main class="w-full max-w-md z-10">
      <div class="flex flex-col items-center mb-12">
        <div
          class="relative w-24 h-24 mb-6 bg-surface-container flex items-center justify-center [filter:drop-shadow(0_0_15px_rgba(255,140,0,0.3))]"
        >
          <span class="material-symbols-outlined text-5xl" style="color: #ff8c00">terminal</span>
        </div>
        <h1
          class="text-4xl font-black tracking-[-0.07em] uppercase text-primary-container leading-none"
        >
          OHARA
        </h1>
      </div>

      <div
        class="bg-surface p-8 border-l-4 border-primary-container shadow-[20px_20px_0px_0px_rgba(28,27,27,1)]"
      >
        <div class="flex justify-between items-start mb-10">
          <div>
            <h2 class="text-xs font-bold uppercase tracking-tighter text-on-surface">
              System Authorization
            </h2>
            <div class="h-[2px] w-8 bg-secondary-container mt-1"></div>
          </div>
          <span class="text-[9px] font-mono text-secondary-fixed-dim bg-on-secondary py-0.5 px-1.5">
            SECURE_NODE_01
          </span>
        </div>

        <div
          v-if="error"
          class="mb-6 p-4 bg-error-container text-on-error-container text-xs font-bold uppercase border-l-4 border-error"
        >
          {{ error }}
        </div>

        <form class="space-y-8" @submit.prevent="handleSubmit">
          <div class="relative group">
            <label
              class="absolute -top-3 right-0 text-[10px] font-bold uppercase tracking-widest text-secondary-container bg-surface px-1 transition-all group-focus-within:text-primary"
              for="system-id"
            >
              USERNAME
            </label>
            <div class="flex items-center bg-surface-container-highest sharp-focus transition-all">
              <span class="material-symbols-outlined px-4 text-outline opacity-50 select-none">
                terminal
              </span>
              <input
                id="system-id"
                v-model="systemId"
                class="w-full bg-transparent border-0 py-4 pr-4 text-sm font-mono text-on-surface placeholder:text-surface-variant focus:ring-0 outline-none"
                name="system-id"
                placeholder="USERNAME"
                type="text"
                required
                :disabled="isLoading"
              />
            </div>
          </div>

          <div class="relative group">
            <label
              class="absolute -top-3 right-0 text-[10px] font-bold uppercase tracking-widest text-secondary-container bg-surface px-1 transition-all group-focus-within:text-primary"
              for="access-key"
            >
              PASSWORD
            </label>
            <div class="flex items-center bg-surface-container-highest sharp-focus transition-all">
              <span class="material-symbols-outlined px-4 text-outline opacity-50 select-none">
                key
              </span>
              <input
                id="access-key"
                v-model="accessKey"
                class="w-full bg-transparent border-0 py-4 pr-4 text-sm font-mono text-on-surface placeholder:text-surface-variant focus:ring-0 outline-none"
                name="access-key"
                placeholder="••••••••••••"
                type="password"
                required
                :disabled="isLoading"
              />
            </div>
          </div>

          <div class="pt-4 flex flex-col gap-4">
            <button
              type="submit"
              :disabled="isLoading"
              class="w-full bg-primary-container text-on-primary-container font-black py-4 uppercase tracking-tighter text-lg hover:bg-primary transition-all active:scale-[0.98] flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <template v-if="isLoading">
                AUTHENTICATING...
                <span class="material-symbols-outlined text-xl animate-spin">refresh</span>
              </template>
              <template v-else>
                AUTHENTICATE
                <span class="material-symbols-outlined text-xl">login</span>
              </template>
            </button>
            <div
              class="flex justify-between items-center text-[10px] font-bold uppercase tracking-tighter"
            >
              <a class="text-secondary hover:text-secondary-fixed-dim transition-colors" href="#">
                Forgot Credentials?
              </a>
              <span class="text-surface-variant">V1.0.4-KINETIC</span>
            </div>
          </div>
        </form>

        <div class="mt-12 grid grid-cols-2 gap-4 border-t-2 border-outline-variant/10 pt-8">
          <div class="bg-surface-container-low p-4 flex flex-col gap-1">
            <span class="text-[8px] font-bold text-surface-variant uppercase">Server Status</span>
            <div class="flex items-center gap-2">
              <span class="w-1.5 h-1.5 bg-secondary animate-pulse"></span>
              <span class="text-[10px] font-mono text-on-surface uppercase tracking-tighter"
                >SYNC_ACTIVE</span
              >
            </div>
          </div>
          <div class="bg-surface-container-low p-4 flex flex-col gap-1">
            <span class="text-[8px] font-bold text-surface-variant uppercase">Node Location</span>
            <div class="flex items-center gap-2 text-on-surface">
              <span class="material-symbols-outlined text-[14px]">language</span>
              <span class="text-[10px] font-mono uppercase tracking-tighter">EDGE_TOKYO_09</span>
            </div>
          </div>
        </div>
      </div>
    </main>

    <div
      class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] bg-primary-container/5 rounded-full blur-[120px] pointer-events-none z-0"
    ></div>
  </div>
</template>

<style scoped>
.sharp-focus:focus-within {
  border-bottom: 2px solid #ff8c00;
}
</style>
