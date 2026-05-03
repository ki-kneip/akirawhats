import { authStore } from '../store/auth'

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const token = authStore.getState().token
  const authHeaders: Record<string, string> = token
    ? { Authorization: `Bearer ${token}` }
    : {}

  const res = await fetch(`${BASE_URL}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...authHeaders,
      ...init?.headers,
    },
    ...init,
  })

  if (!res.ok) {
    if (res.status === 401) {
      authStore.getState().clearAuth()
    }
    let message = res.statusText
    try {
      const body = await res.json()
      if (body?.error) message = body.error
    } catch {}
    throw new Error(message)
  }

  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: "POST", body: JSON.stringify(body) }),
  put: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: "PUT", body: JSON.stringify(body) }),
  delete: <T>(path: string) => request<T>(path, { method: "DELETE" }),
}
