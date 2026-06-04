import { fetchJson, API_BASE } from './client'

export interface PreferencesResponse {
  preferences: Record<string, string>
}

export interface PreferenceResponse {
  key: string
  value: string
}

export async function fetchPreferences(): Promise<PreferencesResponse> {
  return fetchJson<PreferencesResponse>(`${API_BASE}/preferences`)
}

export async function savePreference(key: string, value: string): Promise<PreferenceResponse> {
  return fetchJson<PreferenceResponse>(`${API_BASE}/preferences/${encodeURIComponent(key)}`, {
    method: 'PUT',
    body: JSON.stringify({ value }),
  })
}
