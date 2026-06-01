<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import VaultHeader from '../components/VaultHeader.vue'
import VaultCard from '../components/VaultCard.vue'
import DropZone from '../components/uploads/DropZone.vue'
import TransferItem from '../components/uploads/TransferItem.vue'

import activeTransfersIcon from '../assets/active-transfers.svg'
import { useMangaStore } from '../stores/manga'
import { useAudioStore } from '../stores/audio'
import { useVideoStore } from '../stores/video'
import { usePlayerStore } from '../stores/player'
import { useUploadStore } from '../stores/upload'
import { getMangaPageUrl } from '../api/manga'
import type { MangaRow, AudioRow, VideoRow } from '../types/api'

const router = useRouter()
const mangaStore = useMangaStore()
const audioStore = useAudioStore()
const videoStore = useVideoStore()
const playerStore = usePlayerStore()

const uploadStore = useUploadStore()

const selectedTab = ref<'ALL' | 'CBZ' | 'AUDIO' | 'VIDEO'>('ALL')
const showUploadDialog = ref(false)
const showTransfersPanel = ref(false)

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

const filteredVideo = computed(() => {
  if (selectedTab.value === 'ALL' || selectedTab.value === 'VIDEO') {
    return videoStore.items
  }
  return []
})

const totalMedia = computed(() => mangaStore.total + audioStore.total + videoStore.total)

const floatingButtonsBottomClass = computed(() => {
  return playerStore.currentTrack ? 'bottom-28' : 'bottom-6'
})

onMounted(() => {
  mangaStore.fetchLibrary()
  audioStore.fetchLibrary()
  videoStore.fetchLibrary()
  uploadStore.fetchPendingTransfers()
  window.addEventListener('keydown', handleGlobalKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleGlobalKeydown)
})

function openManga(manga: MangaRow) {
  router.push({
    path: '/reader',
    query: {
      manga: manga.id,
      page: manga.currentPage || 0,
      total: manga.pageCount,
    },
  })
}

function playAudio(audio: AudioRow) {
  playerStore.setQueue(audioStore.items)
  playerStore.play(audio)
}

function openVideo(video: VideoRow) {
  router.push({
    path: `/video/${video.id}`,
  })
}

function formatDuration(seconds: number) {
  if (!seconds) return ''

  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const remainingSeconds = seconds % 60

  if (hours) {
    return `${hours}:${String(minutes).padStart(2, '0')}:${String(remainingSeconds).padStart(2, '0')}`
  }
  return `${minutes}:${String(remainingSeconds).padStart(2, '0')}`
}

function videoStats(video: VideoRow) {
  if (video.completed) return 'WATCHED'

  const parts: string[] = []
  if (video.duration) parts.push(formatDuration(video.duration))
  if (video.width && video.height) parts.push(`${video.height}P`)
  if (video.position && video.duration) {
    parts.push(`${Math.round((video.position / video.duration) * 100)}%`)
  }

  return parts.length ? parts.join(' · ') : 'READY'
}

function handleMangaClick(item: MangaRow | AudioRow | VideoRow) {
  if ('pageCount' in item) {
    openManga(item)
  }
}

async function deleteManga(item: MangaRow | VideoRow) {
  if (!('pageCount' in item)) return

  const confirmed = window.confirm(
    `Delete "${item.title}" from the library? This removes its index, reading progress, and cached pages.`,
  )
  if (!confirmed) return

  await mangaStore.removeManga(item.id)
}

async function deleteVideo(item: MangaRow | VideoRow) {
  if ('pageCount' in item) return

  const confirmed = window.confirm(
    `Delete "${item.title}" from the library? This removes its index but keeps the video file on disk.`,
  )
  if (!confirmed) return

  await videoStore.removeVideo(item.id)
}

function openUploadDialog() {
  showUploadDialog.value = true
}

function closeUploadDialog() {
  showUploadDialog.value = false
}

function openTransfersPanel() {
  showTransfersPanel.value = true
  uploadStore.fetchPendingTransfers()
}

function closeTransfersPanel() {
  showTransfersPanel.value = false
}

function handleFilesSelected(files: File[]) {
  uploadStore.enqueue(files)
}

function clearQueue() {
  uploadStore.clearQueue()
}

function processAll() {
  closeUploadDialog()
  uploadStore.setOnComplete(() => {
    mangaStore.fetchLibrary()
    audioStore.fetchLibrary()
    videoStore.fetchLibrary()
  })
  uploadStore.processAll()
  openTransfersPanel()
}

function handleGlobalKeydown(event: KeyboardEvent) {
  if (event.key !== 'Escape') return

  if (showUploadDialog.value) {
    closeUploadDialog()
  }

  if (showTransfersPanel.value) {
    closeTransfersPanel()
  }
}
</script>

<template>
  <div class="h-full flex flex-col">
    <main class="flex-1 overflow-y-auto">
      <section class="p-4 md:p-8 flex-1 bg-surface">
        <VaultHeader v-model="selectedTab" :totalManga="totalMedia" />

        <div
          v-if="mangaStore.loading || audioStore.loading || videoStore.loading"
          class="text-secondary"
        >
          Loading...
        </div>
        <div v-else-if="mangaStore.error" class="text-error">{{ mangaStore.error }}</div>
        <div v-else-if="audioStore.error" class="text-error">{{ audioStore.error }}</div>
        <div v-else-if="videoStore.error" class="text-error">{{ videoStore.error }}</div>

        <div v-else class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3 md:gap-6">
          <VaultCard
            v-for="manga in filteredManga"
            :key="`manga-${manga.id}`"
            :manga="manga"
            :coverUrl="getMangaPageUrl(manga.id, 0)"
            category="MANGA_ARCHIVE"
            :stats="`${manga.currentPage} / ${manga.pageCount} PAGES`"
            @click="handleMangaClick"
            @delete="deleteManga"
          />

          <VaultCard
            v-for="audio in filteredAudio"
            :key="`audio-${audio.id}`"
            :audio="audio"
            category="AUDIO_ARCHIVE"
            :stats="`${Math.floor(audio.duration / 60)}:${String(audio.duration % 60).padStart(2, '0')} MIN`"
            @click="() => playAudio(audio)"
          />

          <VaultCard
            v-for="video in filteredVideo"
            :key="`video-${video.id}`"
            :video="video"
            category="VIDEO_ARCHIVE"
            :stats="videoStats(video)"
            @click="() => openVideo(video)"
            @delete="deleteVideo"
          />
        </div>
      </section>

      <div class="fixed right-6 z-40 flex flex-col gap-3" :class="floatingButtonsBottomClass">
        <button
          class="relative h-14 w-14 rounded-full bg-surface-container-high border border-white/10 flex items-center justify-center overflow-hidden hover:border-primary-container hover:bg-surface-container-highest transition-colors shadow-[0_8px_24px_rgba(0,0,0,0.45)]"
          type="button"
          aria-label="Open active transfers panel"
          @click="openTransfersPanel"
        >
          <img
            :src="activeTransfersIcon"
            alt="Active transfers"
            class="h-7 w-7 object-contain [filter:invert(92%)_sepia(6%)_saturate(194%)_hue-rotate(320deg)_brightness(98%)_contrast(90%)]"
          />
        </button>

        <button
          class="h-14 w-14 rounded-full bg-primary-container text-on-primary-container text-3xl font-black leading-none hover:bg-primary transition-colors shadow-[0_8px_24px_rgba(0,0,0,0.45)]"
          type="button"
          aria-label="Open upload modal"
          @click="openUploadDialog"
        >
          +
        </button>
      </div>

      <section
        v-if="showUploadDialog"
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
      >
        <button
          class="absolute inset-0 bg-surface-container-lowest/60 backdrop-blur-sm"
          type="button"
          aria-label="Close upload dialog"
          @click="closeUploadDialog"
        ></button>

        <div
          class="relative w-full max-w-4xl max-h-[90vh] overflow-auto border-t-4 border-primary-container bg-surface-container-low p-6 md:p-8"
        >
          <header class="mb-8 flex items-start justify-between">
            <div>
              <h2
                class="text-2xl font-bold uppercase tracking-tighter text-primary-container leading-none"
              >
                Upload_Media
              </h2>
              <p class="mt-2 text-xs font-mono uppercase tracking-widest text-on-surface-variant">
                Awaiting local stream...
              </p>
            </div>
            <button
              class="text-on-surface-variant transition-colors hover:text-on-surface"
              aria-label="Close upload dialog"
              type="button"
              @click="closeUploadDialog"
            >
              <span class="material-symbols-outlined">close</span>
            </button>
          </header>

          <div class="space-y-8">
            <DropZone @filesSelected="handleFilesSelected" />

            <section class="space-y-4">
              <p class="text-[10px] font-bold uppercase tracking-widest text-secondary">
                Queued_Operations
              </p>

              <p
                v-if="uploadStore.rejectedItems.length > 0"
                class="text-xs text-error uppercase tracking-widest"
              >
                Unsupported file skipped: {{ uploadStore.rejectedItems.join(', ') }}
              </p>

              <article
                v-for="item in uploadStore.queuedItems"
                :key="item.id"
                class="flex flex-col gap-2 bg-surface-container-high p-3"
              >
                <div class="flex items-center justify-between">
                  <div class="flex min-w-0 items-center gap-3">
                    <span
                      class="bg-secondary-container px-1 text-[10px] font-bold text-on-secondary-container"
                      >{{ item.ext }}</span
                    >
                    <span class="max-w-[200px] truncate text-sm font-medium tracking-tight">{{
                      item.name
                    }}</span>
                  </div>
                  <span class="text-xs font-mono text-primary-container">{{ item.progress }}%</span>
                </div>

                <div class="h-1 w-full bg-surface">
                  <div
                    class="h-full bg-primary-container"
                    :style="{ width: `${item.progress}%` }"
                  ></div>
                </div>
              </article>
            </section>

            <footer class="flex justify-end gap-4 pt-4 border-t border-surface-container-highest">
              <button
                class="px-4 py-2 text-xs font-bold uppercase text-on-surface-variant transition-colors hover:text-on-surface"
                type="button"
                @click="clearQueue"
              >
                Clear Queue
              </button>
              <button
                class="bg-primary-container px-6 py-2 text-xs font-bold uppercase text-on-primary-container transition-colors hover:bg-primary"
                type="button"
                @click="processAll"
              >
                Commit Upload
              </button>
            </footer>
          </div>
        </div>
      </section>

      <section v-if="showTransfersPanel" class="fixed inset-0 z-50 flex justify-end">
        <button
          class="absolute inset-0 bg-surface-container-lowest/50 backdrop-blur-sm"
          type="button"
          aria-label="Close transfers panel"
          @click="closeTransfersPanel"
        ></button>

        <aside
          class="relative h-full w-full max-w-md bg-surface-container-low border-l border-white/10 flex flex-col"
        >
          <div class="p-6 border-b border-white/10 flex items-center justify-between">
            <h3 class="text-xs font-black uppercase tracking-widest text-primary">
              Active_Transfers
            </h3>
            <button
              class="text-on-surface-variant hover:text-on-surface"
              type="button"
              @click="closeTransfersPanel"
            >
              <span class="material-symbols-outlined">close</span>
            </button>
          </div>

          <div class="flex-1 overflow-y-auto p-6 space-y-6">
            <p
              v-if="uploadStore.loadingTransfers"
              class="text-xs text-on-surface-variant uppercase tracking-widest"
            >
              Loading transfers...
            </p>
            <p
              v-else-if="uploadStore.transfersError"
              class="text-xs text-error uppercase tracking-widest"
            >
              {{ uploadStore.transfersError }}
            </p>
            <p
              v-else-if="uploadStore.visibleTransfers.length === 0"
              class="text-xs text-on-surface-variant uppercase tracking-widest"
            >
              No active transfers
            </p>
            <TransferItem
              v-for="transfer in uploadStore.visibleTransfers"
              :key="transfer.id"
              :name="transfer.name"
              :progress="transfer.progress"
              :sizeInfo="transfer.sizeInfo"
              :status="transfer.status"
              :eta="transfer.eta"
              :speed="transfer.speed"
              :storagePath="transfer.storagePath"
              :canMoveUp="uploadStore.canMoveTransferUp(transfer.id)"
              :canMoveDown="uploadStore.canMoveTransferDown(transfer.id)"
              @cancel="uploadStore.cancelTransfer(transfer.id)"
              @moveUp="uploadStore.moveTransferUp(transfer.id)"
              @moveDown="uploadStore.moveTransferDown(transfer.id)"
            />
          </div>

          <div class="p-6 bg-surface-container-lowest border-t border-white/10">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <p class="text-[9px] text-secondary uppercase font-bold">Total Bandwidth</p>
                <p class="text-lg font-black text-primary leading-none mt-1">
                  {{ uploadStore.totalBandwidth }}
                </p>
              </div>
              <div>
                <p class="text-[9px] text-secondary uppercase font-bold">Files in Queue</p>
                <p class="text-lg font-black text-on-surface leading-none mt-1">
                  {{ uploadStore.visibleTransfers.length }}
                </p>
              </div>
            </div>
          </div>
        </aside>
      </section>
    </main>
  </div>
</template>
