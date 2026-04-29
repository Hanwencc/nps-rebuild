import { apiGet } from './request'

export interface TunnelCounts {
  tcp: number
  udp: number
  socks5: number
  httpProxy: number
  secret: number
  p2p: number
}

export interface SystemStats {
  cpu: number
  mem: number
  swap: number
  ioSend: number
  ioRecv: number
}

export interface DashboardSummary {
  version: string
  bridgeType: string
  bridgePort: string
  tlsEnable: boolean
  tlsBridgePort: number
  serverIp: string
  p2pPort: string
  logLevel: string
  ipLimit: string
  flowStoreInterval: string
  httpProxyPort: string
  httpsProxyPort: string

  clientCount: number
  clientOnlineCount: number
  hostCount: number
  tunnelCount: TunnelCounts

  flow: { in: number; out: number }
  connections: number

  system: SystemStats
  load: string
  history: Array<HistorySample>
}

export interface HistorySample {
  cpu?: number
  virtual_mem?: number
  swap_mem?: number
  load1?: number
  load5?: number
  load15?: number
  io_send?: number
  io_recv?: number
  time?: string
  [k: string]: unknown
}

export const dashboardApi = {
  summary(): Promise<DashboardSummary> {
    return apiGet<DashboardSummary>('/dashboard/summary')
  },
}
