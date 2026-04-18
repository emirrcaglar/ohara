import { fetchJson, API_BASE } from './client'
import type { MangaLibraryResponse, MangaInfo } from '../types/api'

export async function fetchMangaLibrary(): Promise<MangaLibraryResponse> {
  return fetchJson<MangaLibraryResponse>(`${API_BASE}/manga`)
}

export async function fetchMangaInfo(id: number): Promise<MangaInfo> {
  return fetchJson<MangaInfo>(`${API_BASE}/manga/${id}/info`)
}

export async function saveMangaProgress(id: number, page: number): Promise<void> {
  await fetch(`${API_BASE}/manga/${id}/progress/${page}`, {
    method: 'POST'
  })
}

export function getMangaCoverUrl(id: number): string {
  return `/manga/${id}/page/0`
}

export function getMangaPageUrl(id: number, page: number): string {
  return `/manga/${id}/page/${page}`
}
