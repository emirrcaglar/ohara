<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getVideoStreamUrl } from '../api/video'
import { useVideoStore } from '../stores/video'
import type { VideoInfo } from '../types/api'

const route = useRoute()
const router = useRouter()
const videoStore = useVideoStore()

const video = ref<VideoInfo | null>(null)
const loading = ref(false)
const error = ref<string | null>(null)
const playbackError = ref<string | null>(null)

const videoId = computed(() => Number(route.params.id))
const streamUrl = computed(() =>
  Number.isFinite(videoId.value) ? `${getVideoStreamUrl(videoId.value)}?t=${Date.now()}` : '',
)

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

function handleVideoError(event: Event) {
  const media = event.target as HTMLVideoElement
  const code = media.error?.code
  const message = media.error?.message
  const extension = video.value?.fileExtension?.toLowerCase()

  if (extension === '.mkv') {
    playbackError.value =
      'This MKV file is indexed and streaming, but your browser may not support MKV playback natively. Try an MP4/WebM file, or add a future remux/transcode step.'
    return
  }

  playbackError.value = `The browser could not play this video stream${code ? ` (media error ${code})` : ''}${message ? `: ${message}` : '.'}`
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
        class="aspect-video w-full bg-black"
        :src="streamUrl"
        controls
        autoplay
        playsinline
        preload="metadata"
        @error="handleVideoError"
      ></video>

      <p v-if="playbackError" class="p-4 text-sm text-error">
        {{ playbackError }}
      </p>
    </section>
  </main>
</template>
