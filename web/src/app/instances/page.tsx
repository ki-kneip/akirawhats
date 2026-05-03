"use client"

import { useState } from "react"
import { useInstances } from "@/lib/api/instances"
import { InstanceCard } from "@/components/instances/InstanceCard"
import { CreateInstanceDialog } from "@/components/instances/CreateInstanceDialog"
import { Button } from "@/components/ui/Button"
import { Spinner } from "@/components/ui/Spinner"

export default function InstancesPage() {
  const { data: instances, isLoading, error } = useInstances()
  const [showCreate, setShowCreate] = useState(false)

  return (
    <>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <p className="text-sm text-zinc-500">
            {instances?.length ?? 0} instância(s) cadastrada(s)
          </p>
        </div>
        <Button onClick={() => setShowCreate(true)}>+ Nova instância</Button>
      </div>

      {isLoading && (
        <div className="flex items-center justify-center py-20">
          <Spinner className="size-8 text-zinc-400" />
        </div>
      )}

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          Erro ao carregar instâncias: {error.message}
        </div>
      )}

      {!isLoading && instances?.length === 0 && (
        <div className="flex flex-col items-center justify-center gap-4 py-20 text-center">
          <div className="flex size-16 items-center justify-center rounded-full bg-zinc-100">
            <svg className="size-8 text-zinc-400" fill="none" stroke="currentColor" strokeWidth={1.5} viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 1.5H8.25A2.25 2.25 0 0 0 6 3.75v16.5a2.25 2.25 0 0 0 2.25 2.25h7.5A2.25 2.25 0 0 0 18 20.25V3.75a2.25 2.25 0 0 0-2.25-2.25H13.5m-3 0V3h3V1.5m-3 0h3" />
            </svg>
          </div>
          <div>
            <p className="font-medium text-zinc-700">Nenhuma instância criada</p>
            <p className="mt-1 text-sm text-zinc-500">Crie a primeira para começar a enviar mensagens.</p>
          </div>
          <Button onClick={() => setShowCreate(true)}>Criar primeira instância</Button>
        </div>
      )}

      {instances && instances.length > 0 && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {instances.map((instance) => (
            <InstanceCard key={instance.id} instance={instance} />
          ))}
        </div>
      )}

      {showCreate && <CreateInstanceDialog onClose={() => setShowCreate(false)} />}
    </>
  )
}
