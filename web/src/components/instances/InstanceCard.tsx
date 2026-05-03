"use client"

import Link from "next/link"
import { useDeleteInstance } from "@/lib/api/instances"
import { StatusBadge } from "@/components/ui/Badge"
import { Button } from "@/components/ui/Button"
import type { Instance } from "@/lib/api/types"

export function InstanceCard({ instance }: { instance: Instance }) {
  const { mutate: deleteInstance, isPending } = useDeleteInstance()

  return (
    <div className="flex flex-col gap-3 rounded-xl border border-zinc-200 bg-white p-5 shadow-sm transition-shadow hover:shadow-md">
      <div className="flex items-start justify-between gap-2">
        <div>
          <p className="font-semibold text-zinc-900">{instance.id}</p>
          {instance.phone && (
            <p className="mt-0.5 text-sm text-zinc-500">{instance.phone}</p>
          )}
        </div>
        <StatusBadge status={instance.status} />
      </div>

      <div className="flex gap-2">
        <Link
          href={`/instances/${instance.id}`}
          className="flex-1 rounded-lg border border-zinc-300 py-1.5 text-center text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50"
        >
          Abrir
        </Link>
        <Button
          variant="danger"
          size="sm"
          disabled={isPending}
          onClick={() => {
            if (confirm(`Remover instância "${instance.id}"?`)) {
              deleteInstance(instance.id)
            }
          }}
        >
          Remover
        </Button>
      </div>
    </div>
  )
}
