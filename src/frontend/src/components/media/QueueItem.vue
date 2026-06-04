<script setup lang="ts">
import type { AudioRow } from '../../types/api'

defineProps<{
  track: AudioRow
  isActive?: boolean
}>()

const emit = defineEmits<{
  select: [track: AudioRow]
}>()
</script>

<template>
  <div class="group flex gap-4 p-4 transition-colors cursor-pointer border-l-4"
    :class="isActive ? 'bg-surface-container-high border-primary' : 'hover:bg-surface-container-highest border-transparent'"
    @click="emit('select', track)">

    <div class="w-12 h-12 bg-surface-container-lowest shrink-0">
      <div class="w-full h-full bg-gradient-to-br from-surface-container-lowest to-surface-container-high flex items-center justify-center">
        <span class="material-symbols-outlined text-outline-variant">graphic_eq</span>
      </div>
    </div>

    <div class="min-w-0">
      <p class="text-xs font-black uppercase tracking-tight truncate" :class="isActive ? 'text-primary' : 'text-on-surface'">
        {{ track.title }}
      </p>
      <p class="text-[10px] text-white/40 uppercase">{{ track.artist || track.album }}</p>
    </div>

    <div class="ml-auto flex items-center">
      <span v-if="isActive" class="material-symbols-outlined text-secondary text-sm">equalizer</span>
      <span v-else class="material-symbols-outlined text-white/40 text-sm opacity-0 group-hover:opacity-100">drag_handle</span>
    </div>
  </div>
</template>
