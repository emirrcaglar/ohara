import { API_BASE, fetchJson } from './client'

export interface UploadResponse {
  success: boolean
  filename: string
  indexing?: string
}

interface ChunkedUploadInitResponse {
  uploadId: string
  chunkSize: number
}

export async function uploadFile(
  file: File,
  metadataProfile: string,
  onProgress?: (progress: number) => void,
): Promise<UploadResponse> {
  const init = await fetchJson<ChunkedUploadInitResponse>(`${API_BASE}/uploads/init`, {
    method: 'POST',
    body: JSON.stringify({
      filename: file.name,
      size: file.size,
      profile: metadataProfile,
    }),
  })

  const totalChunks = Math.ceil(file.size / init.chunkSize)

  try {
    for (let index = 0; index < totalChunks; index++) {
      const start = index * init.chunkSize
      const end = Math.min(start + init.chunkSize, file.size)
      const chunk = file.slice(start, end)

      await uploadChunk(init.uploadId, index, chunk, (chunkLoaded) => {
        if (!onProgress) return
        const loaded = Math.min(start + chunkLoaded, file.size)
        onProgress(Math.round((loaded / file.size) * 100))
      })
    }

    onProgress?.(100)
    return fetchJson<UploadResponse>(`${API_BASE}/uploads/${init.uploadId}/complete`, {
      method: 'POST',
    })
  } catch (error) {
    try {
      await fetchJson(`${API_BASE}/uploads/${init.uploadId}`, { method: 'DELETE' })
    } catch {
      // Ignore cleanup failures; return the original upload error.
    }
    throw error
  }
}

function uploadChunk(
  uploadId: string,
  chunkIndex: number,
  chunk: Blob,
  onChunkProgress: (loaded: number) => void,
): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()

    xhr.withCredentials = true
    xhr.upload.addEventListener('progress', (event) => {
      onChunkProgress(event.loaded)
    })

    xhr.addEventListener('load', () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve()
      } else {
        reject(new Error(xhr.responseText || `Chunk upload failed: ${xhr.statusText}`))
      }
    })

    xhr.addEventListener('error', () => {
      reject(new Error('Chunk upload failed'))
    })

    xhr.open('PUT', `${API_BASE}/uploads/${uploadId}/chunk`)
    xhr.setRequestHeader('Content-Type', 'application/octet-stream')
    xhr.setRequestHeader('X-Chunk-Index', String(chunkIndex))
    xhr.send(chunk)
  })
}
