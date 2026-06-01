<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getVideoStreamUrl } from '../api/video'
import { useVideoStore } from '../stores/video'
import type { VideoInfo, VideoStateUpdate } from '../types/api'

const route = useRoute()
const router = useRouter()
const videoStore = useVideoStore()

const video = ref<VideoInfo | null>(null)
const videoEl = ref<HTMLVideoElement | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
const playbackError = ref<string | null>(null)
const resumed = ref(false)
const lastSavedAt = ref(0)

const videoId = computed(() => Number(route.params.id))
const streamUrl = computed(() =>
  Number.isFinite(videoId.value) ? `${getVideoStreamUrl(videoId.value)}?t=${Date.now()}` : '',
)

const progressPercent = computed(() => {
  if (!video.value?.duration) return 0
  return Math.min(100, Math.round((video.value.position / video.value.duration) * 100))
})

onMounted(async () => {
  if (!Number.isFinite(videoId.value)) {
    error.value = 'Invalid video ID'
    return
  }

  loading.value = true
  try {
    video.value = await videoStore.getVideoInfo(videoId.value)
    if (!video.value) {
      error.value = 'Video not found'
    }
  } finally {
    loading.value = false
  }
})

onBeforeUnmount(() => {
  saveCurrentState()
})

function currentState(lastError = ''): VideoStateUpdate | null {
  const media = videoEl.value
  const current = video.value
  if (!media || !current) return null

  const duration = Math.floor(
    Number.isFinite(media.duration) ? media.duration : current.duration || 0,
  )
  const position = Math.floor(
    Number.isFinite(media.currentTime) ? media.currentTime : current.position || 0,
  )
  const completed =
    duration > 0 && (media.ended || position >= Math.max(duration - 30, duration * 0.95))

  return {
    duration,
    width: media.videoWidth || current.width || 0,
    height: media.videoHeight || current.height || 0,
    position: completed ? 0 : position,
    completed,
    lastError,
  }
}

async function saveState(state: VideoStateUpdate | null) {
  if (!state || !video.value) return

  video.value = {
    ...video.value,
    duration: state.duration || video.value.duration,
    width: state.width || video.value.width,
    height: state.height || video.value.height,
    position: state.position,
    completed: state.completed,
    lastError: state.lastError,
  }

  try {
    await videoStore.updateVideoState(video.value.id, state)
  } catch {
    // Playback should not be interrupted if progress persistence fails.
  }
}

function saveCurrentState(lastError = '') {
  void saveState(currentState(lastError))
}

function handleLoadedMetadata() {
  const media = videoEl.value
  const savedPosition = video.value?.position || 0

  if (
    !resumed.value &&
    savedPosition > 5 &&
    media?.duration &&
    savedPosition < media.duration - 10
  ) {
    media.currentTime = savedPosition
    resumed.value = true
  }

  saveCurrentState()
}

function handleTimeUpdate() {
  const now = Date.now()
  if (now - lastSavedAt.value < 15000) return

  lastSavedAt.value = now
  saveCurrentState()
}

function handleEnded() {
  saveCurrentState()
}

function handleVideoError(event: Event) {
  const media = event.target as HTMLVideoElement
  const code = media.error?.code
  const message = media.error?.message
  const extension = video.value?.fileExtension?.toLowerCase()

  if (extension === '.mkv') {
    playbackError.value =
      'This MKV file is indexed and streaming, but your browser may not support MKV playback natively. Download the original file or try a browser-playable MP4/WebM file.'
  } else {
    playbackError.value = `The browser could not play this video stream${code ? ` (media error ${code})` : ''}${message ? `: ${message}` : '.'}`
  }

  saveCurrentState(playbackError.value)
}

function formatDuration(seconds: number) {
  if (!seconds) return 'Unknown duration'

  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const remainingSeconds = seconds % 60

  if (hours) {
    return `${hours}:${String(minutes).padStart(2, '0')}:${String(remainingSeconds).padStart(2, '0')}`
  }
  return `${minutes}:${String(remainingSeconds).padStart(2, '0')}`
}

function backToLibrary() {
  router.push('/library')
}
</script>

<template>
  <main class="min-h-full bg-surface p-4 md:p-8">
    <header class="mb-6 flex items-center justify-between gap-4">
      <div>
        <p class="text-[10px] font-bold uppercase tracking-widest text-secondary">Video_Stream</p>
        <h1 class="mt-1 text-2xl font-black uppercase tracking-tight text-on-surface">
          {{ video?.title || 'Loading video...' }}
        </h1>
        <p
          v-if="video"
          class="mt-2 text-xs font-bold uppercase tracking-wider text-on-surface-variant"
        >
          {{ formatDuration(video.duration) }}
          <span v-if="video.width && video.height"> · {{ video.width }}×{{ video.height }}</span>
          <span v-if="video.completed"> · Watched</span>
          <span v-else-if="progressPercent"> · {{ progressPercent }}% watched</span>
        </p>
      </div>

      <button
        class="bg-surface-container-high px-4 py-2 text-xs font-bold uppercase text-on-surface-variant transition-colors hover:bg-surface-container-highest hover:text-on-surface"
        type="button"
        @click="backToLibrary"
      >
        Back to library
      </button>
    </header>

    <p v-if="loading" class="text-secondary">Loading...</p>
    <p v-else-if="error" class="text-error">{{ error }}</p>

    <section v-else class="mx-auto max-w-6xl overflow-hidden bg-surface-container-low shadow-2xl">
      <video
        v-if="streamUrl"
        ref="videoEl"
        class="aspect-video w-full bg-black"
        :src="streamUrl"
        controls
        autoplay
        playsinline
        preload="metadata"
        @loadedmetadata="handleLoadedMetadata"
        @timeupdate="handleTimeUpdate"
        @pause="saveCurrentState()"
        @ended="handleEnded"
        @error="handleVideoError"
      ></video>

      <div v-if="progressPercent" class="h-1 bg-surface-container-highest">
        <div class="h-full bg-primary" :style="{ width: `${progressPercent}%` }"></div>
      </div>

      <p v-if="playbackError" class="p-4 text-sm text-error">
        {{ playbackError }}
      </p>
    </section>
  </main>
</template>
