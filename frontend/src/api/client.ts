const API_BASE = '/api'

interface ApiError {
  message: string
  status: number
}

async function handleResponse<T>(response: Response): Promise<T> {
  const text = await response.text()

  if (!response.ok) {
    const error: ApiError = {
      message: text || response.statusText || 'Request failed',
      status: response.status,
    }
    throw error
  }

  if (!text) {
    return undefined as T
  }

  return JSON.parse(text) as T
}

export async function fetchJson<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    credentials: 'include',
  })
  return handleResponse<T>(response)
}

export { API_BASE }
