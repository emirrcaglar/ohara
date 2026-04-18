<script setup lang="ts">
import MediaDisplay from './MediaDisplay.vue'
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
  <div class="bg-surface-container-low border-t border-white/10 px-6 py-4">
    <div class="flex items-center gap-6">
      <MediaDisplay
        :title="title"
        :subLabel="subLabel"
        imageUrl=""
        :bitrate="bitrate || ''"
        :isPlaying="isPlaying"
        @play="emit('play')"
      />

      <div class="flex-1 flex items-center justify-center">
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

      <div class="w-48 text-right">
        <span class="text-xs font-mono text-secondary uppercase tracking-widest">Now Playing</span>
      </div>
    </div>
  </div>
</template>
