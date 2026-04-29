import { apiDel, apiGet, apiPost, apiPut } from './request'
import type { Page } from './types'

export interface Target {
  TargetStr: string
  LocalProxy?: boolean
}

export interface Tunnel {
  Id: number
  Port: number
  ServerIp: string
  Mode: string
  Status: boolean
  RunStatus: boolean
  Client: { Id: number; VerifyKey: string; Remark: string }
  Ports: string
  Flow: { ExportFlow: number; InletFlow: number; FlowLimit: number }
  Password: string
  Remark: string
  TargetAddr: string
  NoStore: boolean
  LocalPath: string
  StripPre: string
  ProtoVersion: string
  Target: Target
}

export interface TunnelPayload {
  clientId: number
  mode: string
  port?: number
  serverIp?: string
  target?: string
  localProxy?: boolean
  password?: string
  remark?: string
  localPath?: string
  stripPre?: string
  protoVersion?: string
}

export interface TunnelListParams {
  mode?: string
  clientId?: number
  offset?: number
  limit?: number
  search?: string
  sort?: string
  order?: 'asc' | 'desc' | ''
}

export const tunnelApi = {
  list: (params: TunnelListParams) =>
    apiGet<Page<Tunnel>>('/tunnels', params),
  get: (id: number) => apiGet<Tunnel>(`/tunnels/${id}`),
  create: (body: TunnelPayload) =>
    apiPost<{ id: number }>('/tunnels', body),
  update: (id: number, body: TunnelPayload) =>
    apiPut<null>(`/tunnels/${id}`, body),
  remove: (id: number) => apiDel<null>(`/tunnels/${id}`),
  start: (id: number) => apiPost<null>(`/tunnels/${id}/start`),
  stop: (id: number) => apiPost<null>(`/tunnels/${id}/stop`),
  copy: (id: number) => apiPost<{ id: number }>(`/tunnels/${id}/copy`),
}
