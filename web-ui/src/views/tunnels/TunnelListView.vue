<script setup lang="ts">
import { computed, h, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { RouterLink } from 'vue-router'
import {
  NButton,
  NCard,
  NDataTable,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NPopconfirm,
  NSelect,
  NSpace,
  NSwitch,
  NTag,
  NAlert,
  useMessage,
  type DataTableColumns,
  type PaginationProps,
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { tunnelApi, type Tunnel, type TunnelPayload } from '@/api/tunnel'
import { socks5Api, type Socks5GatewayStatus } from '@/api/socks5'
import { clientApi } from '@/api/client'
import type { Client } from '@/api/types'
import { ApiError } from '@/api/request'

const { t } = useI18n()
const route = useRoute()
const message = useMessage()

/**
 * Map the route param ('http') to the backend mode token ('httpProxy').
 * The other mode names line up directly.
 */
const ROUTE_TO_BACKEND: Record<string, string> = {
  tcp: 'tcp',
  udp: 'udp',
  socks5: 'socks5',
  http: 'httpProxy',
  secret: 'secret',
  p2p: 'p2p',
  file: 'file',
}

const mode = computed(() => {
  const m = (route.params.mode as string) || 'tcp'
  return ROUTE_TO_BACKEND[m] ?? m
})

const rows = ref<Tunnel[]>([])
const loading = ref(false)
const search = ref('')
const pagination = reactive<PaginationProps>({
  page: 1,
  pageSize: 10,
  showSizePicker: true,
  pageSizes: [10, 25, 50],
  itemCount: 0,
  onChange: (p: number) => {
    pagination.page = p
    void load()
  },
  onUpdatePageSize: (s: number) => {
    pagination.pageSize = s
    pagination.page = 1
    void load()
  },
})

async function load() {
  loading.value = true
  try {
    const page = await tunnelApi.list({
      mode: mode.value,
      offset: ((pagination.page ?? 1) - 1) * (pagination.pageSize ?? 10),
      limit: pagination.pageSize,
      search: search.value,
    })
    rows.value = page.items
    pagination.itemCount = page.total
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : String(e))
  } finally {
    loading.value = false
  }
}

watch(mode, () => {
  pagination.page = 1
  void load()
  if (mode.value === 'socks5') void loadSocks5Gateway()
})

onMounted(() => {
  void load()
  if (mode.value === 'socks5') void loadSocks5Gateway()
})

async function onToggle(row: Tunnel) {
  try {
    if (row.Status) {
      await tunnelApi.stop(row.Id)
    } else {
      await tunnelApi.start(row.Id)
    }
    void load()
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : String(e))
  }
}
async function onCopy(row: Tunnel) {
  try {
    await tunnelApi.copy(row.Id)
    message.success(t('tunnel.copySuccess'))
    void load()
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : String(e))
  }
}
async function onDelete(row: Tunnel) {
  try {
    await tunnelApi.remove(row.Id)
    void load()
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : String(e))
  }
}

// ----- form drawer ---------------------------------------------------------
const drawer = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const form = reactive<TunnelPayload>({
  clientId: 0,
  mode: 'tcp',
  port: 0,
  serverIp: '0.0.0.0',
  target: '',
  localProxy: false,
  password: '',
  username: '',
  remark: '',
  localPath: '',
  stripPre: '',
  protoVersion: '',
})

// Phase 9 — shared SOCKS5 gateway status. Loaded lazily when the user
// opens the tunnels view in socks5 mode so we can warn that the global
// listener is disabled (port=0) and link them to the settings page.
const socks5Gateway = ref<Socks5GatewayStatus | null>(null)
async function loadSocks5Gateway() {
  try {
    socks5Gateway.value = await socks5Api.gateway()
  } catch {
    socks5Gateway.value = null
  }
}

const clientOptions = ref<{ label: string; value: number }[]>([])
async function loadClients() {
  try {
    const page = await clientApi.list({ offset: 0, limit: 1000 })
    clientOptions.value = page.items.map((c: Client) => ({
      label: `[${c.Id}] ${c.Remark || c.VerifyKey}`,
      value: c.Id,
    }))
  } catch {
    /* ignore */
  }
}

function resetForm() {
  form.clientId = 0
  form.mode = mode.value
  form.port = 0
  form.serverIp = '0.0.0.0'
  form.target = ''
  form.localProxy = false
  form.password = ''
  form.username = ''
  form.remark = ''
  form.localPath = ''
  form.stripPre = ''
  form.protoVersion = ''
}

async function openCreate() {
  await loadClients()
  resetForm()
  editingId.value = null
  drawer.value = true
}

async function openEdit(row: Tunnel) {
  await loadClients()
  editingId.value = row.Id
  form.clientId = row.Client?.Id ?? 0
  form.mode = row.Mode
  form.port = row.Port
  form.serverIp = row.ServerIp
  form.target = row.Target?.TargetStr ?? ''
  form.localProxy = !!row.Target?.LocalProxy
  form.password = row.Password
  form.username = row.Username
  form.remark = row.Remark
  form.localPath = row.LocalPath
  form.stripPre = row.StripPre
  form.protoVersion = row.ProtoVersion
  drawer.value = true
}

async function submit() {
  if (!form.clientId) {
    message.warning(t('tunnel.client') + ' ?')
    return
  }
  if (form.mode === 'socks5' && !form.username) {
    message.warning(t('tunnel.usernameRequired'))
    return
  }
  saving.value = true
  try {
    if (editingId.value) {
      await tunnelApi.update(editingId.value, { ...form })
    } else {
      await tunnelApi.create({ ...form })
    }
    drawer.value = false
    void load()
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : String(e))
  } finally {
    saving.value = false
  }
}

function bytesHuman(n?: number): string {
  if (!n) return '0 B'
  const u = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  let v = n
  while (v >= 1024 && i < u.length - 1) {
    v /= 1024
    i++
  }
  return v.toFixed(2) + ' ' + u[i]
}

const isPortMode = computed(
  () => !['secret', 'p2p', 'file', 'socks5'].includes(mode.value),
)
const isSocks5Mode = computed(() => mode.value === 'socks5')
const isSecretMode = computed(() =>
  ['secret', 'p2p'].includes(mode.value),
)
const isFileMode = computed(() => mode.value === 'file')
const isHttpMode = computed(() => mode.value === 'httpProxy')

const columns = computed<DataTableColumns<Tunnel>>(() => {
  const base: DataTableColumns<Tunnel> = [
    { title: 'ID', key: 'Id', width: 70 },
    { title: t('tunnel.remark'), key: 'Remark', minWidth: 120, ellipsis: { tooltip: true } },
    {
      title: t('tunnel.client'),
      key: 'client',
      width: 120,
      render: (r) => r.Client?.Remark || r.Client?.VerifyKey,
    },
  ]
  if (isPortMode.value) {
    base.push({ title: t('tunnel.port'), key: 'Port', width: 100 })
  }
  if (isSocks5Mode.value) {
    base.push({ title: t('tunnel.username'), key: 'Username', width: 140 })
    base.push({ title: t('tunnel.password'), key: 'Password', width: 140 })
  }
  if (isSecretMode.value) {
    base.push({ title: t('tunnel.password'), key: 'Password', width: 140 })
  }
  if (isFileMode.value) {
    base.push({ title: t('tunnel.localPath'), key: 'LocalPath', minWidth: 160 })
  } else {
    base.push({
      title: t('tunnel.target'),
      key: 'target',
      minWidth: 180,
      render: (r) => r.Target?.TargetStr ?? '',
    })
  }
  base.push(
    {
      title: t('client.inletFlow'),
      key: 'inlet',
      width: 100,
      render: (r) => bytesHuman(r.Flow?.InletFlow),
    },
    {
      title: t('client.exportFlow'),
      key: 'export',
      width: 100,
      render: (r) => bytesHuman(r.Flow?.ExportFlow),
    },
    {
      title: t('common.status'),
      key: 'status',
      width: 100,
      render: (r) =>
        h(
          NTag,
          {
            type: r.Status ? 'success' : 'default',
            size: 'small',
            round: true,
          },
          () => (r.Status ? t('common.open') : t('common.close')),
        ),
    },
    {
      title: t('tunnel.runStatus'),
      key: 'run',
      width: 100,
      render: (r) =>
        h(
          NTag,
          {
            type: r.RunStatus ? 'info' : 'default',
            size: 'small',
            round: true,
          },
          () => (r.RunStatus ? t('tunnel.running') : t('tunnel.stopped')),
        ),
    },
    {
      title: t('common.actions'),
      key: 'actions',
      width: 260,
      fixed: 'right',
      render: (r) =>
        h(NSpace, { size: 4 }, () => [
          h(
            NButton,
            {
              size: 'tiny',
              type: r.Status ? 'warning' : 'primary',
              onClick: () => onToggle(r),
            },
            () => (r.Status ? '⏸' : '▶'),
          ),
          h(
            NButton,
            { size: 'tiny', type: 'success', onClick: () => openEdit(r) },
            () => t('common.edit'),
          ),
          h(
            NButton,
            { size: 'tiny', onClick: () => onCopy(r) },
            () => t('tunnel.copy'),
          ),
          h(
            NPopconfirm,
            { onPositiveClick: () => onDelete(r) },
            {
              trigger: () =>
                h(NButton, { size: 'tiny', type: 'error' }, () =>
                  t('common.delete'),
                ),
              default: () => t('tunnel.deleteConfirm'),
            },
          ),
        ]),
    },
  )
  return base
})

const modeTitle = computed(() => {
  const r = (route.params.mode as string) || 'tcp'
  return t('nav.' + r) + ' · ' + t('tunnel.list')
})

// expose helpers used from template by reference
const showProtoVersion = computed(() => isHttpMode.value)
</script>

<template>
  <NCard :title="modeTitle">
    <template #header-extra>
      <NSpace>
        <NInput
          v-model:value="search"
          :placeholder="t('common.search')"
          clearable
          style="width: 200px"
          @update:value="
            () => {
              pagination.page = 1
              load()
            }
          "
        />
        <NButton @click="load">{{ t('common.refresh') }}</NButton>
        <NButton type="primary" @click="openCreate">
          + {{ t('common.add') }}
        </NButton>
      </NSpace>
    </template>

    <NAlert
      v-if="isSocks5Mode && socks5Gateway"
      :type="socks5Gateway.listening ? 'success' : 'warning'"
      :show-icon="true"
      style="margin-bottom: 12px"
    >
      <template v-if="socks5Gateway.listening">
        {{ t('tunnel.socks5SharedPort') }}: {{ socks5Gateway.addr }}
        <span style="opacity: 0.7">({{ socks5Gateway.routes }})</span>
      </template>
      <template v-else>
        {{ t('tunnel.socks5GatewayDisabled') }}
        <RouterLink to="/settings" style="margin-left: 8px">
          {{ t('tunnel.goToSettings') }}
        </RouterLink>
      </template>
    </NAlert>

    <NDataTable
      :columns="columns"
      :data="rows"
      :loading="loading"
      :pagination="pagination"
      :row-key="(r: Tunnel) => r.Id"
      remote
      :bordered="false"
      :scroll-x="1400"
      striped
    />

    <NDrawer v-model:show="drawer" :width="520" placement="right">
      <NDrawerContent
        :title="editingId ? t('tunnel.edit') : t('tunnel.add')"
        closable
      >
        <NForm label-placement="top">
          <NFormItem :label="t('tunnel.client')">
            <NSelect
              v-model:value="form.clientId"
              :options="clientOptions"
              filterable
            />
          </NFormItem>
          <NFormItem :label="t('tunnel.remark')">
            <NInput v-model:value="form.remark" />
          </NFormItem>
          <NFormItem v-if="isPortMode" :label="t('tunnel.port')">
            <NInputNumber v-model:value="form.port" :min="0" class="!w-full" />
          </NFormItem>
          <NFormItem v-if="isPortMode" :label="t('tunnel.serverIp')">
            <NInput v-model:value="form.serverIp" />
          </NFormItem>
          <template v-if="isSocks5Mode">
            <NFormItem :label="t('tunnel.username')">
              <NInput v-model:value="form.username" />
            </NFormItem>
            <NFormItem :label="t('tunnel.password')">
              <NInput v-model:value="form.password" />
            </NFormItem>
          </template>
          <NFormItem v-if="isSecretMode" :label="t('tunnel.password')">
            <NInput v-model:value="form.password" />
          </NFormItem>
          <NFormItem v-if="isFileMode" :label="t('tunnel.localPath')">
            <NInput v-model:value="form.localPath" />
          </NFormItem>
          <template v-else>
            <NFormItem :label="t('tunnel.target')">
              <NInput
                v-model:value="form.target"
                type="textarea"
                :rows="3"
                :placeholder="t('tunnel.targetTip')"
              />
            </NFormItem>
          </template>
          <NFormItem v-if="isHttpMode" :label="t('tunnel.stripPre')">
            <NInput v-model:value="form.stripPre" />
          </NFormItem>
          <NFormItem v-if="showProtoVersion" :label="t('tunnel.protoVersion')">
            <NInput v-model:value="form.protoVersion" />
          </NFormItem>
          <NFormItem :label="t('tunnel.localProxy')">
            <NSwitch v-model:value="form.localProxy" />
          </NFormItem>
        </NForm>
        <template #footer>
          <NSpace justify="end">
            <NButton @click="drawer = false">{{ t('common.cancel') }}</NButton>
            <NButton type="primary" :loading="saving" @click="submit">
              {{ t('common.save') }}
            </NButton>
          </NSpace>
        </template>
      </NDrawerContent>
    </NDrawer>
  </NCard>
</template>
