import { api } from './client'

export interface AuthResponse {
  token: string
  user: {
    id: string
    email: string
    first_name: string
    last_name: string
  }
}

export function login(email: string, password: string) {
  return api.post<AuthResponse>('/api/auth/login', { email, password })
}

export function register(firstName: string, lastName: string, email: string, password: string) {
  return api.post<AuthResponse>('/api/auth/register', {
    first_name: firstName,
    last_name: lastName,
    email,
    password,
  })
}
