"use client"

"use client"

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { api } from "./client"
import type { Instance, Message, SendTextResponse, GroupInfo } from "./types"

export const instanceKeys = {
  all: ["instances"] as const,
  detail: (id: string) => ["instances", id] as const,
  qr: (id: string) => ["instances", id, "qr"] as const,
  messages: (id: string) => ["instances", id, "messages"] as const,
}

export function useInstances() {
  return useQuery({
    queryKey: instanceKeys.all,
    queryFn: () => api.get<Instance[]>("/api/instance"),
    refetchInterval: 5000,
  })
}

export function useInstance(id: string) {
  return useQuery({
    queryKey: instanceKeys.detail(id),
    queryFn: () => api.get<Instance>(`/api/instance/${id}`),
    refetchInterval: 3000,
  })
}

export function useInstanceQR(id: string, enabled: boolean) {
  return useQuery({
    queryKey: instanceKeys.qr(id),
    queryFn: () => api.get<{ qr: string }>(`/api/instance/${id}/qr`),
    enabled,
    refetchInterval: 3000,
  })
}

export function useCreateInstance() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.post<Instance>("/api/instance", { id }),
    onSuccess: () => qc.invalidateQueries({ queryKey: instanceKeys.all }),
  })
}

export function useDeleteInstance() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.delete(`/api/instance/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: instanceKeys.all }),
  })
}

export function useSetWebhook(instanceId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (url: string) =>
      api.post(`/api/instance/${instanceId}/webhook`, { url }),
    onSuccess: () => qc.invalidateQueries({ queryKey: instanceKeys.detail(instanceId) }),
  })
}

export function useSendText(instanceId: string) {
  return useMutation({
    mutationFn: ({ to, message }: { to: string; message: string }) =>
      api.post<SendTextResponse>(`/api/instance/${instanceId}/send/text`, {
        to,
        message,
      }),
  })
}

export function useSendImage(instanceId: string) {
  return useMutation({
    mutationFn: ({ to, file, caption }: { to: string; file: File; caption?: string }) => {
      const form = new FormData()
      form.append("to", to)
      form.append("file", file)
      if (caption) form.append("caption", caption)
      return api.upload<SendTextResponse>(`/api/instance/${instanceId}/send/image`, form)
    },
  })
}

export function useInstanceMessages(instanceId: string) {
  return useQuery({
    queryKey: instanceKeys.messages(instanceId),
    queryFn: () => api.get<Message[]>(`/api/instance/${instanceId}/messages`),
    refetchInterval: 5000,
  })
}

export function useGroups(instanceId: string, enabled: boolean) {
  return useQuery({
    queryKey: ["instances", instanceId, "groups"],
    queryFn: () => api.get<GroupInfo[]>(`/api/instance/${instanceId}/groups`),
    enabled,
    staleTime: 30_000,
  })
}
