"use client"

import { useState } from "react"
import { useCreateInstance } from "@/lib/api/instances"
import { Button } from "@/components/ui/Button"
import { Spinner } from "@/components/ui/Spinner"
import { QRCodeDisplay } from "./QRCodeDisplay"
import type { Instance } from "@/lib/api/types"

interface Props {
  onClose: () => void
}

export function CreateInstanceDialog({ onClose }: Props) {
  const [id, setId] = useState("")
  const [created, setCreated] = useState<Instance | null>(null)
  const { mutate, isPending, error } = useCreateInstance()

  function handleCreate() {
    const trimmed = id.trim()
    if (!trimmed) return
    mutate(trimmed, { onSuccess: (instance) => setCreated(instance) })
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl">
        {!created ? (
          <>
            <h2 className="mb-1 text-lg font-semibold text-zinc-900">Nova instância</h2>
            <p className="mb-4 text-sm text-zinc-500">
              Escolha um ID único para identificar esta conexão WhatsApp.
            </p>

            <input
              type="text"
              placeholder="ex: minha-empresa"
              value={id}
              onChange={(e) => setId(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleCreate()}
              className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-green-500 focus:ring-1 focus:ring-green-500"
              autoFocus
            />

            {error && (
              <p className="mt-2 text-sm text-red-600">{error.message}</p>
            )}

            <div className="mt-4 flex justify-end gap-2">
              <Button variant="ghost" onClick={onClose}>
                Cancelar
              </Button>
              <Button onClick={handleCreate} disabled={isPending || !id.trim()}>
                {isPending && <Spinner className="size-4" />}
                Criar
              </Button>
            </div>
          </>
        ) : (
          <QRCodeDisplay instance={created} onClose={onClose} />
        )}
      </div>
    </div>
  )
}
