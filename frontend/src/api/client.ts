const API_BASE = '/api'

interface ApiError {
  message: string
  status: number
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const error: ApiError = {
      message: response.statusText || 'Request failed',
      status: response.status
    }
    throw error
  }
  return response.json()
}

export async function fetchJson<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers
    }
  })
  return handleResponse<T>(response)
}

export { API_BASE }
