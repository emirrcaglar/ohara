<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { fetchLatestDeployment } from '../api/deployments'

const latestDeploymentAt = ref<Date | null>(null)
const now = ref(new Date())
let timer: number | undefined

const uptimeLabel = computed(() => {
  if (!latestDeploymentAt.value) return 'NYE'

  const diffMs = Math.max(now.value.getTime() - latestDeploymentAt.value.getTime(), 0)
  const totalMinutes = Math.floor(diffMs / 60000)
  const hours = Math.floor(totalMinutes / 60)
  const minutes = totalMinutes % 60

  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`
})

async function loadLatestDeployment() {
  try {
    const deployment = await fetchLatestDeployment()
    latestDeploymentAt.value = deployment ? new Date(deployment.deployedAt) : null
    now.value = new Date()
  } catch {
    latestDeploymentAt.value = null
  }
}

onMounted(() => {
  void loadLatestDeployment()
  timer = window.setInterval(() => {
    now.value = new Date()
  }, 60000)
})

onUnmounted(() => {
  if (timer !== undefined) window.clearInterval(timer)
})
</script>

<template>
  <footer
    class="hidden md:flex h-8 bg-surface-container-lowest px-4 items-center justify-between text-[7px] uppercase tracking-[0.3em] font-bold text-white/20 overflow-hidden"
  >
    <div class="flex items-center gap-4">
      <div class="flex items-center gap-1.5 text-primary-container">
        <span>UP: {{ uptimeLabel }}</span>
      </div>
    </div>
    <span>© 2026 OHARA</span>
  </footer>
</template>
