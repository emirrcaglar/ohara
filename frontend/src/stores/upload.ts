import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { uploadFile } from '../api/upload'

export type UploadStatus = 'active' | 'complete'

export interface UploadQueueItem {
  id: number
  name: string
  ext: string
  progress: number
  status: UploadStatus
  file: File
}

export interface TransferItemData {
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

export const useUploadStore = defineStore('upload', () => {
  const queuedItems = ref<UploadQueueItem[]>([])
  const transfers = ref<TransferItemData[]>([])
  const metadataProfile = ref('AUTO_DETECT_SCRAPER_V2')
  const autoExtract = ref(true)
  const verifyHash = ref(true)
  const overwriteExisting = ref(false)
  const metadataFetch = ref(true)

  const totalBandwidth = computed(() => {
    const bps = transfers.value
      .filter(t => t.status === 'active')
      .reduce((sum, t) => sum + t.bytesPerSecond, 0)
    return `${((bps * 8) / (1024 * 1024)).toFixed(2)} Mbps`
  })

  const hasActiveTransfers = computed(() =>
    transfers.value.some(t => t.status === 'active')
  )

  function enqueue(files: File[]) {
    const allowedExtensions = ['.cbz', '.mp3', '.flac', '.ogg', '.m4a', '.wav', '.aac']
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
      file,
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
      bytesPerSecond: 0,
    }))

    transfers.value = [...newTransfers, ...transfers.value]
    queuedItems.value = []

    for (let i = 0; i < itemsToUpload.length; i++) {
      const queueItem = itemsToUpload[i]
      const transfer = newTransfers[i]
      if (!transfer) continue

      try {
        await uploadFile(queueItem.file, metadataProfile.value, (progress) => {
          const t = transfers.value.find(t => t.id === transfer.id)
          if (t) {
            t.progress = progress
            updateTransferStats(t, queueItem.file.size, progress)
          }
        })
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
    transfer.bytesPerSecond = bytesPerSecond
    transfer.speed = `${(bytesPerSecond / (1024 * 1024)).toFixed(2)} MB/s`
    transfer.eta = formatEta(Math.ceil(remainingBytes / bytesPerSecond))
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
    if (minutes === 0) return `00:${String(remainingSeconds).padStart(2, '0')}`
    return `${String(minutes).padStart(2, '0')}:${String(remainingSeconds).padStart(2, '0')}`
  }

  return {
    queuedItems,
    transfers,
    metadataProfile,
    autoExtract,
    verifyHash,
    overwriteExisting,
    metadataFetch,
    totalBandwidth,
    hasActiveTransfers,
    enqueue,
    clearQueue,
    processAll,
  }
})
