<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, computed } from 'vue'
import {
  NCard,
  NGrid,
  NGridItem,
  NStatistic,
  NSpin,
  NTag,
  NSpace,
  NAlert,
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { dashboardApi, type DashboardSummary } from '@/api/dashboard'
import { useAuthStore } from '@/stores/auth'
import Sparkline from '@/components/Sparkline.vue'

const { t } = useI18n()
const auth = useAuthStore()

const summary = ref<DashboardSummary | null>(null)
const loading = ref(false)
const error = ref('')
let timer: number | null = null

// Rolling buffers for charts. Seed once from server-provided
// `history` (10 sampled points) on first load, then push every 5s
// poll on top so the line scrolls.
const WINDOW = 60
const cpuSeries = ref<number[]>([])
const memSeries = ref<number[]>([])
const swapSeries = ref<number[]>([])
const load1Series = ref<number[]>([])
const inRateSeries = ref<number[]>([]) // bytes/s
const outRateSeries = ref<number[]>([])

let lastFlowIn = -1
let lastFlowOut = -1
let lastSampleAt = 0
let seeded = false

function pushCapped(arr: number[], v: number) {
  arr.push(v)
  if (arr.length > WINDOW) arr.shift()
}

function seedFromHistory(s: DashboardSummary) {
  if (seeded || !s.history || s.history.length === 0) return
  for (const h of s.history) {
    cpuSeries.value.push(Number(h.cpu) || 0)
    memSeries.value.push(Number(h.virtual_mem) || 0)
    swapSeries.value.push(Number(h.swap_mem) || 0)
    load1Series.value.push(Number(h.load1) || 0)
  }
  seeded = true
}

function ingest(s: DashboardSummary) {
  seedFromHistory(s)
  pushCapped(cpuSeries.value, s.system.cpu || 0)
  pushCapped(memSeries.value, s.system.mem || 0)
  pushCapped(swapSeries.value, s.system.swap || 0)
  let l1 = 0
  try {
    const parsed = JSON.parse(s.load || '{}')
    l1 = Number(parsed.load1) || 0
  } catch {
    /* ignore */
  }
  pushCapped(load1Series.value, l1)

  // Flow rate (bytes per second) computed from delta
  const now = Date.now()
  const dt = lastSampleAt ? (now - lastSampleAt) / 1000 : 0
  if (lastFlowIn >= 0 && dt > 0) {
    const dIn = Math.max(0, (s.flow.in || 0) - lastFlowIn)
    const dOut = Math.max(0, (s.flow.out || 0) - lastFlowOut)
    pushCapped(inRateSeries.value, dIn / dt)
    pushCapped(outRateSeries.value, dOut / dt)
  } else {
    pushCapped(inRateSeries.value, 0)
    pushCapped(outRateSeries.value, 0)
  }
  lastFlowIn = s.flow.in || 0
  lastFlowOut = s.flow.out || 0
  lastSampleAt = now
}

async function load() {
  loading.value = true
  try {
    summary.value = await dashboardApi.summary()
    if (summary.value) ingest(summary.value)
    error.value = ''
  } catch (e: unknown) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

function fmtBytes(n: number): string {
  if (!n || n < 0) return '0 B'
  const u = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  let v = n
  while (v >= 1024 && i < u.length - 1) {
    v /= 1024
    i++
  }
  return `${v.toFixed(2)} ${u[i]}`
}

function fmtRate(n: number): string {
  return fmtBytes(n) + '/s'
}

const tunnelTotal = computed(() => {
  const c = summary.value?.tunnelCount
  if (!c) return 0
  return c.tcp + c.udp + c.socks5 + c.httpProxy + c.secret + c.p2p
})

const cpuNow = computed(() => Math.round(summary.value?.system.cpu ?? 0))
const memNow = computed(() => Math.round(summary.value?.system.mem ?? 0))
const swapNow = computed(() => Math.round(summary.value?.system.swap ?? 0))

const loadParsed = computed(() => {
  try {
    return JSON.parse(summary.value?.load || '{}') as {
      load1?: number
      load5?: number
      load15?: number
    }
  } catch {
    return {} as { load1?: number; load5?: number; load15?: number }
  }
})

const inRateNow = computed(
  () => inRateSeries.value[inRateSeries.value.length - 1] ?? 0,
)
const outRateNow = computed(
  () => outRateSeries.value[outRateSeries.value.length - 1] ?? 0,
)

onMounted(() => {
  if (auth.isAdmin) {
    void load()
    timer = window.setInterval(load, 5000)
  }
})

onBeforeUnmount(() => {
  if (timer !== null) window.clearInterval(timer)
})
</script>

<template>
  <div class="space-y-4">
    <NAlert v-if="!auth.isAdmin" type="info">
      {{ t('dashboard.userOnly') }}
    </NAlert>

    <template v-else>
      <NSpin :show="loading && !summary">
        <NGrid :cols="4" :x-gap="16" :y-gap="16" responsive="screen" item-responsive>
          <!-- top KPI row -->
          <NGridItem span="1 m:1">
            <NCard class="nps-stat-card">
              <NStatistic :label="t('dashboard.clients')">
                <span>{{ summary?.clientOnlineCount ?? 0 }}</span>
                <template #suffix> / {{ summary?.clientCount ?? 0 }}</template>
              </NStatistic>
            </NCard>
          </NGridItem>
          <NGridItem span="1 m:1">
            <NCard class="nps-stat-card">
              <NStatistic :label="t('dashboard.hosts')" :value="summary?.hostCount ?? 0" />
            </NCard>
          </NGridItem>
          <NGridItem span="1 m:1">
            <NCard class="nps-stat-card">
              <NStatistic :label="t('dashboard.tunnels')" :value="tunnelTotal" />
            </NCard>
          </NGridItem>
          <NGridItem span="1 m:1">
            <NCard class="nps-stat-card">
              <NStatistic :label="t('dashboard.connections')" :value="summary?.connections ?? 0" />
            </NCard>
          </NGridItem>

          <!-- tunnel distribution -->
          <NGridItem :span="2">
            <NCard :title="t('dashboard.tunnelByMode')">
              <NSpace>
                <NTag type="info">TCP {{ summary?.tunnelCount.tcp ?? 0 }}</NTag>
                <NTag type="info">UDP {{ summary?.tunnelCount.udp ?? 0 }}</NTag>
                <NTag type="info">SOCKS5 {{ summary?.tunnelCount.socks5 ?? 0 }}</NTag>
                <NTag type="info">HTTP {{ summary?.tunnelCount.httpProxy ?? 0 }}</NTag>
                <NTag type="success">SECRET {{ summary?.tunnelCount.secret ?? 0 }}</NTag>
                <NTag type="success">P2P {{ summary?.tunnelCount.p2p ?? 0 }}</NTag>
              </NSpace>
            </NCard>
          </NGridItem>

          <!-- flow trend -->
          <NGridItem :span="2">
            <NCard :title="t('dashboard.flow')">
              <div class="grid grid-cols-2 gap-4">
                <div class="nps-metric">
                  <div class="nps-metric-label">
                    <span class="nps-dot" style="background:#22c55e" />
                    {{ t('dashboard.inFlow') }}
                  </div>
                  <div class="nps-metric-value nps-mono">
                    {{ fmtRate(inRateNow) }}
                  </div>
                  <div class="nps-metric-sub nps-mono opacity-60">
                    Σ {{ fmtBytes(summary?.flow.in ?? 0) }}
                  </div>
                  <Sparkline
                    :data="inRateSeries"
                    color="#22c55e"
                    gradient="#22c55e"
                    :height="46"
                  />
                </div>
                <div class="nps-metric">
                  <div class="nps-metric-label">
                    <span class="nps-dot" style="background:#06b6d4" />
                    {{ t('dashboard.outFlow') }}
                  </div>
                  <div class="nps-metric-value nps-mono">
                    {{ fmtRate(outRateNow) }}
                  </div>
                  <div class="nps-metric-sub nps-mono opacity-60">
                    Σ {{ fmtBytes(summary?.flow.out ?? 0) }}
                  </div>
                  <Sparkline
                    :data="outRateSeries"
                    color="#06b6d4"
                    gradient="#06b6d4"
                    :height="46"
                  />
                </div>
              </div>
            </NCard>
          </NGridItem>

          <!-- system resources with sparklines -->
          <NGridItem :span="2">
            <NCard :title="t('dashboard.system')">
              <div class="space-y-3">
                <div class="nps-row">
                  <div class="flex items-center justify-between">
                    <span class="text-sm opacity-70">CPU</span>
                    <span class="nps-mono text-sm">{{ cpuNow }}%</span>
                  </div>
                  <Sparkline
                    :data="cpuSeries"
                    :min="0"
                    :max="100"
                    color="#6366f1"
                    gradient="#6366f1"
                    :height="40"
                  />
                </div>
                <div class="nps-row">
                  <div class="flex items-center justify-between">
                    <span class="text-sm opacity-70">MEM</span>
                    <span class="nps-mono text-sm">{{ memNow }}%</span>
                  </div>
                  <Sparkline
                    :data="memSeries"
                    :min="0"
                    :max="100"
                    color="#06b6d4"
                    gradient="#06b6d4"
                    :height="40"
                  />
                </div>
                <div class="nps-row">
                  <div class="flex items-center justify-between">
                    <span class="text-sm opacity-70">SWAP</span>
                    <span class="nps-mono text-sm">{{ swapNow }}%</span>
                  </div>
                  <Sparkline
                    :data="swapSeries"
                    :min="0"
                    :max="100"
                    color="#a855f7"
                    gradient="#a855f7"
                    :height="40"
                  />
                </div>
                <div class="nps-row">
                  <div class="flex items-center justify-between">
                    <span class="text-sm opacity-70">{{ t('dashboard.load') }}</span>
                    <span class="nps-mono text-sm">
                      {{ (loadParsed.load1 ?? 0).toFixed(2) }} ·
                      {{ (loadParsed.load5 ?? 0).toFixed(2) }} ·
                      {{ (loadParsed.load15 ?? 0).toFixed(2) }}
                    </span>
                  </div>
                  <Sparkline
                    :data="load1Series"
                    color="#f59e0b"
                    gradient="#f59e0b"
                    :height="40"
                  />
                </div>
              </div>
            </NCard>
          </NGridItem>

          <!-- server info -->
          <NGridItem :span="2">
            <NCard :title="t('dashboard.serverInfo')">
              <div class="nps-info-grid">
                <div class="nps-info-row">
                  <span class="nps-info-label">{{ t('dashboard.version') }}</span>
                  <span class="nps-info-value nps-mono">{{ summary?.version || '-' }}</span>
                </div>
                <div class="nps-info-row">
                  <span class="nps-info-label">Bridge</span>
                  <span class="nps-info-value">
                    <NTag size="small" type="info" round>
                      {{ (summary?.bridgeType || 'tcp').toUpperCase() }}
                    </NTag>
                    <span class="nps-mono">:{{ summary?.bridgePort || '-' }}</span>
                  </span>
                </div>
                <div class="nps-info-row">
                  <span class="nps-info-label">TLS Bridge</span>
                  <span class="nps-info-value">
                    <NTag
                      :type="summary?.tlsEnable ? 'success' : 'default'"
                      size="small"
                      round
                    >
                      {{ summary?.tlsEnable ? t('dashboard.tlsOn') : t('dashboard.tlsOff') }}
                    </NTag>
                    <span v-if="summary?.tlsEnable" class="nps-mono">
                      :{{ summary?.tlsBridgePort }}
                    </span>
                  </span>
                </div>
                <div class="nps-info-row">
                  <span class="nps-info-label">HTTP Proxy</span>
                  <span class="nps-info-value nps-mono">:{{ summary?.httpProxyPort || '-' }}</span>
                </div>
                <div class="nps-info-row">
                  <span class="nps-info-label">HTTPS Proxy</span>
                  <span class="nps-info-value nps-mono">:{{ summary?.httpsProxyPort || '-' }}</span>
                </div>
                <div class="nps-info-row">
                  <span class="nps-info-label">P2P</span>
                  <span class="nps-info-value nps-mono">
                    {{ summary?.serverIp || '-' }}:{{ summary?.p2pPort || '-' }}
                  </span>
                </div>
                <div class="nps-info-row">
                  <span class="nps-info-label">Log level</span>
                  <span class="nps-info-value nps-mono">{{ summary?.logLevel || '-' }}</span>
                </div>
              </div>
            </NCard>
          </NGridItem>
        </NGrid>
      </NSpin>

      <NAlert v-if="error" type="error">{{ error }}</NAlert>
    </template>
  </div>
</template>

<style scoped>
.nps-metric {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  border-radius: 12px;
  background: rgba(127, 127, 127, 0.04);
  border: 1px solid var(--n-border-color, rgba(127, 127, 127, 0.12));
}
.nps-metric-label {
  font-size: 12px;
  opacity: 0.7;
  display: flex;
  align-items: center;
  gap: 6px;
}
.nps-metric-value {
  font-size: 22px;
  font-weight: 600;
  letter-spacing: 0.02em;
}
.nps-metric-sub {
  font-size: 11px;
}
.nps-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  display: inline-block;
}
.nps-row {
  padding: 4px 0;
}
.nps-info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px 16px;
}
@media (max-width: 720px) {
  .nps-info-grid {
    grid-template-columns: 1fr;
  }
}
.nps-info-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 12px;
  border-radius: 10px;
  background: rgba(127, 127, 127, 0.04);
  border: 1px solid var(--n-border-color, rgba(127, 127, 127, 0.1));
  min-height: 38px;
}
.nps-info-label {
  font-size: 12px;
  opacity: 0.65;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.nps-info-value {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}
</style>
