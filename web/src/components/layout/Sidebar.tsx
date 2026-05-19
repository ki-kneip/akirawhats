"use client"

import Link from "next/link"
import { usePathname, useRouter } from "next/navigation"
import { useInstances } from "@/lib/api/instances"
import { useAuthStore } from "@/lib/store/auth"

export function Sidebar() {
  const pathname = usePathname()
  const router = useRouter()
  const { data: instances } = useInstances()
  const connectedCount = instances?.filter((i) => i.status === "connected").length ?? 0
  const { user, clearAuth } = useAuthStore()

  function handleLogout() {
    clearAuth()
    router.push("/login")
  }

  const links = [
    {
      href: "/instances",
      label: "Instâncias",
      icon: (
        <svg className="size-5" fill="none" stroke="currentColor" strokeWidth={1.5} viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 1.5H8.25A2.25 2.25 0 0 0 6 3.75v16.5a2.25 2.25 0 0 0 2.25 2.25h7.5A2.25 2.25 0 0 0 18 20.25V3.75a2.25 2.25 0 0 0-2.25-2.25H13.5m-3 0V3h3V1.5m-3 0h3m-3 8.25h3m-3 3.75h3M6.75 20.25h10.5" />
        </svg>
      ),
    },
    {
      href: "/profile",
      label: "Meu perfil",
      icon: (
        <svg className="size-5" fill="none" stroke="currentColor" strokeWidth={1.5} viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" />
        </svg>
      ),
    },
  ]

  return (
    <aside className="flex h-full w-60 flex-col bg-slate-900 text-slate-200">
      <div className="flex h-16 items-center gap-2 border-b border-slate-700 px-5">
        <span className="text-lg font-bold text-white">Mandy</span>
        <span className="text-xs text-slate-400">Dashboard</span>
      </div>

      <nav className="flex-1 space-y-1 p-3">
        {links.map((link) => {
          const active = pathname.startsWith(link.href)
          return (
            <Link
              key={link.href}
              href={link.href}
              className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                active
                  ? "bg-slate-700 text-white"
                  : "text-slate-300 hover:bg-slate-800 hover:text-white"
              }`}
            >
              {link.icon}
              {link.label}
              {link.href === "/instances" && connectedCount > 0 && (
                <span className="ml-auto rounded-full bg-green-500 px-2 py-0.5 text-xs text-white">
                  {connectedCount}
                </span>
              )}
            </Link>
          )
        })}
      </nav>

      <div className="border-t border-slate-700 p-3">
        <div className="mb-2 px-3">
          <p className="truncate text-sm font-medium text-white">{user?.first_name} {user?.last_name}</p>
          <p className="truncate text-xs text-slate-400">{user?.email}</p>
          {user?.role === "admin" && (
            <span className="mt-1 inline-block rounded bg-amber-500/20 px-1.5 py-0.5 text-xs font-medium text-amber-300">admin</span>
          )}
        </div>
        <button
          onClick={handleLogout}
          className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-slate-300 transition-colors hover:bg-slate-800 hover:text-white"
        >
          <svg className="size-5" fill="none" stroke="currentColor" strokeWidth={1.5} viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15M12 9l-3 3m0 0 3 3m-3-3h12.75" />
          </svg>
          Sair
        </button>
      </div>
    </aside>
  )
}
