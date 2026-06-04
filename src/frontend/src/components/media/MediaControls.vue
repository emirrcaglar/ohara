<script setup lang="ts">
const props = defineProps<{
  currentTime: string
  totalTime: string
  progress: number
  volume: number
  duration: number
  canSkip: boolean
  isPlaying: boolean
  expanded?: boolean
}>()

const emit = defineEmits<{
  play: []
  pause: []
  seek: [time: number]
  volumeChange: [volume: number]
  next: []
  previous: []
}>()

function handleSeek(event: Event) {
  const target = event.target as HTMLInputElement
  const progress = Number(target.value)
  const time = props.duration > 0 ? (progress / 100) * props.duration : 0
  emit('seek', time)
}

function handleVolumeChange(event: Event) {
  const target = event.target as HTMLInputElement
  emit('volumeChange', Number(target.value))
}
</script>

<template>
  <div class="bg-transparent px-2 py-1 flex flex-col gap-2 md:gap-3 w-full">
    <div :class="expanded ? '' : 'hidden md:block'" class="space-y-1.5">
      <div
        class="flex justify-between font-mono text-[10px] uppercase tracking-widest text-secondary"
      >
        <span>{{ currentTime }} / {{ totalTime }}</span>
      </div>
      <div class="relative -my-2 py-2">
        <input
          class="media-progress-slider relative z-10 h-5 w-full cursor-pointer appearance-none bg-transparent"
          :disabled="duration <= 0"
          :max="100"
          :min="0"
          :value="progress"
          :aria-label="'Seek through track'"
          :style="{ '--progress': progress + '%' }"
          step="0.1"
          type="range"
          @input="handleSeek"
        />
        <div
          class="pointer-events-none absolute inset-x-0 top-1/2 h-1 -translate-y-1/2 bg-surface-container-highest"
        >
          <div
            class="absolute h-full bg-primary-container"
            :style="{ width: progress + '%' }"
          ></div>
        </div>
      </div>
    </div>

    <div class="flex items-center justify-between gap-2 md:gap-4">
      <div class="flex items-center gap-2 md:gap-5">
        <button
          :class="expanded ? '' : 'hidden md:block'"
          class="text-white/60 hover:text-secondary"
        >
          <span class="material-symbols-outlined">shuffle</span>
        </button>
        <button
          class="text-primary hover:text-primary-container scale-125 disabled:text-white/25 disabled:hover:text-white/25 disabled:cursor-not-allowed"
          :disabled="!canSkip"
          @click="emit('previous')"
        >
          <span class="material-symbols-outlined">skip_previous</span>
        </button>
        <button
          class="w-10 h-10 md:w-14 md:h-14 bg-primary-container text-on-primary-container flex items-center justify-center hover:bg-primary transition-transform active:scale-95"
          @click="isPlaying ? emit('pause') : emit('play')"
        >
          <span
            class="material-symbols-outlined text-2xl md:text-3xl"
            style="font-variation-settings: 'FILL' 1"
            >{{ isPlaying ? 'pause' : 'play_arrow' }}</span
          >
        </button>
        <button
          class="text-primary hover:text-primary-container scale-125 disabled:text-white/25 disabled:hover:text-white/25 disabled:cursor-not-allowed"
          :disabled="!canSkip"
          @click="emit('next')"
        >
          <span class="material-symbols-outlined">skip_next</span>
        </button>
        <button
          :class="expanded ? '' : 'hidden md:block'"
          class="text-white/60 hover:text-secondary"
        >
          <span class="material-symbols-outlined">repeat</span>
        </button>
      </div>

      <div :class="expanded ? 'flex' : 'hidden md:flex'" class="items-center gap-4">
        <div class="flex items-center gap-3">
          <span class="material-symbols-outlined text-sm text-secondary">volume_up</span>
          <div class="relative -my-2 py-2 w-28">
            <input
              class="media-volume-slider relative z-10 h-5 w-full cursor-pointer appearance-none bg-transparent"
              :max="100"
              :min="0"
              :value="volume"
              aria-label="Adjust volume"
              step="1"
              type="range"
              @input="handleVolumeChange"
            />
            <div
              class="pointer-events-none absolute inset-x-0 top-1/2 h-1 -translate-y-1/2 bg-surface-container-highest"
            >
              <div class="absolute h-full bg-secondary" :style="{ width: volume + '%' }"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.media-progress-slider::-webkit-slider-runnable-track,
.media-volume-slider::-webkit-slider-runnable-track {
  height: 0.25rem;
  background: transparent;
}

.media-progress-slider::-webkit-slider-thumb,
.media-volume-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 0.75rem;
  height: 0.75rem;
  border-radius: 9999px;
  background: #fff;
  margin-top: 0rem;
  box-shadow: 0 0 10px rgba(255, 140, 0, 0.5);
}

.media-progress-slider::-moz-range-track,
.media-volume-slider::-moz-range-track {
  height: 0.25rem;
  background: transparent;
}

.media-progress-slider::-moz-range-thumb,
.media-volume-slider::-moz-range-thumb {
  width: 0.75rem;
  height: 0.75rem;
  border: 0;
  border-radius: 9999px;
  background: #fff;
  box-shadow: 0 0 10px rgba(255, 140, 0, 0.5);
}

.media-progress-slider:disabled,
.media-volume-slider:disabled {
  cursor: default;
}
</style>
