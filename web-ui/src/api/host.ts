import { apiDel, apiGet, apiPost, apiPut } from './request'
import type { Page } from './types'
import type { Target } from './tunnel'

export interface Host {
  Id: number
  Host: string
  HeaderChange: string
  HostChange: string
  Location: string
  Remark: string
  Scheme: string
  CertFilePath: string
  KeyFilePath: string
  NoStore: boolean
  IsClose: boolean
  AutoHttps: boolean
  Flow: { ExportFlow: number; InletFlow: number; FlowLimit: number }
  Client: { Id: number; VerifyKey: string; Remark: string }
  Target: Target
}

export interface HostPayload {
  clientId: number
  host: string
  target?: string
  localProxy?: boolean
  header?: string
  hostchange?: string
  remark?: string
  location?: string
  scheme?: string
  keyFilePath?: string
  certFilePath?: string
  autoHttps?: boolean
}

export interface HostListParams {
  clientId?: number
  offset?: number
  limit?: number
  search?: string
}

export const hostApi = {
  list: (params: HostListParams) => apiGet<Page<Host>>('/hosts', params),
  get: (id: number) => apiGet<Host>(`/hosts/${id}`),
  create: (body: HostPayload) => apiPost<{ id: number }>('/hosts', body),
  update: (id: number, body: HostPayload) =>
    apiPut<null>(`/hosts/${id}`, body),
  remove: (id: number) => apiDel<null>(`/hosts/${id}`),
  changeStatus: (id: number, status: boolean) =>
    apiPost<null>(`/hosts/${id}/status`, { status }),
}
