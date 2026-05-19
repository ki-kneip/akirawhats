"use client"

import { use, useEffect, useRef, useState } from "react"
import { useInstance, useSendText, useSendImage, useSetWebhook, useDeleteInstance, useInstanceMessages, useGroups } from "@/lib/api/instances"
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
  const { mutate: sendImage, isPending: sendingImage, data: sentImage, error: sendImageError } = useSendImage(id)
  const { mutate: setWebhook, isPending: savingWebhook, isSuccess: webhookSaved } = useSetWebhook(id)
  const { data: messages = [] } = useInstanceMessages(id)
  const { data: groups = [] } = useGroups(id, instance?.status === "connected")

  const [to, setTo] = useState("")
  const [message, setMessage] = useState("")
  const [imageTo, setImageTo] = useState("")
  const [imageCaption, setImageCaption] = useState("")
  const [imageFile, setImageFile] = useState<File | null>(null)
  const [webhookUrl, setWebhookUrl] = useState("")
  const [showQR, setShowQR] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (instance?.webhookUrl) setWebhookUrl(instance.webhookUrl)
  }, [instance?.webhookUrl])

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

  function handleSendImage() {
    if (!imageTo.trim() || !imageFile) return
    sendImage(
      { to: imageTo.trim(), file: imageFile, caption: imageCaption.trim() || undefined },
      { onSuccess: () => { setImageFile(null); setImageCaption(""); if (fileInputRef.current) fileInputRef.current.value = "" } }
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
            <label className="mb-1 block text-xs font-medium text-zinc-600">Para (número com DDI ou grupo)</label>
            <input
              type="text"
              placeholder="5511999999999"
              value={to}
              onChange={(e) => setTo(e.target.value)}
              disabled={!isConnected}
              className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-green-500 focus:ring-1 focus:ring-green-500 disabled:bg-zinc-50 disabled:text-zinc-400"
            />
            {groups.length > 0 && (
              <select
                onChange={(e) => { if (e.target.value) setTo(e.target.value) }}
                defaultValue=""
                disabled={!isConnected}
                className="mt-1 w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm text-zinc-600 outline-none focus:border-green-500 focus:ring-1 focus:ring-green-500 disabled:opacity-50"
              >
                <option value="">Selecionar grupo...</option>
                {groups.map((g) => (
                  <option key={g.jid} value={g.jid}>{g.name}</option>
                ))}
              </select>
            )}
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

      {/* Enviar imagem */}
      <div className="rounded-xl border border-zinc-200 bg-white p-5 shadow-sm">
        <h2 className="mb-4 font-semibold text-zinc-900">Enviar imagem</h2>

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
              value={imageTo}
              onChange={(e) => setImageTo(e.target.value)}
              disabled={!isConnected}
              className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-green-500 focus:ring-1 focus:ring-green-500 disabled:bg-zinc-50 disabled:text-zinc-400"
            />
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium text-zinc-600">Imagem (max 5 MB)</label>
            <input
              ref={fileInputRef}
              type="file"
              accept="image/*"
              disabled={!isConnected}
              onChange={(e) => setImageFile(e.target.files?.[0] ?? null)}
              className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm text-zinc-600 file:mr-3 file:rounded file:border-0 file:bg-zinc-100 file:px-2 file:py-1 file:text-xs disabled:opacity-50"
            />
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium text-zinc-600">Legenda (opcional)</label>
            <input
              type="text"
              placeholder="Legenda da imagem..."
              value={imageCaption}
              onChange={(e) => setImageCaption(e.target.value)}
              disabled={!isConnected}
              className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-green-500 focus:ring-1 focus:ring-green-500 disabled:bg-zinc-50 disabled:text-zinc-400"
            />
          </div>

          {sendImageError && <p className="text-sm text-red-600">{sendImageError.message}</p>}
          {sentImage && <p className="text-sm text-green-600">Enviado! ID: {sentImage.id}</p>}

          <Button
            onClick={handleSendImage}
            disabled={!isConnected || sendingImage || !imageTo.trim() || !imageFile}
          >
            {sendingImage && <Spinner className="size-4" />}
            Enviar imagem
          </Button>
        </div>
      </div>

      {/* Webhook */}
      <div className="rounded-xl border border-zinc-200 bg-white p-5 shadow-sm">
        <h2 className="mb-4 font-semibold text-zinc-900">Webhook</h2>
        <div className="flex gap-2">
          <input
            type="url"
            placeholder="https://seu-servidor.com/webhook"
            value={webhookUrl}
            onChange={(e) => setWebhookUrl(e.target.value)}
            className="flex-1 rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 focus:ring-1 focus:ring-zinc-500"
          />
          <Button
            onClick={() => setWebhook(webhookUrl)}
            disabled={savingWebhook || !webhookUrl.trim()}
          >
            {savingWebhook ? "Salvando..." : "Salvar"}
          </Button>
        </div>
        {webhookSaved && <p className="mt-2 text-sm text-green-600">Webhook salvo!</p>}
        <p className="mt-2 text-xs text-zinc-400">
          Mensagens recebidas serão enviadas via POST para esta URL.
        </p>
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
              <li key={msg.id} className={`px-5 py-3 ${msg.direction === "out" ? "bg-green-50/50" : ""}`}>
                <div className="flex items-baseline justify-between gap-2">
                  <span className="text-xs font-medium text-zinc-500">
                    {msg.direction === "out" ? "Você → " : ""}{msg.from}
                  </span>
                  <div className="flex shrink-0 items-center gap-1.5">
                    {msg.direction === "out" && (
                      <span className={`text-xs ${
                        msg.status === "read" ? "text-blue-500" :
                        msg.status === "delivered" ? "text-green-500" :
                        "text-zinc-400"
                      }`}>
                        {msg.status === "read" ? "✓✓ lido" : msg.status === "delivered" ? "✓✓ entregue" : "✓ enviado"}
                      </span>
                    )}
                    <span className="text-xs text-zinc-400">
                      {new Date(msg.timestamp).toLocaleString("pt-BR")}
                    </span>
                  </div>
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
