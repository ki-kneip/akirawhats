"use client"

import { useEffect } from "react"
import { useRouter } from "next/navigation"
import { useAuthStore } from "@/lib/store/auth"
import { Sidebar } from "@/components/layout/Sidebar"
import { Header } from "@/components/layout/Header"

export default function ProfileLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const router = useRouter()
  const token = useAuthStore((state) => state.token)

  useEffect(() => {
    if (token === null) {
      router.replace("/login")
    }
  }, [token, router])

  if (token === null) return null

  return (
    <div className="flex h-full">
      <Sidebar />
      <div className="flex flex-1 flex-col overflow-hidden">
        <Header title="Meu perfil" />
        <main className="flex-1 overflow-auto bg-zinc-50 p-6">{children}</main>
      </div>
    </div>
  )
}
