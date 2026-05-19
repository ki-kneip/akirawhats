"use client"

import { useEffect, useState } from "react"
import { useMe, useUpdateProfile, useChangePassword } from "@/lib/api/user"
import { Button } from "@/components/ui/Button"
import { Spinner } from "@/components/ui/Spinner"

export default function ProfilePage() {
  const { data: me, isLoading } = useMe()

  const { mutate: updateProfile, isPending: savingProfile, isSuccess: profileSaved, error: profileError } = useUpdateProfile()
  const { mutate: changePassword, isPending: savingPassword, isSuccess: passwordSaved, error: passwordError } = useChangePassword()

  const [firstName, setFirstName] = useState("")
  const [lastName, setLastName] = useState("")
  const [email, setEmail] = useState("")

  const [currentPassword, setCurrentPassword] = useState("")
  const [newPassword, setNewPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [confirmError, setConfirmError] = useState<string | null>(null)

  useEffect(() => {
    if (me) {
      setFirstName(me.first_name)
      setLastName(me.last_name)
      setEmail(me.email)
    }
  }, [me])

  function handleSaveProfile(e: React.FormEvent) {
    e.preventDefault()
    updateProfile({ first_name: firstName, last_name: lastName, email })
  }

  function handleChangePassword(e: React.FormEvent) {
    e.preventDefault()
    setConfirmError(null)
    if (newPassword !== confirmPassword) {
      setConfirmError("As senhas não coincidem")
      return
    }
    changePassword(
      { currentPassword, newPassword },
      { onSuccess: () => { setCurrentPassword(""); setNewPassword(""); setConfirmPassword("") } }
    )
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Spinner className="size-8 text-zinc-400" />
      </div>
    )
  }

  return (
    <div className="mx-auto max-w-lg space-y-6">
      {/* Perfil */}
      <div className="rounded-xl border border-zinc-200 bg-white p-6 shadow-sm">
        <h1 className="mb-5 text-lg font-semibold text-zinc-900">Meu perfil</h1>

        <form onSubmit={handleSaveProfile} className="space-y-4">
          <div className="flex gap-3">
            <div className="flex-1">
              <label className="mb-1 block text-sm font-medium text-zinc-700">Nome</label>
              <input
                type="text"
                value={firstName}
                onChange={(e) => setFirstName(e.target.value)}
                required
                className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 focus:ring-1 focus:ring-zinc-500"
              />
            </div>
            <div className="flex-1">
              <label className="mb-1 block text-sm font-medium text-zinc-700">Sobrenome</label>
              <input
                type="text"
                value={lastName}
                onChange={(e) => setLastName(e.target.value)}
                className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 focus:ring-1 focus:ring-zinc-500"
              />
            </div>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-zinc-700">Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 focus:ring-1 focus:ring-zinc-500"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-zinc-700">Cargo</label>
            <input
              type="text"
              value={me?.role ?? ""}
              disabled
              className="w-full rounded-lg border border-zinc-200 bg-zinc-50 px-3 py-2 text-sm text-zinc-400"
            />
          </div>

          {profileError && <p className="text-sm text-red-600">{profileError.message}</p>}
          {profileSaved && <p className="text-sm text-green-600">Perfil salvo!</p>}

          <Button type="submit" disabled={savingProfile}>
            {savingProfile ? "Salvando..." : "Salvar perfil"}
          </Button>
        </form>
      </div>

      {/* Trocar senha */}
      <div className="rounded-xl border border-zinc-200 bg-white p-6 shadow-sm">
        <h2 className="mb-5 text-lg font-semibold text-zinc-900">Trocar senha</h2>

        <form onSubmit={handleChangePassword} className="space-y-4">
          <div>
            <label className="mb-1 block text-sm font-medium text-zinc-700">Senha atual</label>
            <input
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              required
              className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 focus:ring-1 focus:ring-zinc-500"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-zinc-700">Nova senha</label>
            <input
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              required
              minLength={8}
              className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 focus:ring-1 focus:ring-zinc-500"
            />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-zinc-700">Confirmar nova senha</label>
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
              className="w-full rounded-lg border border-zinc-300 px-3 py-2 text-sm outline-none focus:border-zinc-500 focus:ring-1 focus:ring-zinc-500"
            />
          </div>

          {(confirmError || passwordError) && (
            <p className="text-sm text-red-600">{confirmError ?? passwordError?.message}</p>
          )}
          {passwordSaved && <p className="text-sm text-green-600">Senha alterada com sucesso!</p>}

          <Button type="submit" disabled={savingPassword}>
            {savingPassword ? "Salvando..." : "Trocar senha"}
          </Button>
        </form>
      </div>
    </div>
  )
}
