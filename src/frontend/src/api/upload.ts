import { API_BASE, ApiError, fetchJson } from './client'

export interface UploadResponse {
  success: boolean
  filename: string
  folderId: number | null
  indexing?: string
}

interface ChunkedUploadInitResponse {
  uploadId: string
  chunkSize: number
  totalChunks: number
  uploadedChunks: number[]
  status: string
  resumed: boolean
  folderId: number | null
}

export interface PendingUpload {
  uploadId: string
  filename: string
  size: number
  lastModified: number
  chunkSize: number
  totalChunks: number
  uploadedChunks: number[]
  uploadedCount: number
  status: 'active' | 'paused' | 'failed' | 'assembling'
  complete: boolean
  createdAt: string
  updatedAt: string
  folderId: number | null
}

interface PendingUploadsResponse {
  uploads: PendingUpload[]
}

export async function checkUploadNames(filenames: string[]): Promise<string[]> {
  const response = await fetchJson<{ existing: string[] }>(`${API_BASE}/uploads/check`, {
    method: 'POST',
    body: JSON.stringify({ filenames }),
  })
  return response.existing
}

export async function listPendingUploads(): Promise<PendingUpload[]> {
  const response = await fetchJson<PendingUploadsResponse>(`${API_BASE}/uploads`)
  return response.uploads
}

export async function pauseUpload(uploadId: string): Promise<void> {
  await fetchJson(`${API_BASE}/uploads/${uploadId}/pause`, { method: 'POST' })
}

export async function cancelUpload(uploadId: string): Promise<void> {
  await fetchJson(`${API_BASE}/uploads/${uploadId}`, { method: 'DELETE' })
}

export async function getUploadStatus(uploadId: string): Promise<{ status: string }> {
  return fetchJson(`${API_BASE}/uploads/${uploadId}/status`)
}

export async function completeUpload(
  uploadId: string,
  signal?: AbortSignal,
): Promise<UploadResponse> {
  return fetchJson(`${API_BASE}/uploads/${uploadId}/complete`, { method: 'POST', signal })
}

export async function uploadFile(
  file: File,
  metadataProfile: string,
  folderId: number | null,
  onProgress?: (progress: number) => void,
  onUploadId?: (uploadId: string) => void | Promise<void>,
  signal?: AbortSignal,
  existingUploadId?: string,
): Promise<UploadResponse> {
  const init = await fetchJson<ChunkedUploadInitResponse>(`${API_BASE}/uploads/init`, {
    method: 'POST',
    signal,
    body: JSON.stringify({
      filename: file.name,
      size: file.size,
      profile: metadataProfile,
      lastModified: file.lastModified,
      folderId,
      uploadId: existingUploadId,
    }),
  })

  await onUploadId?.(init.uploadId)
  if (signal?.aborted) throw new DOMException('Upload aborted', 'AbortError')
  if (init.status === 'complete') {
    onProgress?.(100)
    return { success: true, filename: file.name, folderId: init.folderId }
  }

  const totalChunks = init.totalChunks || Math.ceil(file.size / init.chunkSize)
  const uploadedChunks = new Set(init.uploadedChunks ?? [])
  let confirmedUploadedBytes = uploadedBytesForChunks(uploadedChunks, file.size, init.chunkSize)
  onProgress?.(Math.round((confirmedUploadedBytes / file.size) * 100))

  for (let index = 0; index < totalChunks; index++) {
    if (uploadedChunks.has(index)) continue

    const start = index * init.chunkSize
    const end = Math.min(start + init.chunkSize, file.size)
    const chunk = file.slice(start, end)

    if (signal?.aborted) throw new DOMException('Upload aborted', 'AbortError')

    await uploadChunk(init.uploadId, index, chunk, signal, (chunkLoaded) => {
      if (!onProgress) return
      const loaded = Math.min(confirmedUploadedBytes + chunkLoaded, file.size)
      onProgress(Math.round((loaded / file.size) * 100))
    })

    uploadedChunks.add(index)
    confirmedUploadedBytes += chunk.size
    onProgress?.(Math.round((confirmedUploadedBytes / file.size) * 100))
  }

  onProgress?.(100)
  return completeUpload(init.uploadId, signal)
}

function uploadedBytesForChunks(
  uploadedChunks: Set<number>,
  fileSize: number,
  chunkSize: number,
): number {
  let bytes = 0
  for (const index of uploadedChunks) {
    const start = index * chunkSize
    const end = Math.min(start + chunkSize, fileSize)
    bytes += Math.max(end - start, 0)
  }
  return Math.min(bytes, fileSize)
}

function uploadChunk(
  uploadId: string,
  chunkIndex: number,
  chunk: Blob,
  signal: AbortSignal | undefined,
  onChunkProgress: (loaded: number) => void,
): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()

    xhr.withCredentials = true

    const abortUpload = () => xhr.abort()
    if (signal?.aborted) {
      reject(new DOMException('Upload aborted', 'AbortError'))
      return
    }
    signal?.addEventListener('abort', abortUpload, { once: true })

    xhr.upload.addEventListener('progress', (event) => {
      onChunkProgress(event.loaded)
    })

    xhr.addEventListener('load', () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        signal?.removeEventListener('abort', abortUpload)
        resolve()
      } else {
        signal?.removeEventListener('abort', abortUpload)
        reject(
          new ApiError(xhr.responseText || `Chunk upload failed: ${xhr.statusText}`, xhr.status),
        )
      }
    })

    xhr.addEventListener('error', () => {
      signal?.removeEventListener('abort', abortUpload)
      reject(new TypeError('Connection interrupted'))
    })

    xhr.addEventListener('abort', () => {
      signal?.removeEventListener('abort', abortUpload)
      reject(new DOMException('Upload aborted', 'AbortError'))
    })

    xhr.timeout = 120_000
    xhr.addEventListener('timeout', () => {
      signal?.removeEventListener('abort', abortUpload)
      reject(new ApiError('Upload timed out', 408))
    })

    xhr.open('PUT', `${API_BASE}/uploads/${uploadId}/chunk`)
    xhr.setRequestHeader('Content-Type', 'application/octet-stream')
    xhr.setRequestHeader('X-Chunk-Index', String(chunkIndex))
    xhr.send(chunk)
  })
}
