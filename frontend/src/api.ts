import type { AuthResponse, User, WeightRecord, WeightsResponse } from './types'

const API_BASE = ''

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
    ...init,
  })

  const data = await response.json().catch(() => null)
  if (!response.ok) {
    throw new Error(data?.message ?? '请求失败')
  }

  return data as T
}

export function getCurrentUser(): Promise<AuthResponse> {
  return request<AuthResponse>('/api/me')
}

export function register(payload: { username: string; password: string }): Promise<AuthResponse> {
  return request<AuthResponse>('/api/register', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function login(payload: { username: string; password: string }): Promise<AuthResponse> {
  return request<AuthResponse>('/api/login', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function logout(): Promise<{ ok: boolean }> {
  return request<{ ok: boolean }>('/api/logout', {
    method: 'POST',
  })
}

export async function getWeights(month: string): Promise<WeightRecord[]> {
  const response = await request<WeightsResponse>(`/api/weights?month=${month}`)
  return response.records
}

export function saveWeight(payload: { date: string; weight: number }): Promise<WeightRecord> {
  return request<WeightRecord>('/api/weights', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export type { User, WeightRecord }
