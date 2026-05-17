<script setup lang="ts">
import { ref } from 'vue'
import MediaControls from './MediaControls.vue'

defineProps<{
  title: string
  subLabel: string
  currentTime: string
  totalTime: string
  progress: number
  volume: number
  duration: number
  canSkip: boolean
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

const expanded = ref(false)
</script>

<template>
  <!-- Mobile expanded overlay -->
  <Transition name="slide-up">
    <div
      v-if="expanded"
      class="md:hidden fixed inset-x-0 bottom-0 z-50 bg-surface-container-low border-t-2 border-primary-container"
    >
      <div class="flex justify-center pt-2 pb-1">
        <button
          class="text-on-surface-variant hover:text-on-surface p-1"
          aria-label="Collapse player"
          @click="expanded = false"
        >
          <span class="material-symbols-outlined">expand_more</span>
        </button>
      </div>

      <div class="px-6 py-3 border-b border-white/10">
        <p class="text-[10px] font-mono uppercase tracking-widest text-secondary mb-1">
          Now Playing
        </p>
        <p class="text-xl font-black uppercase tracking-tight text-on-surface truncate">
          {{ title }}
        </p>
        <p class="text-xs font-mono uppercase tracking-widest text-on-surface-variant truncate">
          {{ subLabel }}
        </p>
      </div>

      <div class="px-4 py-4">
        <MediaControls
          :currentTime="currentTime"
          :totalTime="totalTime"
          :progress="progress"
          :volume="volume"
          :duration="duration"
          :canSkip="canSkip"
          :isPlaying="isPlaying"
          :expanded="true"
          @play="emit('play')"
          @pause="emit('pause')"
          @seek="(time) => emit('seek', time)"
          @next="emit('next')"
          @previous="emit('previous')"
          @volumeChange="(v) => emit('volumeChange', v)"
        />
      </div>
    </div>
  </Transition>

  <!-- Compact bar -->
  <div class="bg-surface-container-low border-t border-white/10 px-4 py-2">
    <div class="flex items-center gap-4">
      <!-- Track info — tappable on mobile to expand -->
      <button
        class="min-w-0 w-32 md:w-56 text-left flex items-center gap-2 group md:cursor-default"
        @click="expanded = !expanded"
        aria-label="Expand player"
      >
        <div class="min-w-0 flex-1">
          <p
            class="hidden md:block text-[10px] font-mono uppercase tracking-widest text-secondary mb-1"
          >
            Now Playing
          </p>
          <p class="truncate text-sm font-black uppercase tracking-tight text-on-surface">
            {{ title }}
          </p>
          <p
            class="hidden md:block truncate text-[10px] font-mono uppercase tracking-widest text-on-surface-variant"
          >
            {{ subLabel }}
          </p>
        </div>
        <span
          class="md:hidden material-symbols-outlined text-base text-secondary group-hover:text-on-surface transition-colors"
        >
          expand_less
        </span>
      </button>

      <div class="flex-1 flex items-center justify-center min-w-0">
        <MediaControls
          :currentTime="currentTime"
          :totalTime="totalTime"
          :progress="progress"
          :volume="volume"
          :duration="duration"
          :canSkip="canSkip"
          :isPlaying="isPlaying"
          :expanded="false"
          @play="emit('play')"
          @pause="emit('pause')"
          @seek="(time) => emit('seek', time)"
          @next="emit('next')"
          @previous="emit('previous')"
          @volumeChange="(v) => emit('volumeChange', v)"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.25s ease;
}
.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
}
</style>
