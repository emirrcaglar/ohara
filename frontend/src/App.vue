<script setup lang="ts">
import { ref, watch } from 'vue';
import Sidebar from './components/Sidebar.vue';
import TopBar from './components/TopBar.vue';
import TelemetryCard from './components/TelemetryCard.vue';
import OperationsCard from './components/OperationsCard.vue';
import MediaCard from './components/MediaCard.vue';
import LogsTable from './components/LogsTable.vue';
import MediaBar from './components/media/MediaBar.vue';
import StatusBar from './components/StatusBar.vue';
import { usePlayerStore } from './stores/player';

const playerStore = usePlayerStore();
const audioRef = ref<HTMLAudioElement | null>(null);

watch(() => playerStore.currentTrackUrl, (url) => {
  if (audioRef.value && url) {
    audioRef.value.src = url;
    audioRef.value.load();
  }
});

watch(() => playerStore.isPlaying, (playing) => {
  if (audioRef.value && playerStore.currentTrackUrl) {
    if (playing) {
      audioRef.value.play();
    } else {
      audioRef.value.pause();
    }
  }
});

watch(() => playerStore.volume, (vol) => {
  if (audioRef.value) {
    audioRef.value.volume = vol / 100;
  }
});

function handleTimeUpdate() {
  if (audioRef.value) {
    playerStore.updateCurrentTime(audioRef.value.currentTime);
  }
}

function handleLoadedMetadata() {
  if (audioRef.value) {
    playerStore.updateDuration(audioRef.value.duration);
  }
}

function handlePlay() {
  playerStore.play();
}

function handlePause() {
  playerStore.pause();
}

function handleEnded() {
  playerStore.next();
}
</script>

<template>
  <div class="min-h-screen bg-background text-on-surface font-sans selection:bg-primary-container selection:text-on-primary-container">
    <audio
      ref="audioRef"
      @timeupdate="handleTimeUpdate"
      @loadedmetadata="handleLoadedMetadata"
      @play="handlePlay"
      @pause="handlePause"
      @ended="handleEnded"
    />

    <div class="digital-grain"></div>

    <Sidebar />

    <main class="ml-64 flex flex-col h-screen">
      <TopBar />

      <RouterView />

      <MediaBar
        v-if="playerStore.currentTrack"
        :title="playerStore.currentTrack.title"
        :subLabel="playerStore.currentTrack.album || 'Unknown Album'"
        bitrate="44.1KHZ / 24BIT"
        :isPlaying="playerStore.isPlaying"
        @play="playerStore.play()"
        @pause="playerStore.pause()"
        @next="playerStore.next()"
        @previous="playerStore.previous()"
        @volumeChange="playerStore.setVolume"
      />

      <StatusBar />
    </main>
  </div>
</template>

<style>
/* Custom scrollbar for the dark theme */
::-webkit-scrollbar {
  width: 4px;
}
::-webkit-scrollbar-track {
  background: surface-dim;
}
::-webkit-scrollbar-thumb {
  background: surface-variant;
}
::-webkit-scrollbar-thumb:hover {
  background: #primary-container;
}
</style>

