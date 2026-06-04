const API_BASE = '/api'

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

async function handleResponse<T>(response: Response): Promise<T> {
  const text = await response.text()

  if (!response.ok) {
    throw new ApiError(text || response.statusText || 'Request failed', response.status)
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
