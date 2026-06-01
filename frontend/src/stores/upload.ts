import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { cancelUpload, listPendingUploads, pauseUpload, uploadFile } from '../api/upload'
import type { PendingUpload } from '../api/upload'

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
  id: number | string
  uploadId?: string
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

const allowedExtensions = [
  '.cbz',
  '.mp3',
  '.flac',
  '.ogg',
  '.m4a',
  '.wav',
  '.aac',
  '.mp4',
  '.mkv',
  '.webm',
  '.mov',
  '.avi',
  '.m4v',
]

function fileExtension(fileName: string): string {
  const dotIndex = fileName.lastIndexOf('.')
  if (dotIndex < 0) return ''
  return fileName.slice(dotIndex).toLowerCase()
}

function isSupportedUpload(file: File): boolean {
  const ext = fileExtension(file.name)
  return allowedExtensions.includes(ext) || file.type.startsWith('video/')
}

export const useUploadStore = defineStore('upload', () => {
  const queuedItems = ref<UploadQueueItem[]>([])
  const transfers = ref<TransferItemData[]>([])
  const loadingTransfers = ref(false)
  const transfersError = ref<string | null>(null)
  const processingUploads = ref(false)
  const rejectedItems = ref<string[]>([])
  const uploadControllers = new Map<string | number, AbortController>()
  let onCompleteCallback: (() => void) | null = null

  const totalBandwidth = computed(() => {
    const bps = transfers.value
      .filter((t) => t.status === 'active')
      .reduce((sum, t) => sum + t.bytesPerSecond, 0)
    return `${((bps * 8) / (1024 * 1024)).toFixed(2)} Mbps`
  })

  const hasActiveTransfers = computed(() => transfers.value.some((t) => t.status === 'active'))

  function setOnComplete(fn: (() => void) | null) {
    onCompleteCallback = fn
  }

  function enqueue(files: File[]) {
    rejectedItems.value = []
    const filtered = files.filter((file) => {
      const supported = isSupportedUpload(file)
      if (!supported) rejectedItems.value.push(file.name)
      return supported
    })
    const nextIdBase = queuedItems.value.length + 1
    const newItems: UploadQueueItem[] = filtered.map((file, index) => ({
      id: nextIdBase + index,
      name: file.name,
      ext: fileExtension(file.name).replace('.', '').toUpperCase() || 'FILE',
      progress: 0,
      status: 'active' as const,
      file,
    }))
    queuedItems.value = [...newItems, ...queuedItems.value]
  }

  function clearQueue() {
    queuedItems.value = []
    rejectedItems.value = []
  }

  async function fetchPendingTransfers() {
    loadingTransfers.value = true
    transfersError.value = null
    try {
      const pendingUploads = await listPendingUploads()
      const runtimeTransfers = transfers.value.filter(isRuntimeTransfer)
      const runtimeUploadIds = new Set(
        runtimeTransfers.map((transfer) => transfer.uploadId).filter(Boolean),
      )
      const runtimeTransferNames = new Set(runtimeTransfers.map((transfer) => transfer.name))
      const pendingTransfers = pendingUploads
        .filter(
          (upload) =>
            !runtimeUploadIds.has(upload.uploadId) && !runtimeTransferNames.has(upload.filename),
        )
        .map(pendingUploadToTransfer)
      transfers.value = [...runtimeTransfers, ...pendingTransfers]
    } catch (error) {
      transfersError.value = error instanceof Error ? error.message : 'Failed to load transfers'
    } finally {
      loadingTransfers.value = false
    }
  }

  function processAll() {
    if (processingUploads.value) return
    processingUploads.value = true
    void runUploadWorker()
  }

  async function runUploadWorker() {
    let completedAny = false

    try {
      while (queuedItems.value.length > 0) {
        const queueItem = queuedItems.value[0]
        if (!queueItem) break

        queuedItems.value = queuedItems.value.slice(1)
        const transfer: TransferItemData = {
          id: `local-${Date.now()}-${Math.random().toString(16).slice(2)}`,
          name: queueItem.name,
          progress: 0,
          sizeInfo: formatFileSize(queueItem.file.size),
          status: 'active',
          eta: '--',
          speed: '--',
          startedAt: performance.now(),
          bytesPerSecond: 0,
        }

        transfers.value = [transfer, ...transfers.value]

        try {
          const controller = new AbortController()
          uploadControllers.set(transfer.id, controller)
          await uploadFile(
            queueItem.file,
            '',
            (progress) => {
              const t = transfers.value.find((t) => t.id === transfer.id)
              if (t) {
                t.progress = progress
                updateTransferStats(t, queueItem.file.size, progress)
              }
            },
            (uploadId) => {
              const t = transfers.value.find((t) => t.id === transfer.id)
              if (t) {
                t.uploadId = uploadId
              }
            },
            controller.signal,
          )
          const t = transfers.value.find((t) => t.id === transfer.id)
          if (t) {
            t.progress = 100
            t.status = 'complete'
            updateTransferStats(t, queueItem.file.size, 100)
          }
          completedAny = true
        } catch {
          const t = transfers.value.find((t) => t.id === transfer.id)
          if (t) {
            t.status = 'paused'
            t.speed = '--'
            t.eta = 'reselect file to continue'
            t.bytesPerSecond = 0
          }
        } finally {
          uploadControllers.delete(transfer.id)
        }
      }
    } finally {
      processingUploads.value = false
      if (completedAny) onCompleteCallback?.()
      await fetchPendingTransfers()
    }
  }

  function isRuntimeTransfer(transfer: TransferItemData): boolean {
    return typeof transfer.id === 'string' && transfer.id.startsWith('local-')
  }

  async function pauseTransfer(transferId: string | number) {
    const transfer = transfers.value.find((transfer) => transfer.id === transferId)
    if (!transfer) return

    transfer.status = 'paused'
    transfer.speed = '--'
    transfer.eta = 'reselect file to continue'
    transfer.bytesPerSecond = 0
    uploadControllers.get(transferId)?.abort()

    if (transfer.uploadId) {
      try {
        await pauseUpload(transfer.uploadId)
      } catch (error) {
        transfersError.value = error instanceof Error ? error.message : 'Failed to pause transfer'
      }
    }
  }

  async function cancelTransfer(transferId: string | number) {
    const transfer = transfers.value.find((transfer) => transfer.id === transferId)
    if (!transfer) return

    uploadControllers.get(transferId)?.abort()
    if (transfer.uploadId) {
      try {
        await cancelUpload(transfer.uploadId)
      } catch (error) {
        transfersError.value = error instanceof Error ? error.message : 'Failed to cancel transfer'
        return
      }
    }

    uploadControllers.delete(transferId)
    transfers.value = transfers.value.filter((transfer) => transfer.id !== transferId)
  }

  function pendingUploadToTransfer(upload: PendingUpload): TransferItemData {
    return {
      id: upload.uploadId,
      uploadId: upload.uploadId,
      name: upload.filename,
      progress:
        upload.totalChunks > 0 ? Math.round((upload.uploadedCount / upload.totalChunks) * 100) : 0,
      sizeInfo: `${formatFileSize(upload.size)} • ${upload.uploadedCount}/${upload.totalChunks} chunks`,
      status: upload.status === 'active' || upload.status === 'assembling' ? 'active' : 'paused',
      eta: upload.status === 'assembling' ? 'assembling' : '--',
      speed: upload.status === 'assembling' ? 'finalizing' : '--',
      startedAt: 0,
      bytesPerSecond: 0,
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
    loadingTransfers,
    transfersError,
    processingUploads,
    rejectedItems,
    totalBandwidth,
    hasActiveTransfers,
    enqueue,
    clearQueue,
    fetchPendingTransfers,
    processAll,
    pauseTransfer,
    cancelTransfer,
    setOnComplete,
  }
})
