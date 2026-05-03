"use client"

import { use, useState } from "react"
import { useInstance, useSendText, useDeleteInstance, useInstanceMessages } from "@/lib/api/instances"
import { StatusBadge } from "@/components/ui/Badge"
import { Button } from "@/components/ui/Button"
import { Spinner } from "@/components/ui/Spinner"
import { QRCodeDisplay } from "@/components/instances/QRCodeDisplay"
import { useRouter } from "next/navigation"

export default function InstanceDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = use(params)
  const router = useRouter()
  const { data: instance, isLoading } = useInstance(id)
  const { mutate: deleteInstance } = useDeleteInstance()
  const { mutate: sendText, isPending: sending, data: sent, error: sendError } = useSendText(id)
  const { data: messages = [] } = useInstanceMessages(id)

  const [to, setTo] = useState("")
  const [message, setMessage] = useState("")
  const [showQR, setShowQR] = useState(false)

  function handleDelete() {
    if (!confirm(`Remover instância "${id}"?`)) return
    deleteInstance(id, { onSuccess: () => router.push("/instances") })
  }

  function handleSend() {
    if (!to.trim() || !message.trim()) return
    sendText(
      { to: to.trim(), message: message.trim() },
      { onSuccess: () => setMessage("") }
    )
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Spinner className="size-8 text-zinc-400" />
      </div>
    )
  }

  if (!instance) {
    return (
      <p className="text-sm text-zinc-500">Instância não encontrada.</p>
    )
  }

  const isConnected = instance.status === "connected"

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      {/* Info */}
      <div className="rounded-xl border border-zinc-200 bg-white p-5 shadow-sm">
        <div className="flex items-start justify-between">
          <div>
            <p className="text-xl font-semibold text-zinc-900">{instance.id}</p>
            {instance.phone && (
              <p className="mt-1 text-sm text-zinc-500">{instance.phone}</p>
            )}
          </div>
          <StatusBadge status={instance.status} />
        </div>

        <div className="mt-4 flex gap-2">
          {!isConnected && (
            <Button size="sm" onClick={() => setShowQR(true)}>
              Conectar via QR
            </Button>
          )}
          <Button variant="danger" size="sm" onClick={handleDelete}>
            Remover instância
          </Button>
        </div>
      </div>

      {/* Enviar mensagem de texto */}
      <div className="rounded-xl border border-zinc-200 bg-white p-5 shadow-sm">
        <h2 className="mb-4 font-semibold text-zinc-900">Enviar mensagem de texto</h2>

        {!isConnected && (
          <div className="mb-4 rounded-lg border border-yellow-200 bg-yellow-50 p-3 text-sm text-yellow-800">
            Instância desconectada. Conecte antes de enviar mensagens.
          </div>
        )}

        <div className="space-y-3">
          <div>
            <label className="mb-1 block text-xs font-medium text-zinc-600">Para (número com DDI)</label>
            <input
              type="text"
              placeholder="5511999999999"
              value={to}
              onChange={(e) => setTo(e.target.value)}
              disabled={!isConnected}
              className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-green-500 focus:ring-1 focus:ring-green-500 disabled:bg-zinc-50 disabled:text-zinc-400"
            />
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium text-zinc-600">Mensagem</label>
            <textarea
              rows={3}
              placeholder="Digite a mensagem..."
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              disabled={!isConnected}
              className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-green-500 focus:ring-1 focus:ring-green-500 disabled:bg-zinc-50 disabled:text-zinc-400"
            />
          </div>

          {sendError && (
            <p className="text-sm text-red-600">{sendError.message}</p>
          )}

          {sent && (
            <p className="text-sm text-green-600">
              Enviado! ID: {sent.id}
            </p>
          )}

          <Button
            onClick={handleSend}
            disabled={!isConnected || sending || !to.trim() || !message.trim()}
          >
            {sending && <Spinner className="size-4" />}
            Enviar
          </Button>
        </div>
      </div>

      {/* Mensagens recebidas */}
      <div className="rounded-xl border border-zinc-200 bg-white shadow-sm">
        <div className="flex items-center justify-between border-b border-zinc-100 px-5 py-3">
          <h2 className="font-semibold text-zinc-900">Mensagens recebidas</h2>
          <span className="text-xs text-zinc-400">{messages.length} mensagem{messages.length !== 1 ? "s" : ""}</span>
        </div>

        {messages.length === 0 ? (
          <p className="px-5 py-8 text-center text-sm text-zinc-400">Nenhuma mensagem recebida ainda.</p>
        ) : (
          <ul className="max-h-80 divide-y divide-zinc-100 overflow-y-auto">
            {messages.map((msg) => (
              <li key={msg.id} className="px-5 py-3">
                <div className="flex items-baseline justify-between gap-2">
                  <span className="text-xs font-medium text-zinc-500">{msg.from}</span>
                  <span className="shrink-0 text-xs text-zinc-400">
                    {new Date(msg.timestamp).toLocaleString("pt-BR")}
                  </span>
                </div>
                <p className="mt-1 text-sm text-zinc-800">{msg.body}</p>
              </li>
            ))}
          </ul>
        )}
      </div>

      {showQR && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl">
            <QRCodeDisplay instance={instance} onClose={() => setShowQR(false)} />
          </div>
        </div>
      )}
    </div>
  )
}
