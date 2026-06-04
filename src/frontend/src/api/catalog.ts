import { fetchJson, API_BASE } from './client'
import type { CatalogFolder, CatalogListResponse } from '../types/api'

export async function fetchCatalog(parentId: number | null): Promise<CatalogListResponse> {
  const query = parentId ? `?parentId=${parentId}` : ''
  return fetchJson<CatalogListResponse>(`${API_BASE}/catalog${query}`)
}

export async function fetchAllCatalogs(): Promise<CatalogListResponse> {
  return fetchJson<CatalogListResponse>(`${API_BASE}/catalog/all`)
}

export async function createCatalogFolder(
  name: string,
  parentId: number | null,
): Promise<CatalogFolder> {
  return fetchJson<CatalogFolder>(`${API_BASE}/catalog`, {
    method: 'POST',
    body: JSON.stringify({ name, parentId }),
  })
}

export async function fetchCatalogFolder(id: number): Promise<CatalogFolder> {
  return fetchJson<CatalogFolder>(`${API_BASE}/catalog/${id}`)
}

export async function updateCatalogFolder(
  id: number,
  name: string,
  parentId: number | null,
): Promise<CatalogFolder> {
  return fetchJson<CatalogFolder>(`${API_BASE}/catalog/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ name, parentId }),
  })
}

export async function deleteCatalogFolder(id: number): Promise<void> {
  await fetchJson<void>(`${API_BASE}/catalog/${id}`, {
    method: 'DELETE',
  })
}
