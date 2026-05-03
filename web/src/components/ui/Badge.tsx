import type { InstanceStatus } from "@/lib/api/types"

const styles: Record<InstanceStatus, string> = {
  connected: "bg-green-100 text-green-800",
  connecting: "bg-yellow-100 text-yellow-800",
  qr: "bg-blue-100 text-blue-800",
  disconnected: "bg-zinc-100 text-zinc-600",
  logged_out: "bg-red-100 text-red-700",
}

const labels: Record<InstanceStatus, string> = {
  connected: "Conectado",
  connecting: "Conectando",
  qr: "Aguardando QR",
  disconnected: "Desconectado",
  logged_out: "Deslogado",
}

export function StatusBadge({ status }: { status: InstanceStatus }) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${styles[status]}`}
    >
      {labels[status]}
    </span>
  )
}
