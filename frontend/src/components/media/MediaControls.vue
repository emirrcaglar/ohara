<script setup lang="ts">
defineProps<{
  currentTime: string
  totalTime: string
  progress: number
  volume: number
  isPlaying: boolean
}>()

const emit = defineEmits<{
  play: []
  pause: []
  seek: [time: number]
  volumeChange: [volume: number]
  next: []
  previous: []
}>()
</script>

<template>
  <div class="bg-transparent px-2 py-1 flex flex-col gap-3">
    <div class="space-y-1.5">
      <div class="flex justify-between font-mono text-[10px] uppercase tracking-widest text-secondary">
        <span>{{ currentTime }}</span>
        <span>{{ totalTime }}</span>
      </div>
      <div class="h-1 bg-surface-container-highest relative">
        <div class="absolute h-full bg-primary-container" :style="{ width: progress + '%' }"></div>
        <div class="absolute h-4 w-1 bg-white -top-1.5 shadow-[0_0_10px_rgba(255,140,0,0.5)]" :style="{ left: progress + '%' }"></div>
      </div>
    </div>

    <div class="flex items-center justify-between gap-4">
      <div class="flex items-center gap-5">
        <button class="text-white/60 hover:text-secondary"><span class="material-symbols-outlined">shuffle</span></button>
        <button class="text-primary hover:text-primary-container scale-125" @click="emit('previous')"><span class="material-symbols-outlined">skip_previous</span></button>
        <button class="w-14 h-14 bg-primary-container text-on-primary-container flex items-center justify-center hover:bg-primary transition-transform active:scale-95" @click="isPlaying ? emit('pause') : emit('play')">
          <span class="material-symbols-outlined text-3xl" style="font-variation-settings: 'FILL' 1;">{{ isPlaying ? 'pause' : 'play_arrow' }}</span>
        </button>
        <button class="text-primary hover:text-primary-container scale-125" @click="emit('next')"><span class="material-symbols-outlined">skip_next</span></button>
        <button class="text-white/60 hover:text-secondary"><span class="material-symbols-outlined">repeat</span></button>
      </div>

      <div class="flex items-center gap-4">
        <div class="flex items-center gap-3">
          <span class="material-symbols-outlined text-sm text-secondary">volume_up</span>
          <div class="w-20 h-1 bg-surface-container-highest">
            <div class="h-full bg-secondary" :style="{ width: volume + '%' }"></div>
          </div>
        </div>
        <button class="px-3 py-1.5 bg-surface-container-high border border-outline-variant hover:border-primary text-[10px] font-black uppercase tracking-tighter">
          Cast Device
        </button>
      </div>
    </div>
  </div>
</template>
