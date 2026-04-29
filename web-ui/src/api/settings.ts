import { apiGet, apiPut } from './request'

export interface SettingItem {
  key: string
  label: string
  group: string
  type: 'string' | 'int' | 'bool' | 'enum' | 'password'
  enum?: string[]
  help?: string
  needsRestart: boolean
  bootstrap: boolean
  value: string
}

export interface SettingUpdateResult {
  applied: number
  rejected: string[] | null
}

export const settingsApi = {
  list(): Promise<SettingItem[]> {
    return apiGet<SettingItem[]>('/settings')
  },
  update(payload: Record<string, string>): Promise<SettingUpdateResult> {
    return apiPut<SettingUpdateResult>('/settings', payload)
  },
}
