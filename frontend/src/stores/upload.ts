import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { cancelUpload, listPendingUploads, pauseUpload, uploadFile } from '../api/upload'
import type { PendingUpload } from '../api/upload'

export type UploadStatus = 'queued' | 'active' | 'complete'

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
  status: 'queued' | 'active' | 'complete' | 'paused' | 'failed'
  eta?: string
  speed?: string
  storagePath?: string
  startedAt: number
  bytesPerSecond: number
  file?: File
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
  let nextQueueId = 1
  let resumeTimer: number | undefined

  const totalBandwidth = computed(() => {
    const bps = transfers.value
      .filter((t) => t.status === 'active')
      .reduce((sum, t) => sum + t.bytesPerSecond, 0)
    return `${((bps * 8) / (1024 * 1024)).toFixed(2)} Mbps`
  })

  const visibleTransfers = computed(() => {
    const activeTransfers = transfers.value.filter((transfer) => transfer.status === 'active')
    const waitingTransfers = queuedItems.value.map(queueItemToTransfer)
    const inactiveTransfers = transfers.value.filter((transfer) => transfer.status !== 'active')
    return [...activeTransfers, ...waitingTransfers, ...inactiveTransfers]
  })

  const hasActiveTransfers = computed(() => visibleTransfers.value.length > 0)

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
    const newItems: UploadQueueItem[] = filtered.map((file) => ({
      id: nextQueueId++,
      name: file.name,
      ext: fileExtension(file.name).replace('.', '').toUpperCase() || 'FILE',
      progress: 0,
      status: 'queued' as const,
      file,
    }))
    queuedItems.value = [...queuedItems.value, ...newItems].sort(compareQueueItemNames)
  }

  function clearQueue() {
    queuedItems.value = []
    rejectedItems.value = []
  }

  async function fetchPendingTransfers() {
    if (processingUploads.value) return

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
          file: queueItem.file,
        }

        transfers.value = [...transfers.value, transfer].sort(compareTransferNames)

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
        } catch (error) {
          const t = transfers.value.find((t) => t.id === transfer.id)
          if (t) {
            if (isAbortError(error) && t.status === 'paused') {
              t.speed = '--'
              t.eta = 'reselect file to continue'
            } else if (isSQLiteBusyError(error)) {
              transfers.value = transfers.value.filter((transfer) => transfer.id !== t.id)
              queuedItems.value = [queueItem, ...queuedItems.value]
              scheduleUploadResume()
              break
            } else {
              t.status = 'failed'
              t.speed = '--'
              t.eta = uploadErrorMessage(error)
            }
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

  function scheduleUploadResume() {
    if (resumeTimer !== undefined) return

    resumeTimer = window.setTimeout(() => {
      resumeTimer = undefined
      processAll()
    }, 1500)
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

  async function moveTransferUp(transferId: string | number) {
    if (typeof transferId !== 'string' || !transferId.startsWith('queue-')) return

    const queueId = Number(transferId.slice('queue-'.length))
    const index = queuedItems.value.findIndex((item) => item.id === queueId)
    if (index < 0) return

    if (index > 0) {
      const nextQueue = [...queuedItems.value]
      ;[nextQueue[index - 1], nextQueue[index]] = [nextQueue[index], nextQueue[index - 1]]
      queuedItems.value = nextQueue
      return
    }

    const activeTransfer = transfers.value.find((transfer) => transfer.status === 'active')
    if (activeTransfer) {
      preemptActiveTransfer(activeTransfer, 1)
    }
  }

  async function moveTransferDown(transferId: string | number) {
    const activeTransfer = transfers.value.find((transfer) => transfer.id === transferId)
    if (activeTransfer?.status === 'active' && queuedItems.value.length > 0) {
      preemptActiveTransfer(activeTransfer, 1)
      return
    }

    if (typeof transferId !== 'string' || !transferId.startsWith('queue-')) return

    const queueId = Number(transferId.slice('queue-'.length))
    const index = queuedItems.value.findIndex((item) => item.id === queueId)
    if (index < 0 || index >= queuedItems.value.length - 1) return

    const nextQueue = [...queuedItems.value]
    ;[nextQueue[index], nextQueue[index + 1]] = [nextQueue[index + 1], nextQueue[index]]
    queuedItems.value = nextQueue
  }

  function preemptActiveTransfer(transfer: TransferItemData, queueIndex: number) {
    if (!transfer.file) return

    const nextQueue = [...queuedItems.value]
    nextQueue.splice(queueIndex, 0, transferToQueueItem(transfer))
    queuedItems.value = nextQueue
    transfers.value = transfers.value.filter((item) => item.id !== transfer.id)
    uploadControllers.get(transfer.id)?.abort()
    uploadControllers.delete(transfer.id)
  }

  function transferToQueueItem(transfer: TransferItemData): UploadQueueItem {
    return {
      id: nextQueueId++,
      name: transfer.name,
      ext: fileExtension(transfer.name).replace('.', '').toUpperCase() || 'FILE',
      progress: transfer.progress,
      status: 'queued',
      file: transfer.file as File,
    }
  }

  function canMoveTransferUp(transferId: string | number): boolean {
    if (typeof transferId !== 'string' || !transferId.startsWith('queue-')) return false

    const queueId = Number(transferId.slice('queue-'.length))
    const index = queuedItems.value.findIndex((item) => item.id === queueId)
    return (
      index > 0 ||
      (index === 0 &&
        transfers.value.some((transfer) => transfer.status === 'active' && transfer.file))
    )
  }

  function canMoveTransferDown(transferId: string | number): boolean {
    const activeTransfer = transfers.value.find((transfer) => transfer.id === transferId)
    if (activeTransfer?.status === 'active')
      return queuedItems.value.length > 0 && Boolean(activeTransfer.file)

    if (typeof transferId !== 'string' || !transferId.startsWith('queue-')) return false

    const queueId = Number(transferId.slice('queue-'.length))
    const index = queuedItems.value.findIndex((item) => item.id === queueId)
    return index >= 0 && index < queuedItems.value.length - 1
  }

  async function cancelTransfer(transferId: string | number) {
    if (typeof transferId === 'string' && transferId.startsWith('queue-')) {
      const queueId = Number(transferId.slice('queue-'.length))
      queuedItems.value = queuedItems.value.filter((item) => item.id !== queueId)
      return
    }

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

  function retryTransfer(transferId: string | number) {
    const transfer = transfers.value.find((transfer) => transfer.id === transferId)
    if (!transfer?.file || (transfer.status !== 'failed' && transfer.status !== 'paused')) return

    transfers.value = transfers.value.filter((transfer) => transfer.id !== transferId)
    enqueue([transfer.file])
    processAll()
  }

  function queueItemToTransfer(item: UploadQueueItem): TransferItemData {
    return {
      id: `queue-${item.id}`,
      name: item.name,
      progress: item.progress,
      sizeInfo: `${formatFileSize(item.file.size)} • waiting`,
      status: 'queued',
      eta: 'queued',
      speed: 'waiting',
      startedAt: 0,
      bytesPerSecond: 0,
    }
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

  function compareQueueItemNames(a: UploadQueueItem, b: UploadQueueItem): number {
    return a.name.localeCompare(b.name, undefined, { sensitivity: 'base', numeric: true })
  }

  function compareTransferNames(a: TransferItemData, b: TransferItemData): number {
    return a.name.localeCompare(b.name, undefined, { sensitivity: 'base', numeric: true })
  }

  function isAbortError(error: unknown): boolean {
    return error instanceof DOMException && error.name === 'AbortError'
  }

  function isSQLiteBusyError(error: unknown): boolean {
    const message = uploadErrorMessage(error).toLowerCase()
    return message.includes('sqlite_busy') || message.includes('database is locked')
  }

  function uploadErrorMessage(error: unknown): string {
    if (error instanceof Error && error.message) return error.message
    if (typeof error === 'object' && error && 'message' in error) {
      const message = (error as { message?: unknown }).message
      if (typeof message === 'string' && message) return message
    }
    return 'upload failed'
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
    visibleTransfers,
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
    moveTransferUp,
    moveTransferDown,
    canMoveTransferUp,
    canMoveTransferDown,
    cancelTransfer,
    retryTransfer,
    setOnComplete,
  }
})
