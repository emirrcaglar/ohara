<script setup lang="ts">
defineProps<{
  name: string
  progress: number
  sizeInfo?: string
  status: 'queued' | 'active' | 'complete' | 'paused' | 'failed' | 'exists'
  eta?: string
  speed?: string
  storagePath?: string
  canMoveUp?: boolean
  canMoveDown?: boolean
  needsFile?: boolean
  canResume?: boolean
}>()

defineEmits<{
  cancel: []
  moveUp: []
  moveDown: []
  pause: []
  resume: []
}>()
</script>

<template>
  <div class="space-y-3" :class="{ 'opacity-60': status === 'complete' }">
    <div class="flex justify-between items-start">
      <div class="overflow-hidden">
        <p class="text-xs font-black truncate uppercase tracking-tight text-on-surface">
          {{ name }}
        </p>
        <p v-if="status === 'exists'" class="text-[10px] text-on-surface-variant">Already exists</p>
        <p v-else-if="status === 'complete'" class="text-[10px] text-secondary-fixed-dim font-mono">
          SUCCESS_COMPLETED
        </p>
        <p v-else-if="status === 'failed'" class="text-[10px] text-error font-mono">FAILED</p>
        <p v-else-if="sizeInfo" class="text-[10px] text-on-surface-variant font-mono">
          {{ sizeInfo }}
        </p>
      </div>

      <div class="flex items-center gap-2">
        <button
          v-if="!needsFile && canResume"
          class="shrink-0 text-[10px] font-bold text-primary hover:underline"
          type="button"
          @click="$emit('resume')"
        >
          {{ status === 'queued' ? 'Upload' : status === 'failed' ? 'Retry' : 'Resume' }}
        </button>
        <button
          v-else-if="!needsFile && (status === 'active' || status === 'queued')"
          class="material-symbols-outlined text-[16px] text-on-surface-variant hover:text-primary"
          type="button"
          aria-label="Pause upload"
          @click="$emit('pause')"
        >
          pause
        </button>
        <button
          v-if="status !== 'exists'"
          class="h-8 w-8 text-xl leading-none font-black text-on-surface-variant hover:text-primary disabled:opacity-20 disabled:hover:text-on-surface-variant"
          type="button"
          aria-label="Move transfer up"
          :disabled="!canMoveUp"
          @click="$emit('moveUp')"
        >
          ↑
        </button>
        <button
          v-if="status !== 'exists'"
          class="h-8 w-8 text-xl leading-none font-black text-on-surface-variant hover:text-primary disabled:opacity-20 disabled:hover:text-on-surface-variant"
          type="button"
          aria-label="Move transfer down"
          :disabled="!canMoveDown"
          @click="$emit('moveDown')"
        >
          ↓
        </button>
        <button
          v-if="status !== 'complete'"
          class="material-symbols-outlined text-[16px] text-on-surface-variant hover:text-error"
          type="button"
          aria-label="Cancel transfer"
          @click="$emit('cancel')"
        >
          close
        </button>
      </div>
    </div>

    <div v-if="status !== 'exists'" class="h-1 w-full bg-surface-container-highest">
      <div
        class="h-full transition-all duration-500 ease-out"
        :class="status === 'complete' ? 'bg-secondary-container' : 'bg-primary-container'"
        :style="{ width: status === 'complete' ? '100%' : progress + '%' }"
      ></div>
    </div>

    <div v-if="status === 'complete'" class="flex justify-end">
      <span v-if="storagePath" class="text-[9px] font-mono text-secondary-fixed-dim uppercase">
        STORED: {{ storagePath }}
      </span>
    </div>
    <div v-else-if="status !== 'exists'" class="flex justify-between items-center">
      <span class="text-[9px] font-mono text-on-surface-variant">
        {{ status === 'active' ? 'ETA' : 'STATUS' }}: {{ eta }}
      </span>
      <span class="text-[9px] font-mono text-primary">{{ speed }}</span>
    </div>
  </div>
</template>
