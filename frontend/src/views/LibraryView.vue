<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import VaultHeader from '../components/VaultHeader.vue'
import VaultCard from '../components/VaultCard.vue'
import MediaBar from '../components/media/MediaBar.vue'
import { useMangaStore } from '../stores/manga'
import { useAudioStore } from '../stores/audio'
import { usePlayerStore } from '../stores/player'
import { getMangaCoverUrl } from '../api/manga'
import type { MangaRow, AudioRow } from '../types/api'

const router = useRouter()
const mangaStore = useMangaStore()
const audioStore = useAudioStore()
const playerStore = usePlayerStore()

const audioRef = ref<HTMLAudioElement | null>(null)
const selectedTab = ref<'ALL' | 'CBZ' | 'AUDIO'>('ALL')

const filteredManga = computed(() => {
  if (selectedTab.value === 'ALL' || selectedTab.value === 'CBZ') {
    return mangaStore.items
  }
  return []
})

const filteredAudio = computed(() => {
  if (selectedTab.value === 'ALL' || selectedTab.value === 'AUDIO') {
    return audioStore.items
  }
  return []
})

onMounted(() => {
  mangaStore.fetchLibrary()
  audioStore.fetchLibrary()
})

watch(() => playerStore.currentTrackUrl, (url) => {
  if (audioRef.value && url) {
    audioRef.value.src = url
    audioRef.value.load()
  }
})

watch(() => playerStore.isPlaying, (playing) => {
  if (audioRef.value && playerStore.currentTrackUrl) {
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

function openManga(manga: MangaRow) {
  playerStore.pause()
  playerStore.clearQueue()
  router.push({
    path: '/reader',
    query: {
      manga: manga.id,
      page: manga.currentPage || 0,
      total: manga.pageCount
    }
  })
}

function playAudio(audio: AudioRow) {
  playerStore.clearQueue()
  playerStore.play(audio)
  const url = playerStore.currentTrackUrl
  if (audioRef.value && url) {
    audioRef.value.src = url
    audioRef.value.load()
    audioRef.value.play()
  }
}

function handleMangaClick(item: MangaRow | AudioRow) {
  if ('pageCount' in item) {
    openManga(item)
  }
}
</script>

<template>
  <div class="h-full flex flex-col">
    <audio
      ref="audioRef"
      @timeupdate="handleTimeUpdate"
      @loadedmetadata="handleLoadedMetadata"
      @play="handlePlay"
      @pause="handlePause"
      @ended="handleEnded"
    />

    <main class="flex-1 overflow-y-auto">
      <section class="p-8 flex-1 bg-surface">
        <VaultHeader v-model="selectedTab" :totalManga="mangaStore.total + audioStore.total" />

        <div v-if="mangaStore.loading || audioStore.loading" class="text-secondary">Loading...</div>
        <div v-else-if="mangaStore.error" class="text-error">{{ mangaStore.error }}</div>
        <div v-else-if="audioStore.error" class="text-error">{{ audioStore.error }}</div>

        <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-6">
          <VaultCard
            v-for="manga in filteredManga"
            :key="`manga-${manga.id}`"
            :manga="manga"
            :coverUrl="getMangaCoverUrl(manga.id)"
            category="MANGA_ARCHIVE"
            :stats="`${manga.currentPage} / ${manga.pageCount} PAGES`"
            @click="handleMangaClick"
          />

          <VaultCard
            v-for="audio in filteredAudio"
            :key="`audio-${audio.id}`"
            :audio="audio"
            category="AUDIO_ARCHIVE"
            :stats="`${Math.floor(audio.duration / 60)}:${String(audio.duration % 60).padStart(2, '0')} MIN`"
            @click="() => playAudio(audio)"
          />
        </div>
      </section>
    </main>

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
  </div>
</template>
