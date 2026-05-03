"use client"

import { useEffect, useRef, useState } from "react"
import QRCode from "qrcode"
import { useInstance, useInstanceQR } from "@/lib/api/instances"
import { Spinner } from "@/components/ui/Spinner"
import { Button } from "@/components/ui/Button"
import type { Instance } from "@/lib/api/types"

interface Props {
  instance: Instance
  onClose: () => void
}

export function QRCodeDisplay({ instance, onClose }: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [connected, setConnected] = useState(false)

  const { data: status } = useInstance(instance.id)
  const isWaitingForQR = status?.status === "qr" || status?.status === "connecting"
  const { data: qrData } = useInstanceQR(instance.id, isWaitingForQR)

  useEffect(() => {
    if (status?.status === "connected") setConnected(true)
  }, [status?.status])

  useEffect(() => {
    const qrString = qrData?.qr ?? instance.qr
    if (qrString && canvasRef.current) {
      QRCode.toCanvas(canvasRef.current, qrString, { width: 256 })
    }
  }, [qrData?.qr, instance.qr])

  if (connected) {
    return (
      <div className="flex flex-col items-center gap-4 py-4">
        <div className="flex size-16 items-center justify-center rounded-full bg-green-100">
          <svg className="size-8 text-green-600" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
          </svg>
        </div>
        <div className="text-center">
          <p className="font-semibold text-zinc-900">Conectado!</p>
          <p className="text-sm text-zinc-500">{status?.phone}</p>
        </div>
        <Button onClick={onClose}>Fechar</Button>
      </div>
    )
  }

  return (
    <div className="flex flex-col items-center gap-4">
      <h2 className="text-lg font-semibold text-zinc-900">Escanear QR Code</h2>
      <p className="text-center text-sm text-zinc-500">
        Abra o WhatsApp no celular e escaneie o código abaixo.
      </p>

      <div className="flex size-64 items-center justify-center rounded-xl border border-zinc-200 bg-zinc-50">
        {instance.qr || qrData?.qr ? (
          <canvas ref={canvasRef} />
        ) : (
          <Spinner className="size-8 text-zinc-400" />
        )}
      </div>

      <p className="flex items-center gap-1.5 text-xs text-zinc-400">
        <Spinner className="size-3" />
        Aguardando conexão...
      </p>

      <Button variant="ghost" size="sm" onClick={onClose}>
        Cancelar
      </Button>
    </div>
  )
}
