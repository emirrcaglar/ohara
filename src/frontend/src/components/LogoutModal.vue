<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const props = defineProps<{ open: boolean }>()
const emit = defineEmits<{ close: [] }>()

const authStore = useAuthStore()
const router = useRouter()

const currentTime = ref('')
const showPasswordModal = ref(false)
const currentPassword = ref('')
const newPassword = ref('')
const confirmNewPassword = ref('')
const passwordError = ref('')
const passwordSuccess = ref('')
const isChangingPassword = ref(false)
let clockInterval: ReturnType<typeof setInterval> | null = null

function updateTime() {
  const now = new Date()
  const hh = String(now.getUTCHours()).padStart(2, '0')
  const mm = String(now.getUTCMinutes()).padStart(2, '0')
  const ss = String(now.getUTCSeconds()).padStart(2, '0')
  currentTime.value = `${hh}:${mm}:${ss} UTC`
}

function resetPasswordForm() {
  currentPassword.value = ''
  newPassword.value = ''
  confirmNewPassword.value = ''
  passwordError.value = ''
  passwordSuccess.value = ''
  isChangingPassword.value = false
}

function getErrorMessage(error: unknown) {
  if (error instanceof Error) return error.message
  if (typeof error === 'object' && error && 'message' in error) {
    return String(error.message)
  }
  return 'Password update failed'
}

function openPasswordModal() {
  resetPasswordForm()
  showPasswordModal.value = true
}

async function handlePasswordSubmit() {
  passwordError.value = ''
  passwordSuccess.value = ''

  if (!currentPassword.value || !newPassword.value || !confirmNewPassword.value) {
    passwordError.value = 'All password fields are required'
    return
  }

  if (newPassword.value !== confirmNewPassword.value) {
    passwordError.value = 'New passwords do not match'
    return
  }

  isChangingPassword.value = true
  try {
    await authStore.changePassword(currentPassword.value, newPassword.value)
    resetPasswordForm()
    passwordSuccess.value = 'Credential update accepted'
  } catch (error: unknown) {
    passwordError.value = getErrorMessage(error)
  } finally {
    isChangingPassword.value = false
  }
}

async function handleLogout() {
  await authStore.logout()
  emit('close')
  router.push('/login')
}

watch(
  () => props.open,
  (open) => {
    if (!open) {
      showPasswordModal.value = false
      resetPasswordForm()
    }
  },
)

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
      class="fixed inset-0 bg-background/80 backdrop-blur-sm z-60 flex flex-col-reverse items-start justify-start gap-4 overflow-y-auto p-6 md:flex-row md:items-end"
      @click.self="emit('close')"
    >
      <div
        class="w-full max-w-72 bg-surface-container-lowest border-2 border-primary-container p-6 shadow-[0px_0px_40px_-10px_rgba(255,140,0,0.4)] relative"
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
          <button
            type="button"
            class="w-full bg-transparent text-secondary border border-secondary py-3 flex items-center justify-center gap-3 active:scale-95 transition-transform mt-1"
            @click="openPasswordModal"
          >
            <span class="material-symbols-outlined text-sm font-bold">key</span>
            <span class="text-[10px] font-bold tracking-[0.2em]">CHANGE_PASSWORD</span>
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

      <form
        v-if="showPasswordModal"
        class="w-full max-w-80 bg-surface-container-lowest border-2 border-secondary p-6 shadow-[0px_0px_40px_-10px_rgba(255,177,196,0.5)] relative"
        @submit.prevent="handlePasswordSubmit"
      >
        <div class="flex items-center gap-3 mb-6 border-b border-surface-container-high pb-4">
          <span class="material-symbols-outlined text-secondary">shield_lock</span>
          <div class="text-lg font-black leading-none text-on-surface tracking-tighter uppercase">
            PASSWORD_UPDATE
          </div>
        </div>

        <div
          v-if="passwordError"
          class="mb-4 bg-error-container p-3 text-[10px] font-bold uppercase tracking-widest text-on-error-container"
        >
          {{ passwordError }}
        </div>
        <div
          v-if="passwordSuccess"
          class="mb-4 bg-primary-container p-3 text-[10px] font-bold uppercase tracking-widest text-on-primary-container"
        >
          {{ passwordSuccess }}
        </div>

        <div class="space-y-4 mb-8">
          <div class="space-y-1">
            <label
              class="text-[8px] text-secondary/90 uppercase font-bold tracking-[0.2em]"
              for="current-password"
            >
              OLD_PASSWORD
            </label>
            <input
              id="current-password"
              v-model="currentPassword"
              class="w-full bg-surface-container text-on-surface p-2 text-xs font-mono focus:ring-0 focus:outline-none password-field"
              name="current-password"
              placeholder="********"
              type="password"
              autocomplete="current-password"
              required
              :disabled="isChangingPassword"
            />
          </div>
          <div class="space-y-1">
            <label
              class="text-[8px] text-secondary/90 uppercase font-bold tracking-[0.2em]"
              for="new-password"
            >
              NEW_PASSWORD
            </label>
            <input
              id="new-password"
              v-model="newPassword"
              class="w-full bg-surface-container text-on-surface p-2 text-xs font-mono focus:ring-0 focus:outline-none password-field"
              name="new-password"
              placeholder="********"
              type="password"
              autocomplete="new-password"
              required
              :disabled="isChangingPassword"
            />
          </div>
          <div class="space-y-1">
            <label
              class="text-[8px] text-secondary/90 uppercase font-bold tracking-[0.2em]"
              for="confirm-new-password"
            >
              CONFIRM_NEW_PASSWORD
            </label>
            <input
              id="confirm-new-password"
              v-model="confirmNewPassword"
              class="w-full bg-surface-container text-on-surface p-2 text-xs font-mono focus:ring-0 focus:outline-none password-field"
              name="confirm-new-password"
              placeholder="********"
              type="password"
              autocomplete="new-password"
              required
              :disabled="isChangingPassword"
            />
          </div>
        </div>

        <button
          type="submit"
          class="w-full bg-secondary text-on-secondary font-black py-4 flex items-center justify-center gap-3 active:scale-95 transition-transform disabled:opacity-50 disabled:cursor-not-allowed"
          :disabled="isChangingPassword"
        >
          <span class="material-symbols-outlined font-bold">
            {{ isChangingPassword ? 'refresh' : 'sync_lock' }}
          </span>
          <span class="tracking-widest text-xs">
            {{ isChangingPassword ? 'SUBMITTING' : 'SUBMIT_CHANGES' }}
          </span>
        </button>

        <div class="absolute -top-1 -right-1 w-4 h-4 border-t-2 border-r-2 border-secondary"></div>
        <div class="absolute -bottom-1 -left-1 w-2 h-2 bg-secondary"></div>
        <div class="absolute bottom-0 right-0 p-1">
          <div class="text-[8px] opacity-50 font-mono">SEC_AUTH_MODULE_v3</div>
        </div>
      </form>
    </div>
  </Teleport>
</template>

<style scoped>
.password-field:focus {
  border-bottom: 2px solid #ffb1c4;
}
</style>
