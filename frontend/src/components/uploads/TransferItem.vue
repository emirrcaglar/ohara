<script setup lang="ts">
defineProps<{
  name: string
  progress: number
  sizeInfo?: string
  status: 'active' | 'complete' | 'paused'
  eta?: string
  speed?: string
  storagePath?: string
}>()
</script>

<template>
  <div class="space-y-3" :class="{ 'opacity-60': status === 'complete' }">
    <div class="flex justify-between items-start">
      <div class="overflow-hidden">
        <p class="text-xs font-black truncate uppercase tracking-tight text-on-surface">
          {{ name }}
        </p>
        <p v-if="status === 'complete'" class="text-[10px] text-secondary-fixed-dim font-mono">
          SUCCESS_COMPLETED
        </p>
        <p v-else-if="sizeInfo" class="text-[10px] text-on-surface-variant font-mono">
          {{ sizeInfo }}
        </p>
      </div>

      <span
        class="material-symbols-outlined text-[14px] text-secondary"
        :style="status === 'complete' ? 'font-variation-settings: \'FILL\' 1;' : ''"
      >
        {{ status === 'complete' ? 'check_circle' : 'cloud_upload' }}
      </span>
    </div>

    <div class="h-1 w-full bg-surface-container-highest">
      <div
        class="h-full transition-all duration-500 ease-out"
        :class="status === 'complete' ? 'bg-secondary-container' : 'bg-primary-container'"
        :style="{ width: status === 'complete' ? '100%' : progress + '%' }"
      ></div>
    </div>

    <div v-if="status === 'complete'" class="flex justify-end">
      <span class="text-[9px] font-mono text-secondary-fixed-dim uppercase">
        STORED: {{ storagePath }}
      </span>
    </div>
    <div v-else class="flex justify-between items-center">
      <span class="text-[9px] font-mono text-on-surface-variant">ETA: {{ eta }}</span>
      <span class="text-[9px] font-mono text-primary">{{ speed }}</span>
    </div>
  </div>
</template>
