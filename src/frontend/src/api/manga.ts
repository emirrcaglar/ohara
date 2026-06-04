import { fetchJson, API_BASE } from './client'
import type { MangaLibraryResponse, MangaInfo } from '../types/api'

export async function fetchMangaLibrary(): Promise<MangaLibraryResponse> {
  return fetchJson<MangaLibraryResponse>(`${API_BASE}/manga`)
}

export async function fetchMangaInfo(id: number): Promise<MangaInfo> {
  return fetchJson<MangaInfo>(`${API_BASE}/manga/${id}/info`)
}

export async function saveMangaProgress(id: number, page: number): Promise<void> {
  await fetchJson<void>(`${API_BASE}/manga/${id}/progress/${page}`, {
    method: 'POST',
  })
}

export async function moveMangaToCatalog(id: number, catalogId: number | null): Promise<void> {
  await fetchJson<void>(`${API_BASE}/manga/${id}/catalog`, {
    method: 'PUT',
    body: JSON.stringify({ catalogId }),
  })
}

export async function deleteManga(id: number): Promise<void> {
  await fetchJson<void>(`${API_BASE}/manga/${id}`, {
    method: 'DELETE',
  })
}

export function getMangaPageUrl(id: number, page: number): string {
  return `${API_BASE}/manga/${id}/page/${page}`
}
