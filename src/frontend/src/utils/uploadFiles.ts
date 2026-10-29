export interface UploadFileHandle {
  kind: 'file'
  getFile(): Promise<File>
  queryPermission(options: { mode: 'read' }): Promise<PermissionState>
}

const handles = new WeakMap<File, UploadFileHandle>()
export const uploadAccept = '.cbz,.mp3,.wav,.flac,.ogg,.m4a,.aac,.mp4,.mkv,.webm,.mov,.avi,.m4v'

export function uploadFileHandle(file: File): UploadFileHandle | undefined {
  return handles.get(file)
}

// A picker is already an explicit selection gesture. Retain its handle where
// supported, without requesting additional or persistent permissions.
export async function selectUploadFiles(fallback: () => void): Promise<File[]> {
  const picker = (
    window as Window & {
      showOpenFilePicker?: (options: object) => Promise<UploadFileHandle[]>
    }
  ).showOpenFilePicker
  if (!picker) {
    fallback()
    return []
  }
  try {
    const selected = await picker.call(window, {
      multiple: true,
      types: [
        { description: 'Media', accept: { 'application/octet-stream': uploadAccept.split(',') } },
      ],
    })
    return await Promise.all(
      selected.map(async (handle) => {
        const file = await handle.getFile()
        handles.set(file, handle)
        return file
      }),
    )
  } catch (error) {
    if (!(error instanceof DOMException && error.name === 'AbortError')) fallback()
    return []
  }
}

export async function droppedUploadFiles(data: DataTransfer | null): Promise<File[]> {
  if (!data) return []
  // Capture Files and handle promises before the first await; the browser
  // releases drag data when the drop event returns.
  const entries = Array.from(data.items ?? [])
    .filter((item) => item.kind === 'file')
    .map((item) => {
      const file = item.getAsFile()
      const getHandle = (
        item as DataTransferItem & {
          getAsFileSystemHandle?: () => Promise<UploadFileHandle | { kind: 'directory' } | null>
        }
      ).getAsFileSystemHandle
      return { file, handle: getHandle?.call(item).catch(() => null) }
    })
  const fallback = Array.from(data.files ?? [])
  const files = await Promise.all(
    entries.map(async ({ file, handle }) => {
      const resolved = await handle
      if (file && resolved?.kind === 'file') handles.set(file, resolved)
      return file
    }),
  )
  const selected = files.filter((file): file is File => file !== null)
  return selected.length ? selected : fallback
}

export function matchesUploadFile(
  file: File,
  expected: { name: string; size: number; lastModified: number },
): boolean {
  return (
    file.name === expected.name &&
    file.size === expected.size &&
    file.lastModified === expected.lastModified
  )
}

export async function restoreUploadFile(job: {
  handle?: UploadFileHandle
  name: string
  size: number
  lastModified: number
}): Promise<{ file?: File; reason?: 'unavailable' | 'changed' }> {
  if (!job.handle) return { reason: 'unavailable' }
  try {
    // Only check existing access. Expired access waits for normal selection.
    if ((await job.handle.queryPermission({ mode: 'read' })) !== 'granted')
      return { reason: 'unavailable' }
    const file = await job.handle.getFile()
    if (!matchesUploadFile(file, job)) return { reason: 'changed' }
    handles.set(file, job.handle)
    return { file }
  } catch {
    return { reason: 'unavailable' }
  }
}
