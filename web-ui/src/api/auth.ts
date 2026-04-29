import { apiGet, apiPost } from './request'
import type { UserInfo } from './types'

export const authApi = {
  login: (username: string, password: string) =>
    apiPost<UserInfo>('/auth/login', { username, password }),
  logout: () => apiPost<null>('/auth/logout'),
  me: () => apiGet<UserInfo>('/auth/me'),
}
