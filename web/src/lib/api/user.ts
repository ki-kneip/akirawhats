"use client"

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { api } from "./client"
import { useAuthStore } from "../store/auth"

interface ProfileDTO {
  id: string
  first_name: string
  last_name: string
  email: string
  role: string
}

interface UpdateProfileBody {
  first_name?: string
  last_name?: string
  email?: string
}

export function useMe() {
  return useQuery({
    queryKey: ["me"],
    queryFn: () => api.get<ProfileDTO>("/api/user/me"),
  })
}

export function useUpdateProfile() {
  const qc = useQueryClient()
  const setAuth = useAuthStore((s) => s.setAuth)
  const token = useAuthStore((s) => s.token)
  const user = useAuthStore((s) => s.user)
  return useMutation({
    mutationFn: (body: UpdateProfileBody) => api.put<ProfileDTO>("/api/user/me", body),
    onSuccess: (updated) => {
      qc.setQueryData(["me"], updated)
      if (token && user) {
        setAuth(token, { ...user, ...updated })
      }
    },
  })
}

export function useChangePassword() {
  return useMutation({
    mutationFn: ({ currentPassword, newPassword }: { currentPassword: string; newPassword: string }) =>
      api.put("/api/user/me/password", { current_password: currentPassword, new_password: newPassword }),
  })
}
