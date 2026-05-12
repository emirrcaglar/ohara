<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import VaultHeader from '../components/VaultHeader.vue'
import VaultCard from '../components/VaultCard.vue'
import DropZone from '../components/uploads/DropZone.vue'
import TransferItem from '../components/uploads/TransferItem.vue'
import SystemInput from '../components/uploads/SystemInput.vue'
import SystemParameters from '../components/uploads/SystemParameters.vue'
import activeTransfersIcon from '../assets/active-transfers.svg'
import { useMangaStore } from '../stores/manga'
import { useAudioStore } from '../stores/audio'
import { usePlayerStore } from '../stores/player'
import { getMangaPageUrl } from '../api/manga'
import { uploadFile } from '../api/upload'
import type { MangaRow, AudioRow } from '../types/api'

const router = useRouter()
const mangaStore = useMangaStore()
const audioStore = useAudioStore()
const playerStore = usePlayerStore()

const selectedTab = ref<'ALL' | 'CBZ' | 'AUDIO'>('ALL')
const metadataProfile = ref('AUTO_DETECT_SCRAPER_V2')
const autoExtract = ref(true)
const verifyHash = ref(true)
const overwriteExisting = ref(false)
const metadataFetch = ref(true)
const showUploadDialog = ref(false)
const showTransfersPanel = ref(false)

type UploadStatus = 'active' | 'complete'

interface UploadQueueItem {
  id: number
  name: string
  ext: string
  progress: number
  status: UploadStatus
  file: File
}

interface TransferItemData {
  id: number
  name: string
  progress: number
  sizeInfo?: string
  status: 'active' | 'complete' | 'paused'
  eta?: string
  speed?: string
  storagePath?: string
  startedAt: number
  bytesPerSecond: number
}

const queuedItems = ref<UploadQueueItem[]>([])
const transfers = ref<TransferItemData[]>([])

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

const floatingButtonsBottomClass = computed(() => {
  return playerStore.currentTrack ? 'bottom-28' : 'bottom-6'
})

const totalBandwidth = computed(() => {
  const activeBytesPerSecond = transfers.value
    .filter(transfer => transfer.status === 'active')
    .reduce((sum, transfer) => sum + transfer.bytesPerSecond, 0)

  return `${((activeBytesPerSecond * 8) / (1024 * 1024)).toFixed(2)} Mbps`
})

onMounted(() => {
  mangaStore.fetchLibrary()
  audioStore.fetchLibrary()
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
      total: manga.pageCount
    }
  })
}

function playAudio(audio: AudioRow) {
  playerStore.clearQueue()
  playerStore.play(audio)
}

function handleMangaClick(item: MangaRow | AudioRow) {
  if ('pageCount' in item) {
    openManga(item)
  }
}

function openUploadDialog() {
  showUploadDialog.value = true
}

function closeUploadDialog() {
  showUploadDialog.value = false
}

function openTransfersPanel() {
  showTransfersPanel.value = true
}

function closeTransfersPanel() {
  showTransfersPanel.value = false
}

function handleFilesSelected(files: File[]) {
  const allowedExtensions = ['.cbz', '.mp3', '.wav']
  const filtered = files.filter(file => {
    const ext = '.' + file.name.split('.').pop()?.toLowerCase()
    return allowedExtensions.includes(ext)
  })
  const nextIdBase = queuedItems.value.length + 1
  const newItems: UploadQueueItem[] = filtered.map((file, index) => ({
    id: nextIdBase + index,
    name: file.name,
    ext: file.name.split('.').pop()?.toUpperCase() ?? 'FILE',
    progress: 0,
    status: 'active' as const,
    file
  }))

  queuedItems.value = [...newItems, ...queuedItems.value]
}

function clearQueue() {
  queuedItems.value = []
}

async function processAll() {
  const itemsToUpload = [...queuedItems.value]
  const nextTransferId = transfers.value.length + 1
  const newTransfers: TransferItemData[] = itemsToUpload.map((item, index) => ({
    id: nextTransferId + index,
    name: item.name,
    progress: 0,
    sizeInfo: formatFileSize(item.file.size),
    status: 'active' as const,
    eta: '--',
    speed: '--',
    startedAt: performance.now(),
    bytesPerSecond: 0
  }))

  transfers.value = [...newTransfers, ...transfers.value]
  queuedItems.value = []
  closeUploadDialog()
  openTransfersPanel()

  for (let i = 0; i < itemsToUpload.length; i++) {
    const queueItem = itemsToUpload[i]
    const transfer = newTransfers[i]
    if (!transfer) continue

    try {
      await uploadFile(
        queueItem.file,
        metadataProfile.value,
        (progress) => {
          const t = transfers.value.find(t => t.id === transfer.id)
          if (t) {
            t.progress = progress
            updateTransferStats(t, queueItem.file.size, progress)
          }
        }
      )
      const t = transfers.value.find(t => t.id === transfer.id)
      if (t) {
        t.progress = 100
        t.status = 'complete'
        updateTransferStats(t, queueItem.file.size, 100)
      }
    } catch {
      const t = transfers.value.find(t => t.id === transfer.id)
      if (t) t.status = 'complete'
    }
  }
}

function updateTransferStats(transfer: TransferItemData, fileSize: number, progress: number) {
  if (progress <= 0 || transfer.startedAt <= 0) {
    transfer.speed = '--'
    transfer.eta = '--'
    return
  }

  const elapsedSeconds = (performance.now() - transfer.startedAt) / 1000
  if (elapsedSeconds <= 0) {
    transfer.speed = '--'
    transfer.eta = '--'
    return
  }

  const uploadedBytes = fileSize * (progress / 100)
  const remainingBytes = Math.max(fileSize - uploadedBytes, 0)
  const bytesPerSecond = uploadedBytes / elapsedSeconds

  if (bytesPerSecond <= 0) {
    transfer.speed = '--'
    transfer.eta = '--'
    transfer.bytesPerSecond = 0
    return
  }

  const etaSeconds = Math.ceil(remainingBytes / bytesPerSecond)

  transfer.bytesPerSecond = bytesPerSecond
  transfer.speed = `${(bytesPerSecond / (1024 * 1024)).toFixed(2)} MB/s`
  transfer.eta = formatEta(etaSeconds)
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
}

function formatEta(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) return '--'
  if (seconds === 0) return '00:00'

  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  if (minutes === 0) {
    return `00:${String(remainingSeconds).padStart(2, '0')}`
  }
  return `${String(minutes).padStart(2, '0')}:${String(remainingSeconds).padStart(2, '0')}`
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
        <VaultHeader v-model="selectedTab" :totalManga="mangaStore.total + audioStore.total" />

        <div v-if="mangaStore.loading || audioStore.loading" class="text-secondary">Loading...</div>
        <div v-else-if="mangaStore.error" class="text-error">{{ mangaStore.error }}</div>
        <div v-else-if="audioStore.error" class="text-error">{{ audioStore.error }}</div>

        <div v-else class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3 md:gap-6">
          <VaultCard
            v-for="manga in filteredManga"
            :key="`manga-${manga.id}`"
            :manga="manga"
            :coverUrl="getMangaPageUrl(manga.id, 0)"
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

      <div
        class="fixed right-6 z-40 flex flex-col gap-3"
        :class="floatingButtonsBottomClass"
      >
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

      <section v-if="showUploadDialog" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <button
          class="absolute inset-0 bg-surface-container-lowest/60 backdrop-blur-sm"
          type="button"
          aria-label="Close upload dialog"
          @click="closeUploadDialog"
        ></button>

        <div class="relative w-full max-w-4xl max-h-[90vh] overflow-auto border-t-4 border-primary-container bg-surface-container-low p-6 md:p-8">
          <header class="mb-8 flex items-start justify-between">
            <div>
              <h2 class="text-2xl font-bold uppercase tracking-tighter text-primary-container leading-none">Upload_Media</h2>
              <p class="mt-2 text-xs font-mono uppercase tracking-widest text-on-surface-variant">Awaiting local stream...</p>
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

              <div class="grid grid-cols-1 gap-8 md:grid-cols-2">
                <SystemInput
                  label="Metadata_Profile"
                  :value="metadataProfile"
                icon="expand_more"
              />
            </div>

            <SystemParameters
              :autoExtract="autoExtract"
              :verifyHash="verifyHash"
              :overwriteExisting="overwriteExisting"
            />

            <section class="space-y-4">
              <p class="text-[10px] font-bold uppercase tracking-widest text-secondary">Queued_Operations</p>

              <article
                v-for="item in queuedItems"
                :key="item.id"
                class="flex flex-col gap-2 bg-surface-container-high p-3"
              >
                <div class="flex items-center justify-between">
                  <div class="flex min-w-0 items-center gap-3">
                    <span class="bg-secondary-container px-1 text-[10px] font-bold text-on-secondary-container">{{ item.ext }}</span>
                    <span class="max-w-[200px] truncate text-sm font-medium tracking-tight">{{ item.name }}</span>
                  </div>
                  <span class="text-xs font-mono text-primary-container">{{ item.progress }}%</span>
                </div>

                <div class="h-1 w-full bg-surface">
                  <div class="h-full bg-primary-container" :style="{ width: `${item.progress}%` }"></div>
                </div>
              </article>
            </section>

            <footer class="flex flex-col gap-4 border-t border-surface-container-highest pt-4 md:flex-row md:items-center md:justify-between">
              <label class="group flex cursor-pointer items-center gap-3">
                <div class="relative">
                  <input v-model="metadataFetch" class="peer sr-only" type="checkbox" />
                  <div class="h-5 w-10 bg-surface-container-highest transition-colors peer-checked:bg-secondary-container"></div>
                  <div class="absolute left-1 top-1 h-3 w-3 bg-white transition-transform peer-checked:translate-x-5"></div>
                </div>
                <span class="text-xs font-bold uppercase tracking-tighter text-on-surface-variant transition-colors group-hover:text-secondary">Metadata Fetch</span>
              </label>

              <div class="flex gap-4">
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
              </div>
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

        <aside class="relative h-full w-full max-w-md bg-surface-container-low border-l border-white/10 flex flex-col">
          <div class="p-6 border-b border-white/10 flex items-center justify-between">
            <h3 class="text-xs font-black uppercase tracking-widest text-primary">Active_Transfers</h3>
            <button class="text-on-surface-variant hover:text-on-surface" type="button" @click="closeTransfersPanel">
              <span class="material-symbols-outlined">close</span>
            </button>
          </div>

          <div class="flex-1 overflow-y-auto p-6 space-y-6">
            <p v-if="transfers.length === 0" class="text-xs text-on-surface-variant uppercase tracking-widest">No active transfers</p>
            <TransferItem
              v-for="transfer in transfers"
              :key="transfer.id"
              :name="transfer.name"
              :progress="transfer.progress"
              :sizeInfo="transfer.sizeInfo"
              :status="transfer.status"
              :eta="transfer.eta"
              :speed="transfer.speed"
              :storagePath="transfer.storagePath"
            />
          </div>

          <div class="p-6 bg-surface-container-lowest border-t border-white/10">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <p class="text-[9px] text-secondary uppercase font-bold">Total Bandwidth</p>
                <p class="text-lg font-black text-primary leading-none mt-1">{{ totalBandwidth }}</p>
              </div>
              <div>
                <p class="text-[9px] text-secondary uppercase font-bold">Files in Queue</p>
                <p class="text-lg font-black text-on-surface leading-none mt-1">{{ transfers.length }}</p>
              </div>
            </div>
          </div>
        </aside>
      </section>
    </main>
  </div>
</template>
