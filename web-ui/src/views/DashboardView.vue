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
  NProgress,
  NAlert,
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { dashboardApi, type DashboardSummary } from '@/api/dashboard'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const auth = useAuthStore()

const summary = ref<DashboardSummary | null>(null)
const loading = ref(false)
const error = ref('')
let timer: number | null = null

async function load() {
  loading.value = true
  try {
    summary.value = await dashboardApi.summary()
    error.value = ''
  } catch (e: unknown) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

function fmtBytes(n: number): string {
  if (!n) return '0 B'
  const u = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  let v = n
  while (v >= 1024 && i < u.length - 1) {
    v /= 1024
    i++
  }
  return `${v.toFixed(2)} ${u[i]}`
}

const tunnelTotal = computed(() => {
  const c = summary.value?.tunnelCount
  if (!c) return 0
  return c.tcp + c.udp + c.socks5 + c.httpProxy + c.secret + c.p2p
})

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
          <NGridItem span="1 m:1">
            <NCard>
              <NStatistic :label="t('dashboard.clients')">
                <span>{{ summary?.clientOnlineCount ?? 0 }}</span>
                <template #suffix> / {{ summary?.clientCount ?? 0 }}</template>
              </NStatistic>
            </NCard>
          </NGridItem>
          <NGridItem span="1 m:1">
            <NCard>
              <NStatistic :label="t('dashboard.hosts')" :value="summary?.hostCount ?? 0" />
            </NCard>
          </NGridItem>
          <NGridItem span="1 m:1">
            <NCard>
              <NStatistic :label="t('dashboard.tunnels')" :value="tunnelTotal" />
            </NCard>
          </NGridItem>
          <NGridItem span="1 m:1">
            <NCard>
              <NStatistic :label="t('dashboard.connections')" :value="summary?.connections ?? 0" />
            </NCard>
          </NGridItem>

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

          <NGridItem :span="2">
            <NCard :title="t('dashboard.flow')">
              <NSpace>
                <NStatistic :label="t('dashboard.inFlow')" :value="fmtBytes(summary?.flow.in ?? 0)" />
                <NStatistic :label="t('dashboard.outFlow')" :value="fmtBytes(summary?.flow.out ?? 0)" />
              </NSpace>
            </NCard>
          </NGridItem>

          <NGridItem :span="2">
            <NCard :title="t('dashboard.system')">
              <div class="space-y-3">
                <div>
                  <div class="flex justify-between text-sm">
                    <span>CPU</span>
                    <span>{{ Math.round(summary?.system.cpu ?? 0) }}%</span>
                  </div>
                  <NProgress
                    type="line"
                    :percentage="Math.min(100, Math.round(summary?.system.cpu ?? 0))"
                    :show-indicator="false"
                  />
                </div>
                <div>
                  <div class="flex justify-between text-sm">
                    <span>MEM</span>
                    <span>{{ Math.round(summary?.system.mem ?? 0) }}%</span>
                  </div>
                  <NProgress
                    type="line"
                    :percentage="Math.min(100, Math.round(summary?.system.mem ?? 0))"
                    :show-indicator="false"
                  />
                </div>
                <div>
                  <div class="flex justify-between text-sm">
                    <span>SWAP</span>
                    <span>{{ Math.round(summary?.system.swap ?? 0) }}%</span>
                  </div>
                  <NProgress
                    type="line"
                    :percentage="Math.min(100, Math.round(summary?.system.swap ?? 0))"
                    :show-indicator="false"
                  />
                </div>
                <div class="text-sm text-slate-500">{{ t('dashboard.load') }}: {{ summary?.load }}</div>
              </div>
            </NCard>
          </NGridItem>

          <NGridItem :span="2">
            <NCard :title="t('dashboard.serverInfo')">
              <NSpace vertical>
                <div>{{ t('dashboard.version') }}: {{ summary?.version }}</div>
                <div>{{ t('dashboard.bridge') }}: {{ summary?.bridgeType }} :{{ summary?.bridgePort }}</div>
                <div>HTTP Proxy: :{{ summary?.httpProxyPort || '-' }}</div>
                <div>HTTPS Proxy: :{{ summary?.httpsProxyPort || '-' }}</div>
                <div>P2P: {{ summary?.serverIp || '-' }}:{{ summary?.p2pPort || '-' }}</div>
                <div>Log level: {{ summary?.logLevel || '-' }}</div>
              </NSpace>
            </NCard>
          </NGridItem>
        </NGrid>
      </NSpin>

      <NAlert v-if="error" type="error">{{ error }}</NAlert>
    </template>
  </div>
</template>
