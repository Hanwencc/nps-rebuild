import { apiDel, apiGet, apiPost, apiPut } from './request'

export interface ApiToken {
  id: number
  keyId: string
  remark: string
  allowedPathPrefix: string
  allowedMethods: string[]
  allowIps: string[]
  expiresAt: number // unix seconds; 0 = never
  createdAt: number
  lastUsedAt: number
  lastUsedIp: string
  disabled: boolean
}

export interface TokenWriteRequest {
  remark?: string
  allowedPathPrefix?: string
  allowedMethods?: string[]
  allowIps?: string[]
  expiresAt?: number
  disabled?: boolean
}

export interface TokenSecretReveal {
  token: ApiToken
  /** plaintext secret — only returned by create / rotate, ONCE. */
  secret: string
}

export const tokenApi = {
  list(): Promise<ApiToken[]> {
    return apiGet<ApiToken[]>('/tokens')
  },
  get(id: number): Promise<ApiToken> {
    return apiGet<ApiToken>(`/tokens/${id}`)
  },
  create(payload: TokenWriteRequest): Promise<TokenSecretReveal> {
    return apiPost<TokenSecretReveal>('/tokens', payload)
  },
  update(id: number, payload: TokenWriteRequest): Promise<ApiToken> {
    return apiPut<ApiToken>(`/tokens/${id}`, payload)
  },
  remove(id: number): Promise<void> {
    return apiDel<void>(`/tokens/${id}`)
  },
  rotate(id: number): Promise<TokenSecretReveal> {
    return apiPost<TokenSecretReveal>(`/tokens/${id}/rotate`, {})
  },
}
