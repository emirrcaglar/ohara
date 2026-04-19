<script setup lang="ts">
import MediaControls from './MediaControls.vue'

defineProps<{
  title: string
  subLabel: string
  bitrate?: string
  isPlaying: boolean
}>()

const emit = defineEmits<{
  play: []
  pause: []
  next: []
  previous: []
  seek: [time: number]
  volumeChange: [volume: number]
}>()
</script>

<template>
  <div class="bg-surface-container-low border-t border-white/10 px-4 py-2">
    <div class="flex items-center gap-4">
      <div class="min-w-0 w-56">
        <p class="text-[10px] font-mono uppercase tracking-widest text-secondary mb-1">Now Playing</p>
        <p class="truncate text-sm font-black uppercase tracking-tight text-on-surface">{{ title }}</p>
        <p class="truncate text-[10px] font-mono uppercase tracking-widest text-on-surface-variant">{{ subLabel }}</p>
      </div>

      <div class="flex-1 flex items-center justify-center min-w-0">
        <MediaControls
          :currentTime="'0:00'"
          :totalTime="'0:00'"
          :progress="0"
          :volume="75"
          :isPlaying="isPlaying"
          @play="emit('play')"
          @pause="emit('pause')"
          @next="emit('next')"
          @previous="emit('previous')"
          @volumeChange="(v) => emit('volumeChange', v)"
        />
      </div>

      <div class="w-36 text-right pr-1">
        <span class="text-[10px] font-mono text-secondary uppercase tracking-widest">{{ bitrate || 'Audio' }}</span>
      </div>
    </div>
  </div>
</template>
