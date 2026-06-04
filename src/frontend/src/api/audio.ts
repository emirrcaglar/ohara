import { fetchJson, API_BASE } from './client'
import type { AudioLibraryResponse } from '../types/api'

export async function fetchAudioLibrary(): Promise<AudioLibraryResponse> {
  return fetchJson<AudioLibraryResponse>(`${API_BASE}/audio`)
}

export function getAudioStreamUrl(id: number): string {
  return `/audio/${id}/stream`
}
