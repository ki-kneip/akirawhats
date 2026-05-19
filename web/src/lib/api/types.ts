export type InstanceStatus = "disconnected" | "connecting" | "qr" | "connected" | "logged_out"

export interface Instance {
  id: string
  status: InstanceStatus
  phone?: string
  qr?: string
  webhookUrl?: string
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
  direction: "in" | "out"
  status: "sent" | "delivered" | "read"
}

export interface GroupInfo {
  jid: string
  name: string
}
