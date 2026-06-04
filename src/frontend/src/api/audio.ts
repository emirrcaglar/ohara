import { fetchJson, API_BASE } from './client'
import type { AudioLibraryResponse } from '../types/api'

export async function fetchAudioLibrary(): Promise<AudioLibraryResponse> {
  return fetchJson<AudioLibraryResponse>(`${API_BASE}/audio`)
}

export async function moveAudioToCatalog(id: number, catalogId: number | null): Promise<void> {
  await fetchJson<void>(`${API_BASE}/audio/${id}/catalog`, {
    method: 'PUT',
    body: JSON.stringify({ catalogId }),
  })
}

export function getAudioStreamUrl(id: number): string {
  return `/audio/${id}/stream`
}
