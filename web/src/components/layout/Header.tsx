"use client"

import { useQuery } from "@tanstack/react-query"
import { api } from "@/lib/api/client"

function ApiStatus() {
  const { isSuccess, isError } = useQuery({
    queryKey: ["ping"],
    queryFn: () => api.get<{ message: string }>("/api/ping"),
    refetchInterval: 15000,
  })

  if (!isSuccess && !isError) return null

  return (
    <span className="flex items-center gap-1.5 text-xs text-zinc-500">
      <span
        className={`size-2 rounded-full ${isSuccess ? "bg-green-500" : "bg-red-500"}`}
      />
      API {isSuccess ? "online" : "offline"}
    </span>
  )
}

export function Header({ title }: { title: string }) {
  return (
    <header className="flex h-16 shrink-0 items-center justify-between border-b border-zinc-200 bg-white px-6">
      <h1 className="text-base font-semibold text-zinc-900">{title}</h1>
      <ApiStatus />
    </header>
  )
}
