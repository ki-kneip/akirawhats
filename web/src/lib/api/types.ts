export type InstanceStatus = "disconnected" | "connecting" | "qr" | "connected" | "logged_out"

export interface Instance {
  id: string
  status: InstanceStatus
  phone?: string
  qr?: string
}

export interface User {
  id: string
  name: string
  email: string
}

export interface SendTextResponse {
  id: string
  timestamp: number
}

export interface ApiError {
  error: string
}

export interface Message {
  id: string
  from: string
  body: string
  timestamp: string
}
