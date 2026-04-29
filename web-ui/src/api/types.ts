/* Common API types shared across the SPA. */

export interface Envelope<T = unknown> {
  code: number
  message: string
  data?: T
}

export interface Page<T> {
  total: number
  items: T[]
}

export interface UserInfo {
  isAdmin: boolean
  username: string
  clientId: number
  authed: boolean
  webVersion?: string
}

export interface FlowInfo {
  ExportFlow: number
  InletFlow: number
  FlowLimit: number
}

export interface ConfigInfo {
  U: string
  P: string
  Compress: boolean
  Crypt: boolean
}

export interface RateInfo {
  NowRate: number
}

export interface Client {
  Id: number
  VerifyKey: string
  Addr: string
  Remark: string
  Status: boolean
  IsConnect: boolean
  IsTls: boolean
  RateLimit: number
  Flow: FlowInfo
  Rate: RateInfo
  Cnf: ConfigInfo
  NoStore: boolean
  NoDisplay: boolean
  MaxConn: number
  NowConn: number
  WebUserName: string
  WebPassword: string
  ConfigConnAllow: boolean
  MaxTunnelNum: number
  Version: string
  BlackIpList: string[]
  CreateTime: string
  LastOnlineTime: string
  IpWhite: boolean
  IpWhitePass: string
  IpWhiteList: string[]
}

export interface ClientPayload {
  vkey?: string
  remark?: string
  u?: string
  p?: string
  compress?: boolean
  crypt?: boolean
  configConnAllow?: boolean
  rateLimit?: number
  maxConn?: number
  maxTunnel?: number
  flowLimit?: number
  webUsername?: string
  webPassword?: string
  blackIpList?: string[]
  ipWhite?: boolean
  ipWhitePass?: string
  ipWhiteList?: string[]
}

export interface QuickInfo {
  id: number
  vkey: string
  remark: string
  ip: string
  bridgePort: number
  bridgeType: string
  tlsPort: number
  /** SHA-256 hex of the bridge cert (colon-separated lowercase). Empty
   * when nps was not restarted after the fingerprint feature shipped. */
  tlsFingerprint?: string
}
