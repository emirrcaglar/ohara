import { fetchJson, API_BASE } from './client'
import type { Folder, FolderListResponse } from '../types/api'

export async function fetchFolderChildren(parentId: number | null): Promise<FolderListResponse> {
  const query = parentId ? `?parentId=${parentId}` : ''
  return fetchJson<FolderListResponse>(`${API_BASE}/folder${query}`)
}

export async function fetchAllFolders(): Promise<FolderListResponse> {
  return fetchJson<FolderListResponse>(`${API_BASE}/folder/all`)
}

export async function createFolder(name: string, parentId: number | null): Promise<Folder> {
  return fetchJson<Folder>(`${API_BASE}/folder`, {
    method: 'POST',
    body: JSON.stringify({ name, parentId }),
  })
}

export async function fetchFolderById(id: number): Promise<Folder> {
  return fetchJson<Folder>(`${API_BASE}/folder/${id}`)
}

export async function updateFolder(
  id: number,
  name: string,
  parentId: number | null,
): Promise<Folder> {
  return fetchJson<Folder>(`${API_BASE}/folder/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name, parentId }),
  })
}

export async function deleteFolder(id: number): Promise<void> {
  await fetchJson<void>(`${API_BASE}/folder/${id}`, {
    method: 'DELETE',
  })
}
