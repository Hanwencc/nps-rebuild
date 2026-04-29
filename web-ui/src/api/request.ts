import axios, {
  type AxiosInstance,
  type AxiosResponse,
  AxiosHeaders,
} from 'axios'
import type { Envelope } from './types'

/**
 * Single shared axios instance. Two cross-cutting concerns wired here:
 *
 *   • response interceptor unwraps the {code,message,data} envelope and
 *     surfaces non-zero codes as rejected promises so call sites can
 *     simply `try/catch`.
 *
 *   • 401 responses force a redirect to /ui/login (the router will
 *     bring the user back after sign-in via a `redirect` query).
 */
const http: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
  withCredentials: true,
  headers: { 'X-Requested-With': 'XMLHttpRequest' },
})

http.interceptors.response.use(
  (resp: AxiosResponse<Envelope>) => {
    const env = resp.data
    if (env && typeof env === 'object' && 'code' in env) {
      if (env.code === 0) {
        return env.data as unknown as AxiosResponse
      }
      return Promise.reject(new ApiError(env.code, env.message ?? 'error'))
    }
    return resp.data as unknown as AxiosResponse
  },
  (err) => {
    const status = err?.response?.status
    const data = err?.response?.data as Envelope | undefined
    const message = data?.message ?? err?.message ?? 'request failed'
    const url: string = err?.config?.url ?? ''

    // The /auth/me probe is intentionally fired on every navigation
    // to detect existing sessions. A 401 here is expected and must
    // NOT bounce the user to /login (the router itself decides that).
    const isMeProbe = url.endsWith('/auth/me')

    if (status === 401 && !isMeProbe && location.hash.indexOf('#/login') === -1) {
      const back = encodeURIComponent(location.hash.slice(1) || '/')
      location.hash = `#/login?redirect=${back}`
    }
    return Promise.reject(new ApiError(data?.code ?? status ?? -1, message))
  },
)

/** Strongly-typed error so call sites can switch on numeric codes. */
export class ApiError extends Error {
  constructor(
    public code: number,
    public override message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

/* Convenience wrappers — all return the unwrapped `data` payload. */
export async function apiGet<T>(url: string, params?: unknown): Promise<T> {
  return (await http.get(url, { params })) as unknown as T
}
export async function apiPost<T>(url: string, body?: unknown): Promise<T> {
  return (await http.post(url, body, {
    headers: new AxiosHeaders({ 'Content-Type': 'application/json' }),
  })) as unknown as T
}
export async function apiPut<T>(url: string, body?: unknown): Promise<T> {
  return (await http.put(url, body, {
    headers: new AxiosHeaders({ 'Content-Type': 'application/json' }),
  })) as unknown as T
}
export async function apiDel<T>(url: string): Promise<T> {
  return (await http.delete(url)) as unknown as T
}

export default http
