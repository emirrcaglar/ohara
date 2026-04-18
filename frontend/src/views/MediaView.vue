<script setup lang="ts">
import { onMounted, watch, ref } from 'vue'
import MediaDisplay from '../components/media/MediaDisplay.vue'
import MediaControls from '../components/media/MediaControls.vue'
import QueueItem from '../components/media/QueueItem.vue'
import { useAudioStore } from '../stores/audio'
import { usePlayerStore } from '../stores/player'
import type { AudioRow } from '../types/api'

const audioStore = useAudioStore()
const playerStore = usePlayerStore()

const audioRef = ref<HTMLAudioElement | null>(null)

onMounted(() => {
  audioStore.fetchLibrary()
})

watch(() => playerStore.currentTrackUrl, (url) => {
  if (audioRef.value && url) {
    audioRef.value.src = url
    audioRef.value.load()
  }
})

watch(() => playerStore.isPlaying, (playing) => {
  if (audioRef.value) {
    if (playing) {
      audioRef.value.play()
    } else {
      audioRef.value.pause()
    }
  }
})

watch(() => playerStore.volume, (vol) => {
  if (audioRef.value) {
    audioRef.value.volume = vol / 100
  }
})

function handleTimeUpdate() {
  if (audioRef.value) {
    playerStore.updateCurrentTime(audioRef.value.currentTime)
  }
}

function handleLoadedMetadata() {
  if (audioRef.value) {
    playerStore.updateDuration(audioRef.value.duration)
  }
}

function handlePlay() {
  playerStore.play()
}

function handlePause() {
  playerStore.pause()
}

function handleEnded() {
  playerStore.next()
}

function selectTrack(track: AudioRow) {
  playerStore.play(track)
  if (audioRef.value) {
    audioRef.value.src = playerStore.currentTrackUrl || ''
    audioRef.value.play()
  }
}
</script>

<template>
  <main class="flex-1 flex flex-col md:flex-row bg-background overflow-hidden">
    <audio
      ref="audioRef"
      @timeupdate="handleTimeUpdate"
      @loadedmetadata="handleLoadedMetadata"
      @play="handlePlay"
      @pause="handlePause"
      @ended="handleEnded"
    />

    <section class="flex-1 flex flex-col p-8 gap-6 min-h-0">
      <MediaDisplay
        :title="playerStore.currentTrack?.title || 'NO_TRACK_SELECTED'"
        :subLabel="playerStore.currentTrack?.album || 'Select_from_queue'"
        imageUrl=""
        bitrate="44.1KHZ / 24BIT"
        :isPlaying="playerStore.isPlaying"
        @play="playerStore.play()"
      />

      <div class="bg-surface-container-low p-6">
        <MediaControls
          :currentTime="playerStore.formattedCurrentTime"
          :totalTime="playerStore.formattedDuration"
          :progress="playerStore.progress"
          :volume="playerStore.volume"
          :isPlaying="playerStore.isPlaying"
          @play="playerStore.play()"
          @pause="playerStore.pause()"
          @next="playerStore.next()"
          @previous="playerStore.previous()"
          @volumeChange="playerStore.setVolume"
        />
      </div>
    </section>

    <aside class="w-full md:w-96 bg-surface-container-low border-l border-white/5 flex flex-col">
      <div class="p-8 pb-4 flex items-center justify-between">
        <h3 class="text-sm font-black uppercase tracking-widest text-secondary">Up_Next</h3>
        <span class="text-[10px] font-mono text-white/40">QUEUE: {{ audioStore.totalItems }}</span>
      </div>

      <div class="flex-1 overflow-y-auto px-4">
        <QueueItem
          v-for="track in audioStore.items"
          :key="track.id"
          :track="track"
          :isActive="playerStore.currentTrack?.id === track.id"
          @select="selectTrack"
        />
      </div>

      <div class="p-8 bg-surface-container-lowest font-mono text-[9px] text-white/20">
        SIGNAL: NOMINAL // AES-256-GCM
      </div>
    </aside>

  </main>
</template>
