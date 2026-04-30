import { apiGet } from './request'

export interface Socks5GatewayStatus {
  listening: boolean
  addr: string
  port: number
  routes: number
}

export const socks5Api = {
  gateway: () => apiGet<Socks5GatewayStatus>('/socks5/gateway'),
}
