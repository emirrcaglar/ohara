import { computed, markRaw, onScopeDispose, ref, toRaw } from 'vue'
import { defineStore } from 'pinia'
import { ApiError } from '../api/client'
import {
  cancelUpload,
  checkUploadNames,
  completeUpload,
  getUploadStatus,
  listPendingUploads,
  pauseUpload,
  uploadFile,
} from '../api/upload'
import type { PendingUpload } from '../api/upload'
import { compareByNaturalText } from '../utils/naturalSort'
import { matchesUploadFile, restoreUploadFile, uploadFileHandle } from '../utils/uploadFiles'
import { uploadStorage } from '../utils/uploadStorage'
import type { SavedUpload } from '../utils/uploadStorage'

export interface TransferItemData {
  id: string
  uploadId?: string
  name: string
  progress: number
  sizeInfo?: string
  status: 'queued' | 'active' | 'complete' | 'paused' | 'failed' | 'exists'
  eta?: string
  speed?: string
  storagePath?: string
  needsFile?: boolean
  canResume?: boolean
  folderId: number | null
}

interface UploadJob extends Omit<SavedUpload, 'status'> {
  status: TransferItemData['status']
  file?: File
  sourceIssue?: 'unavailable' | 'changed'
  message?: string
  bytesPerSecond: number
  startedAt: number
  startingProgress: number
  retries: number
  retryAt: number
  readyToFinalize?: boolean
}

const extensions = [
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
const compareNames = compareByNaturalText<UploadJob>((item) => item.name)

export const useUploadStore = defineStore('upload', () => {
  const jobs = ref<UploadJob[]>([])
  const remoteTransfers = ref<TransferItemData[]>([])
  const loadingTransfers = ref(false)
  const transfersError = ref<string | null>(null)
  const storageNotice = ref<string | null>(null)
  const rejectedItems = ref<string[]>([])
  const processingUploads = ref(false)
  let userId: number | null = null
  let generation = 0
  let initialization: Promise<void> = Promise.resolve()
  let worker: Promise<void> | undefined
  let controller: AbortController | undefined
  let lockController: AbortController | undefined
  let storageAvailable = true
  let timer: number | undefined
  let writes: Promise<unknown> = Promise.resolve()
  let onCompleteCallback: (() => void) | null = null

  const queuedItems = computed(() =>
    jobs.value.filter((job) => job.status === 'queued' && job.file),
  )
  const draftItems = computed(() => jobs.value.filter((job) => !job.confirmed))
  const transfers = computed(() =>
    jobs.value.filter((job) => job.status !== 'queued').map(toTransfer),
  )
  const visibleTransfers = computed(() => {
    const local = [...jobs.value].sort((a, b) => a.order - b.order).map(toTransfer)
    return [
      ...local.filter((item) => item.status === 'active'),
      ...local.filter((item) => item.status === 'queued'),
      ...local.filter((item) => item.status !== 'active' && item.status !== 'queued'),
      ...remoteTransfers.value.filter(
        (item) => !jobs.value.some((job) => job.uploadId === item.uploadId),
      ),
    ]
  })
  const hasActiveTransfers = computed(() => visibleTransfers.value.length > 0)
  const totalBandwidth = computed(
    () =>
      ((jobs.value.reduce((sum, job) => sum + job.bytesPerSecond, 0) * 8) / 1024 ** 2).toFixed(2) +
      ' Mbps',
  )

  function savedJob(job: UploadJob): SavedUpload {
    return {
      id: job.id,
      userId: job.userId,
      uploadId: job.uploadId,
      name: job.name,
      size: job.size,
      lastModified: job.lastModified,
      folderId: job.folderId,
      order: job.order,
      confirmed: job.confirmed,
      status: job.status === 'complete' ? 'queued' : job.status,
      progress: job.progress,
      handle: job.handle && toRaw(job.handle),
    }
  }

  // Serialize mutations so an older save cannot resurrect a canceled job.
  function write(operation: () => Promise<unknown>): Promise<void> {
    const result = writes.then(operation).then(
      () => {},
      () => {
        storageAvailable = false
        storageNotice.value =
          'This browser cannot save the queue. Keep this tab open; server progress is still saved.'
      },
    )
    writes = result
    return result
  }

  function save(job: UploadJob): Promise<void> {
    const snapshot = savedJob(job)
    return write(() => uploadStorage.save(snapshot))
  }

  function forget(job: { userId: number; id: string }): Promise<void> {
    return write(() => uploadStorage.remove(job.userId, job.id))
  }

  function fromSaved(saved: SavedUpload): UploadJob {
    return {
      ...saved,
      handle: saved.handle && markRaw(saved.handle),
      status: saved.status === 'active' ? 'queued' : saved.status,
      bytesPerSecond: 0,
      startedAt: 0,
      startingProgress: 0,
      retries: 0,
      retryAt: 0,
    }
  }

  function initialize(nextUserId: number | null) {
    if (userId === nextUserId) return
    userId = nextUserId
    const version = ++generation
    controller?.abort()
    lockController?.abort()
    if (timer !== undefined) window.clearTimeout(timer)
    timer = undefined
    const previousWorker = worker
    jobs.value = []
    remoteTransfers.value = []
    rejectedItems.value = []
    transfersError.value = null
    storageNotice.value = null
    storageAvailable = true
    onCompleteCallback = null
    initialization = (async () => {
      await previousWorker
      await writes
      if (version !== generation || nextUserId === null) return
      loadingTransfers.value = true
      try {
        const saved = await uploadStorage.list(nextUserId)
        const restored = await Promise.all(
          saved
            .sort((a, b) => a.order - b.order)
            .map(async (record) => {
              const job = fromSaved(record)
              if (record.status === 'exists') return job
              const source = await restoreUploadFile(record)
              job.file = source.file && markRaw(source.file)
              job.sourceIssue = source.reason
              return job
            }),
        )
        if (version !== generation) return
        jobs.value = restored
      } catch {
        if (version === generation) {
          storageAvailable = false
          storageNotice.value =
            'This browser cannot save the queue. Server progress is still saved.'
        }
      } finally {
        if (version === generation) loadingTransfers.value = false
      }
    })()
    void initialization.then(() => {
      if (version !== generation || userId === null) return
      void fetchPendingTransfers()
      startWorker()
    })
  }

  async function enqueue(files: File[], folderId: number | null = null) {
    await initialization
    if (userId === null) return
    const version = generation
    // This request contains names only. The init endpoint repeats the check
    // before any chunks are sent, including when this preflight is offline.
    const existingNames = new Set(
      await checkUploadNames(files.map((file) => file.name)).catch(() => []),
    )
    if (version !== generation) return
    rejectedItems.value = []
    const added: UploadJob[] = []
    for (const file of files) {
      const ext = file.name.slice(file.name.lastIndexOf('.')).toLowerCase()
      if (!extensions.includes(ext) && !file.type.startsWith('video/')) {
        rejectedItems.value.push(file.name)
        continue
      }
      // Reconnect an interrupted source to its original destination, even when
      // the library is currently showing a different folder.
      const interrupted = jobs.value.find(
        (job) =>
          !['complete', 'exists'].includes(job.status) && !job.file && matchesUploadFile(file, job),
      )
      if (interrupted) {
        attachFile(interrupted, file)
        await save(interrupted)
        continue
      }
      const exists =
        existingNames.has(file.name) ||
        [...jobs.value, ...added].some(
          (job) =>
            !['complete', 'exists'].includes(job.status) &&
            job.name.toLowerCase() === file.name.toLowerCase(),
        )
      const handle = uploadFileHandle(file)
      added.push({
        id: 'local-' + Date.now() + '-' + Math.random().toString(16).slice(2),
        userId,
        name: file.name,
        size: file.size,
        lastModified: file.lastModified,
        folderId,
        order: 0,
        confirmed: false,
        status: exists ? 'exists' : 'queued',
        progress: 0,
        file: exists ? undefined : markRaw(file),
        handle: !exists && handle ? markRaw(handle) : undefined,
        bytesPerSecond: 0,
        startedAt: 0,
        startingProgress: 0,
        retries: 0,
        retryAt: 0,
      })
    }
    const startOrder = Math.max(-1, ...jobs.value.map((job) => job.order)) + 1
    added.sort(compareNames).forEach((job, index) => {
      job.order = startOrder + index
    })
    jobs.value.push(...added)
    await Promise.all(added.map(save))
    startWorker()
  }

  function attachFile(job: UploadJob, file: File) {
    const handle = uploadFileHandle(file)
    job.file = markRaw(file)
    job.handle = handle && markRaw(handle)
    job.sourceIssue = undefined
    job.message = undefined
    job.retries = 0
    job.retryAt = 0
    if (job.status === 'failed') job.status = 'paused'
  }

  async function reselectTransfers(files: File[]) {
    await initialization
    if (userId === null || files.length === 0) return
    const version = generation
    const matched: UploadJob[] = []
    let remoteError: string | null = null
    let pending: PendingUpload[] = []

    // Fetch server-only records once for the whole selection. Existing local
    // jobs already contain the source identity and original destination.
    if (
      remoteTransfers.value.some(
        (item) => item.needsFile && !jobs.value.some((job) => job.uploadId === item.uploadId),
      )
    ) {
      try {
        pending = await listPendingUploads()
      } catch {
        remoteError = 'Unable to check saved uploads. Try selecting their files again.'
      }
    }
    if (version !== generation) return

    for (const job of jobs.value) {
      if (job.file || job.readyToFinalize || ['complete', 'exists'].includes(job.status)) continue
      const file = files.find((file) => matchesUploadFile(file, job))
      if (!file) continue
      attachFile(job, file)
      job.confirmed = true
      job.status = 'queued'
      matched.push(job)
    }

    for (const upload of pending) {
      if (!remoteTransfers.value.some((item) => item.uploadId === upload.uploadId)) continue
      if (jobs.value.some((job) => job.uploadId === upload.uploadId)) continue
      const file = files.find((file) =>
        matchesUploadFile(file, {
          name: upload.filename,
          size: upload.size,
          lastModified: upload.lastModified,
        }),
      )
      if (!file) continue
      const job = fromSaved({
        id: 'local-' + upload.uploadId,
        userId,
        uploadId: upload.uploadId,
        name: upload.filename,
        size: upload.size,
        lastModified: upload.lastModified,
        folderId: upload.folderId,
        order: Math.max(-1, ...jobs.value.map((job) => job.order)) + 1,
        status: 'queued',
        confirmed: true,
        progress: upload.totalChunks
          ? Math.round((upload.uploadedCount / upload.totalChunks) * 100)
          : 0,
      })
      attachFile(job, file)
      jobs.value.push(job)
      matched.push(job)
    }

    if (!matched.length) {
      transfersError.value =
        remoteError ?? 'No matching files. Select the original, unchanged files to resume.'
      return
    }
    await Promise.all(matched.map(save))
    if (version !== generation) return
    transfersError.value = remoteError
    startWorker()
  }

  async function clearQueue() {
    const drafts = jobs.value.filter((job) => !job.confirmed)
    jobs.value = jobs.value.filter((job) => job.confirmed)
    rejectedItems.value = []
    await Promise.all(drafts.map(forget))
  }

  async function fetchPendingTransfers() {
    await initialization
    if (userId === null) return
    const version = generation
    loadingTransfers.value = true
    try {
      const pending = await listPendingUploads()
      if (version === generation) {
        remoteTransfers.value = pending.map(pendingTransfer)
        transfersError.value = null
        for (const job of jobs.value) {
          if (['active', 'complete', 'exists'].includes(job.status)) continue
          const server = pending.find((item) => item.uploadId === job.uploadId)
          if (server) {
            job.progress = server.totalChunks
              ? Math.round((server.uploadedCount / server.totalChunks) * 100)
              : 0
            job.readyToFinalize = server.complete
          } else if (job.uploadId) {
            // The last response can be lost after the server finished. Check
            // that checkpoint even when local file access has already expired.
            const result = await getUploadStatus(job.uploadId).catch(() => undefined)
            if (version !== generation) return
            if (result?.status === 'complete') await finishJob(job)
            else if (result?.status === 'cancelled') {
              jobs.value = jobs.value.filter((item) => item.id !== job.id)
              await forget(job)
            }
          }
        }
      }
    } catch (error) {
      if (version === generation) transfersError.value = errorMessage(error)
    } finally {
      if (version === generation) {
        loadingTransfers.value = false
        startWorker()
      }
    }
  }

  async function processAll() {
    await initialization
    const waiting = jobs.value.filter(
      (job) => job.status === 'queued' || (!job.confirmed && job.status === 'exists'),
    )
    waiting.forEach((job) => {
      job.confirmed = true
    })
    await Promise.all(waiting.map(save))
    startWorker()
  }

  function startWorker() {
    if (worker || userId === null) return
    const version = generation
    processingUploads.value = true
    worker = runWorker(version)
      .catch((error) => {
        if (version === generation) transfersError.value = errorMessage(error)
      })
      .finally(() => {
        worker = undefined
        processingUploads.value = false
        if (version === generation) scheduleResume()
      })
  }

  async function runWorker(version: number) {
    const run = async () => {
      while (version === generation) {
        const job = [...jobs.value]
          .sort((a, b) => a.order - b.order)
          .find(
            (job) =>
              job.confirmed &&
              job.status === 'queued' &&
              (job.file || job.readyToFinalize) &&
              job.retryAt <= Date.now(),
          )
        if (!job) break
        // Another tab may have finished/canceled/paused while we waited for the lock.
        try {
          await writes
          const saved = storageAvailable
            ? await uploadStorage.get(job.userId, job.id)
            : savedJob(job)
          if (version !== generation) break
          if (!saved) {
            jobs.value = jobs.value.filter((item) => item.id !== job.id)
            continue
          }
          if (['paused', 'failed', 'exists'].includes(saved.status)) {
            job.status = saved.status
            continue
          }
          job.uploadId = saved.uploadId
        } catch {
          /* Storage can be disabled; in-memory uploads still work. */
        }
        if (version !== generation) break
        await sendJob(job, version)
      }
    }
    if (navigator.locks) {
      const pendingLock = new AbortController()
      lockController = pendingLock
      try {
        await navigator.locks.request(
          'ohara-uploads-' + userId,
          { signal: pendingLock.signal },
          run,
        )
      } finally {
        if (lockController === pendingLock) lockController = undefined
      }
    } else {
      await run()
    }
  }

  async function sendJob(job: UploadJob, version: number) {
    const file = job.file
    if (!file && !job.readyToFinalize) return
    if (job.handle && !job.readyToFinalize) {
      const source = await restoreUploadFile(job)
      if (version !== generation) return
      if (!source.file) {
        job.file = undefined
        job.sourceIssue = source.reason
        return
      }
      job.file = markRaw(source.file)
    }
    if (job.status !== 'queued' || !jobs.value.includes(job)) return
    job.status = 'active'
    job.startedAt = performance.now()
    job.startingProgress = job.progress
    job.message = undefined
    controller = new AbortController()
    const activeController = controller
    await save(job)
    let firstProgress = true
    try {
      if (job.readyToFinalize && job.uploadId) {
        await completeUpload(job.uploadId, activeController.signal)
      } else {
        await uploadFile(
          (job.file ?? file) as File,
          '',
          job.folderId,
          (progress) => {
            if (version !== generation || job.status !== 'active') return
            // The first callback is confirmed server progress; don't count old bytes.
            if (firstProgress) {
              job.startingProgress = progress
              job.startedAt = performance.now()
              firstProgress = false
            }
            job.progress = progress
            const seconds = (performance.now() - job.startedAt) / 1000
            job.bytesPerSecond =
              seconds > 0
                ? ((Math.max(0, progress - job.startingProgress) / 100) * job.size) / seconds
                : 0
          },
          async (uploadId) => {
            job.uploadId = uploadId
            if (version === generation && !jobs.value.includes(job)) {
              await cancelUpload(uploadId).catch(() => {})
              throw new DOMException('Upload canceled', 'AbortError')
            }
            // Save the server ID before sending the first chunk.
            await save(job)
          },
          activeController.signal,
          job.uploadId,
        )
      }
      if (version !== generation || !jobs.value.includes(job)) return
      await finishJob(job)
    } catch (error) {
      if (version !== generation || !jobs.value.includes(job) || job.status !== 'active') return
      if (isExistingUploadError(error)) {
        job.status = 'exists'
        job.file = undefined
        job.handle = undefined
        job.progress = 0
        job.readyToFinalize = false
      } else if (isTransient(error)) {
        job.retries++
        job.retryAt = Date.now() + Math.min(30_000, 1500 * 2 ** Math.min(job.retries - 1, 5))
        job.status = 'queued'
        job.message = navigator.onLine
          ? 'Connection interrupted · retrying'
          : 'Waiting for connection'
      } else {
        job.status = 'failed'
        job.message = errorMessage(error)
      }
      await save(job)
    } finally {
      job.bytesPerSecond = 0
      if (controller === activeController) controller = undefined
    }
  }

  async function finishJob(job: UploadJob) {
    job.status = 'complete'
    job.progress = 100
    job.file = undefined
    job.handle = undefined
    await forget(job)
    onCompleteCallback?.()
  }

  function scheduleResume() {
    if (timer !== undefined) window.clearTimeout(timer)
    timer = undefined
    const waiting = jobs.value.filter(
      (job) => job.confirmed && job.status === 'queued' && (job.file || job.readyToFinalize),
    )
    if (!waiting.length) return
    const delay = Math.max(0, Math.min(...waiting.map((job) => job.retryAt)) - Date.now())
    timer = window.setTimeout(() => {
      timer = undefined
      startWorker()
    }, delay)
  }

  async function pauseTransfer(id: string) {
    const job = jobs.value.find((job) => job.id === id)
    if (!job || ['complete', 'exists'].includes(job.status)) return
    const wasActive = job.status === 'active'
    job.status = 'paused'
    job.message = undefined
    job.bytesPerSecond = 0
    if (wasActive) controller?.abort()
    await save(job)
    if (job.uploadId) await pauseUpload(job.uploadId).catch(() => {})
  }

  async function retryTransfer(id: string) {
    const job = jobs.value.find((job) => job.id === id)
    if (
      !job ||
      (!job.file && !job.readyToFinalize) ||
      !['queued', 'paused', 'failed'].includes(job.status)
    )
      return
    job.status = 'queued'
    job.confirmed = true
    job.message = undefined
    job.retryAt = 0
    job.retries = 0
    await save(job)
    startWorker()
  }

  async function cancelTransfer(id: string) {
    const job = jobs.value.find((job) => job.id === id)
    const uploadId = job?.uploadId ?? remoteTransfers.value.find((item) => item.id === id)?.uploadId
    if (job) {
      const wasActive = job.status === 'active'
      jobs.value = jobs.value.filter((item) => item.id !== id)
      if (wasActive) controller?.abort()
      await forget(job)
    }
    remoteTransfers.value = remoteTransfers.value.filter((item) => item.uploadId !== uploadId)
    if (uploadId) {
      try {
        await cancelUpload(uploadId)
      } catch {
        transfersError.value =
          'Removed from this browser. Reconnect and cancel the saved server upload to discard its chunks.'
      }
    }
  }

  function orderedWaiting() {
    return [...jobs.value]
      .filter((job) => job.status === 'queued' || job.status === 'active')
      .sort((a, b) => a.order - b.order)
  }

  function canMoveTransferUp(id: string) {
    return orderedWaiting().findIndex((job) => job.id === id) > 0
  }
  function canMoveTransferDown(id: string) {
    const queue = orderedWaiting()
    const index = queue.findIndex((job) => job.id === id)
    return index >= 0 && index < queue.length - 1
  }

  async function moveTransfer(id: string, direction: number) {
    const queue = orderedWaiting()
    const index = queue.findIndex((job) => job.id === id)
    const job = queue[index]
    const adjacent = queue[index + direction]
    if (!job || !adjacent) return
    ;[job.order, adjacent.order] = [adjacent.order, job.order]
    const active = [job, adjacent].find((item) => item.status === 'active')
    if (active) {
      active.status = 'queued'
      active.bytesPerSecond = 0
      controller?.abort()
    }
    await Promise.all([save(job), save(adjacent)])
    startWorker()
  }

  function toTransfer(job: UploadJob): TransferItemData {
    const needsFile =
      !job.file && !job.readyToFinalize && !['complete', 'exists'].includes(job.status)
    const status = needsFile ? 'paused' : job.status
    const remaining =
      job.bytesPerSecond > 0
        ? Math.ceil((job.size * (1 - job.progress / 100)) / job.bytesPerSecond)
        : 0
    return {
      id: job.id,
      uploadId: job.uploadId,
      name: job.name,
      progress: job.progress,
      folderId: job.folderId,
      status,
      needsFile,
      sizeInfo: formatFileSize(job.size),
      canResume:
        job.status === 'paused' ||
        job.status === 'failed' ||
        (job.status === 'queued' && !job.confirmed),
      eta: needsFile
        ? job.sourceIssue === 'changed'
          ? 'File changed · select the original'
          : 'Select the original file to continue'
        : (job.message ??
          (status === 'paused'
            ? 'Paused'
            : status === 'queued'
              ? job.confirmed
                ? 'Waiting'
                : 'Ready to upload'
              : job.progress === 100 && status === 'active'
                ? 'Finalizing'
                : remaining
                  ? formatEta(remaining)
                  : '--')),
      speed: job.bytesPerSecond > 0 ? (job.bytesPerSecond / 1024 ** 2).toFixed(2) + ' MB/s' : '--',
    }
  }

  function pendingTransfer(upload: PendingUpload): TransferItemData {
    return {
      id: upload.uploadId,
      uploadId: upload.uploadId,
      name: upload.filename,
      folderId: upload.folderId,
      progress: upload.totalChunks
        ? Math.round((upload.uploadedCount / upload.totalChunks) * 100)
        : 0,
      sizeInfo: formatFileSize(upload.size),
      status: 'paused',
      needsFile: true,
      eta: 'Select the original file to continue',
      speed: '--',
    }
  }

  function online() {
    for (const job of jobs.value) job.retryAt = 0
    startWorker()
  }
  window.addEventListener('online', online)
  onScopeDispose(() => {
    initialize(null)
    window.removeEventListener('online', online)
  })

  return {
    queuedItems,
    draftItems,
    transfers,
    visibleTransfers,
    loadingTransfers,
    transfersError,
    storageNotice,
    processingUploads,
    rejectedItems,
    totalBandwidth,
    hasActiveTransfers,
    initialize,
    enqueue,
    clearQueue,
    fetchPendingTransfers,
    processAll,
    pauseTransfer,
    cancelTransfer,
    retryTransfer,
    reselectTransfers,
    canMoveTransferUp,
    canMoveTransferDown,
    moveTransferUp: (id: string) => moveTransfer(id, -1),
    moveTransferDown: (id: string) => moveTransfer(id, 1),
    setOnComplete: (fn: (() => void) | null) => {
      onCompleteCallback = fn
    },
  }
})

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Upload failed'
}

function isExistingUploadError(error: unknown): boolean {
  if (!(error instanceof ApiError) || error.status !== 409) return false
  try {
    return JSON.parse(error.message).code === 'already_exists'
  } catch {
    return false
  }
}

function isTransient(error: unknown): boolean {
  const message = errorMessage(error).toLowerCase()
  return (
    error instanceof TypeError ||
    (error instanceof ApiError &&
      (error.status >= 500 || error.status === 408 || error.status === 429)) ||
    message.includes('sqlite_busy') ||
    message.includes('database is locked')
  )
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 ** 2) return (bytes / 1024).toFixed(1) + ' KB'
  if (bytes < 1024 ** 3) return (bytes / 1024 ** 2).toFixed(1) + ' MB'
  return (bytes / 1024 ** 3).toFixed(2) + ' GB'
}

function formatEta(seconds: number): string {
  return (
    String(Math.floor(seconds / 60)).padStart(2, '0') + ':' + String(seconds % 60).padStart(2, '0')
  )
}
