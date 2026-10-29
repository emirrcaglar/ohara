import { fetchJson, API_BASE } from './client'
import type { AudioLibraryResponse } from '../types/api'

export async function fetchAudioLibrary(): Promise<AudioLibraryResponse> {
  return fetchJson<AudioLibraryResponse>(`${API_BASE}/audio`)
}

export async function moveAudioToFolder(id: number, folderId: number | null): Promise<void> {
  await fetchJson<void>(`${API_BASE}/audio/${id}/folder`, {
    method: 'PUT',
    body: JSON.stringify({ folderId }),
  })
}

export function getAudioStreamUrl(id: number): string {
  return `/audio/${id}/stream`
}
