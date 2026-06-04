import { API_BASE, fetchJson } from './client'

export interface UploadResponse {
  success: boolean
  filename: string
  catalogId: number | null
  indexing?: string
}

interface ChunkedUploadInitResponse {
  uploadId: string
  chunkSize: number
  totalChunks: number
  uploadedChunks: number[]
  status: string
  resumed: boolean
  catalogId: number | null
}

export interface PendingUpload {
  uploadId: string
  filename: string
  size: number
  chunkSize: number
  totalChunks: number
  uploadedChunks: number[]
  uploadedCount: number
  status: 'active' | 'paused' | 'failed' | 'assembling'
  complete: boolean
  createdAt: string
  updatedAt: string
  catalogId: number | null
}

interface PendingUploadsResponse {
  uploads: PendingUpload[]
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

export async function uploadFile(
  file: File,
  metadataProfile: string,
  catalogId: number | null,
  onProgress?: (progress: number) => void,
  onUploadId?: (uploadId: string) => void,
  signal?: AbortSignal,
): Promise<UploadResponse> {
  const init = await fetchJson<ChunkedUploadInitResponse>(`${API_BASE}/uploads/init`, {
    method: 'POST',
    signal,
    body: JSON.stringify({
      filename: file.name,
      size: file.size,
      profile: metadataProfile,
      lastModified: file.lastModified,
      catalogId,
    }),
  })

  onUploadId?.(init.uploadId)

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
  return fetchJson<UploadResponse>(`${API_BASE}/uploads/${init.uploadId}/complete`, {
    method: 'POST',
  })
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
        reject(new Error(xhr.responseText || `Chunk upload failed: ${xhr.statusText}`))
      }
    })

    xhr.addEventListener('error', () => {
      signal?.removeEventListener('abort', abortUpload)
      reject(new Error('Chunk upload failed'))
    })

    xhr.addEventListener('abort', () => {
      signal?.removeEventListener('abort', abortUpload)
      reject(new DOMException('Upload aborted', 'AbortError'))
    })

    xhr.open('PUT', `${API_BASE}/uploads/${uploadId}/chunk`)
    xhr.setRequestHeader('Content-Type', 'application/octet-stream')
    xhr.setRequestHeader('X-Chunk-Index', String(chunkIndex))
    xhr.send(chunk)
  })
}
