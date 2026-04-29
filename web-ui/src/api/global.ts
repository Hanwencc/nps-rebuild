import { apiGet, apiPut } from './request'

export interface GlobalConfig {
  blackIpList: string[]
  serverUrl: string
}

export const globalApi = {
  get(): Promise<GlobalConfig> {
    return apiGet<GlobalConfig>('/global')
  },
  update(payload: GlobalConfig): Promise<void> {
    return apiPut<void>('/global', payload)
  },
}
