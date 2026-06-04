import { fetchJson, API_BASE } from './client'
import type { VideoInfo, VideoLibraryResponse, VideoStateUpdate } from '../types/api'

export async function fetchVideoLibrary(): Promise<VideoLibraryResponse> {
  return fetchJson<VideoLibraryResponse>(`${API_BASE}/video`)
}

export async function fetchVideoInfo(id: number): Promise<VideoInfo> {
  return fetchJson<VideoInfo>(`${API_BASE}/video/${id}/info`)
}

export async function saveVideoState(id: number, state: VideoStateUpdate): Promise<void> {
  await fetchJson<void>(`${API_BASE}/video/${id}/state`, {
    method: 'PUT',
    body: JSON.stringify(state),
  })
}

export async function deleteVideo(id: number): Promise<void> {
  await fetchJson<void>(`${API_BASE}/video/${id}`, {
    method: 'DELETE',
  })
}

export function getVideoStreamUrl(id: number): string {
  return `/video/${id}/stream`
}
