import { apiDel, apiGet, apiPost, apiPut } from './request'
import type { Client, ClientPayload, Page, QuickInfo } from './types'

export interface ListParams {
  offset?: number
  limit?: number
  search?: string
  sort?: string
  order?: 'asc' | 'desc' | ''
}

export const clientApi = {
  list: (params: ListParams) => apiGet<Page<Client>>('/clients', params),
  get: (id: number) => apiGet<Client>(`/clients/${id}`),
  create: (body: ClientPayload) => apiPost<{ id: number }>('/clients', body),
  update: (id: number, body: ClientPayload) =>
    apiPut<null>(`/clients/${id}`, body),
  remove: (id: number) => apiDel<null>(`/clients/${id}`),
  changeStatus: (id: number, status: boolean) =>
    apiPost<null>(`/clients/${id}/status`, { status }),
  quickInfo: (id: number) => apiGet<QuickInfo>(`/clients/${id}/quickinfo`),
}
