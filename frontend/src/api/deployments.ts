import { fetchJson, API_BASE } from './client'

export interface Deployment {
  id: number
  deployedAt: string
  createdAt: string
}

interface LatestDeploymentResponse {
  deployment: Deployment | null
}

export async function fetchLatestDeployment(): Promise<Deployment | null> {
  const response = await fetchJson<LatestDeploymentResponse>(`${API_BASE}/deployments/latest`)
  return response.deployment
}
