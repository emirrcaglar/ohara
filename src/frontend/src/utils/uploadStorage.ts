import type { UploadFileHandle } from './uploadFiles'

// Store references and small queue records only, never file contents.
export interface SavedUpload {
  id: string
  userId: number
  uploadId?: string
  name: string
  size: number
  lastModified: number
  folderId: number | null
  order: number
  confirmed: boolean
  status: 'queued' | 'active' | 'paused' | 'failed' | 'exists'
  progress: number
  handle?: UploadFileHandle
}

let connection: Promise<IDBDatabase> | undefined

function database(): Promise<IDBDatabase> {
  if (!connection) {
    connection = new Promise<IDBDatabase>((resolve, reject) => {
      const request = indexedDB.open('ohara-uploads', 1)
      request.onupgradeneeded = () => {
        const jobs = request.result.createObjectStore('jobs', { keyPath: ['userId', 'id'] })
        jobs.createIndex('userId', 'userId')
      }
      request.onsuccess = () => {
        const db = request.result
        db.onversionchange = () => {
          db.close()
          connection = undefined
        }
        resolve(db)
      }
      request.onerror = () => reject(request.error)
      request.onblocked = () => reject(new Error('Upload storage is unavailable'))
    }).catch((error) => {
      connection = undefined
      throw error
    })
  }
  return connection
}

async function transaction<T>(
  mode: IDBTransactionMode,
  operation: (store: IDBObjectStore) => IDBRequest<T>,
): Promise<T> {
  const db = await database()
  return new Promise((resolve, reject) => {
    const tx = db.transaction('jobs', mode)
    const request = operation(tx.objectStore('jobs'))
    tx.oncomplete = () => resolve(request.result)
    tx.onabort = () => reject(tx.error ?? new Error('Upload storage is unavailable'))
  })
}

export const uploadStorage = {
  list(userId: number): Promise<SavedUpload[]> {
    return transaction('readonly', (store) => store.index('userId').getAll(userId))
  },
  get(userId: number, id: string): Promise<SavedUpload | undefined> {
    return transaction('readonly', (store) => store.get([userId, id]))
  },
  async save(job: SavedUpload): Promise<void> {
    try {
      await transaction('readwrite', (store) => store.put(job))
    } catch (error) {
      // Some browsers cannot serialize handles. Retain the queue for normal
      // file reselection instead.
      if (!job.handle || !(error instanceof DOMException) || error.name !== 'DataCloneError')
        throw error
      await transaction('readwrite', (store) => store.put({ ...job, handle: undefined }))
    }
  },
  async remove(userId: number, id: string): Promise<void> {
    await transaction('readwrite', (store) => store.delete([userId, id]))
  },
}
