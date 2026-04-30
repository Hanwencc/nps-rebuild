<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  NButton,
  NCard,
  NDataTable,
  NInput,
  NInputGroup,
  NModal,
  NPopconfirm,
  NSpace,
  NTag,
  NText,
  NTooltip,
  useDialog,
  useMessage,
  type DataTableColumns,
  type PaginationProps,
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { clientApi } from '@/api/client'
import type { Client, QuickInfo } from '@/api/types'
import { useAuthStore } from '@/stores/auth'
import { ApiError } from '@/api/request'

const { t } = useI18n()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const auth = useAuthStore()

const rows = ref<Client[]>([])
const total = ref(0)
const loading = ref(false)
const search = ref('')

const pagination = reactive<PaginationProps>({
  page: 1,
  pageSize: 10,
  showSizePicker: true,
  pageSizes: [10, 25, 50, 100],
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

async function load() {
  loading.value = true
  try {
    const page = await clientApi.list({
      offset: ((pagination.page ?? 1) - 1) * (pagination.pageSize ?? 10),
      limit: pagination.pageSize,
      search: search.value,
    })
    rows.value = page.items
    total.value = page.total
    pagination.itemCount = page.total
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : String(e))
  } finally {
    loading.value = false
  }
}

async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    message.success(t('common.copied'))
  } catch {
    message.error('clipboard not available')
  }
}

// ----- quick install command modal --------------------------------------
//
// Operators expect three flavours of install snippet:
//   1. Plain TCP/KCP — the historical one-liner.
//   2. TLS — same flags + tls_enable + tls_server_fingerprint pin.
//      Falls back to the unsafe form (no pin) when nps hasn't yet
//      published its cert fingerprint, with a visible warning.
//   3. docker-compose — copy/paste-able service block, useful for
//      fleets managed via compose.
//
// All three are rendered in one modal so the operator can pick the
// right one without re-clicking the table button.

const quickInfo = ref<QuickInfo | null>(null)
const quickModalOpen = ref(false)
const quickLoading = ref(false)

const winSuffix = computed(() =>
  /windows/i.test(navigator.userAgent) || /Win/.test(navigator.platform)
    ? '.exe'
    : '',
)

const cmdPlain = computed(() => {
  const i = quickInfo.value
  if (!i) return ''
  return `./npc${winSuffix.value} -server=${i.ip}:${i.bridgePort} -vkey=${i.vkey} -type=${i.bridgeType}`
})

// TLS bridge always speaks TCP — the kcp listener is unencrypted, so
// `-type=tcp` is hard-coded here regardless of the server's primary
// bridge_type.
const cmdTls = computed(() => {
  const i = quickInfo.value
  if (!i) return ''
  const base = `./npc${winSuffix.value} -server=${i.ip}:${i.tlsPort} -vkey=${i.vkey} -type=tcp -tls_enable=true`
  return i.tlsFingerprint
    ? `${base} -tls_server_fingerprint=${i.tlsFingerprint}`
    : base
})

const cmdCompose = computed(() => {
  const i = quickInfo.value
  if (!i) return ''
  const useTls = !!i.tlsFingerprint
  const port = useTls ? i.tlsPort : i.bridgePort
  const lines = [
    'version: "3.8"',
    'services:',
    '  npc:',
    '    image: ghcr.io/hanwencc/npc:latest',
    `    container_name: npc-${i.id}`,
    '    restart: always',
    '    network_mode: host',
    '    command:',
    `      - -server=${i.ip}:${port}`,
    `      - -vkey=${i.vkey}`,
    `      - -type=${useTls ? 'tcp' : i.bridgeType}`,
  ]
  if (useTls) {
    lines.push('      - -tls_enable=true')
    lines.push(`      - -tls_server_fingerprint=${i.tlsFingerprint}`)
  }
  return lines.join('\n')
})

async function onCopyCmd(row: Client) {
  quickLoading.value = true
  quickInfo.value = null
  quickModalOpen.value = true
  try {
    quickInfo.value = await clientApi.quickInfo(row.Id)
  } catch (e) {
    quickModalOpen.value = false
    message.error(e instanceof ApiError ? e.message : String(e))
  } finally {
    quickLoading.value = false
  }
}

async function onToggleStatus(row: Client) {
  try {
    await clientApi.changeStatus(row.Id, !row.Status)
    row.Status = !row.Status
    message.success('OK')
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : String(e))
  }
}

function onAdd() {
  router.push({ name: 'client-new' })
}
function onEdit(row: Client) {
  router.push({ name: 'client-edit', params: { id: row.Id } })
}
function onDelete(row: Client) {
  dialog.warning({
    title: t('common.confirm'),
    content: t('client.deleteConfirm'),
    positiveText: t('common.confirm'),
    negativeText: t('common.cancel'),
    onPositiveClick: async () => {
      try {
        await clientApi.remove(row.Id)
        message.success('OK')
        void load()
      } catch (e) {
        message.error(e instanceof ApiError ? e.message : String(e))
      }
    },
  })
}

const columns = computed<DataTableColumns<Client>>(() => [
  { title: t('client.id'), key: 'Id', width: 80 },
  { title: t('client.remark'), key: 'Remark', minWidth: 120, ellipsis: { tooltip: true } },
  { title: t('client.version'), key: 'Version', width: 90 },
  {
    title: t('client.vkey'),
    key: 'VerifyKey',
    minWidth: 160,
    render: (row) =>
      row.NoStore
        ? h(NTag, { type: 'info', size: 'small' }, () => 'public')
        : row.VerifyKey,
  },
  { title: t('client.address'), key: 'Addr', width: 160 },
  {
    title: t('client.inletFlow'),
    key: 'inlet',
    width: 110,
    render: (row) => bytesHuman(row.Flow?.InletFlow),
  },
  {
    title: t('client.exportFlow'),
    key: 'export',
    width: 110,
    render: (row) => bytesHuman(row.Flow?.ExportFlow),
  },
  {
    title: t('client.speed'),
    key: 'speed',
    width: 110,
    render: (row) => bytesHuman(row.Rate?.NowRate) + '/s',
  },
  {
    title: t('common.status'),
    key: 'status',
    width: 100,
    render: (row) =>
      h(
        NTag,
        { type: row.Status ? 'success' : 'default', size: 'small', round: true },
        () => (row.Status ? t('common.open') : t('common.close')),
      ),
  },
  {
    title: t('client.connect'),
    key: 'connect',
    width: 140,
    render: (row) => {
      if (!row.IsConnect) {
        return h(
          NTag,
          { type: 'default', size: 'small', round: true },
          () => t('common.offline'),
        )
      }
      return h(NSpace, { size: 4, wrap: false }, () => [
        h(
          NTag,
          { type: 'info', size: 'small', round: true },
          () => t('common.online'),
        ),
        h(
          NTag,
          {
            type: row.IsTls ? 'success' : 'warning',
            size: 'small',
            round: true,
            bordered: false,
          },
          () => (row.IsTls ? t('client.tls') : t('client.tcp')),
        ),
      ])
    },
  },
  {
    title: t('common.actions'),
    key: 'actions',
    width: 240,
    fixed: 'right',
    render: (row) =>
      h(NSpace, { size: 6 }, () => [
        auth.isAdmin
          ? h(
              NTooltip,
              {},
              {
                trigger: () =>
                  h(
                    NButton,
                    {
                      size: 'tiny',
                      type: row.Status ? 'warning' : 'primary',
                      onClick: () => onToggleStatus(row),
                    },
                    () => (row.Status ? '⏸' : '▶'),
                  ),
                default: () =>
                  row.Status ? t('common.close') : t('common.open'),
              },
            )
          : null,
        h(
          NButton,
          {
            size: 'tiny',
            type: 'success',
            onClick: () => onEdit(row),
          },
          () => t('common.edit'),
        ),
        h(
          NButton,
          {
            size: 'tiny',
            onClick: () => onCopyCmd(row),
          },
          () => t('client.quickCmd'),
        ),
        auth.isAdmin
          ? h(
              NPopconfirm,
              {
                onPositiveClick: () => onDelete(row),
              },
              {
                trigger: () =>
                  h(NButton, { size: 'tiny', type: 'error' }, () =>
                    t('common.delete'),
                  ),
                default: () => t('client.deleteConfirm'),
              },
            )
          : null,
      ]),
  },
])

onMounted(load)
</script>

<template>
  <NCard :title="t('client.list')">
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
        <NButton v-if="auth.isAdmin" type="primary" @click="onAdd">
          + {{ t('common.add') }}
        </NButton>
      </NSpace>
    </template>

    <NDataTable
      :columns="columns"
      :data="rows"
      :loading="loading"
      :pagination="pagination"
      :row-key="(r: Client) => r.Id"
      remote
      :bordered="false"
      :scroll-x="1400"
      striped
    />
  </NCard>

  <NModal
    v-model:show="quickModalOpen"
    preset="card"
    :title="t('client.quickCmdTitle')"
    style="max-width: 760px"
    :bordered="false"
    size="huge"
  >
    <NSpace v-if="quickLoading" justify="center">
      <NText>...</NText>
    </NSpace>
    <NSpace v-else-if="quickInfo" vertical size="large" style="width: 100%">
      <div>
        <NText strong>{{ t('client.quickCmdPlain') }}</NText>
        <NInputGroup style="margin-top: 6px">
          <NInput
            :value="cmdPlain"
            type="text"
            readonly
            @focus="(e: FocusEvent) => (e.target as HTMLInputElement).select()"
          />
          <NButton type="primary" @click="copyText(cmdPlain)">
            {{ t('client.quickCmdCopy') }}
          </NButton>
        </NInputGroup>
      </div>

      <div>
        <NText strong>
          {{
            quickInfo.tlsFingerprint
              ? t('client.quickCmdTls')
              : t('client.quickCmdTlsNoFp')
          }}
        </NText>
        <NInputGroup style="margin-top: 6px">
          <NInput
            :value="cmdTls"
            type="text"
            readonly
            @focus="(e: FocusEvent) => (e.target as HTMLInputElement).select()"
          />
          <NButton type="primary" @click="copyText(cmdTls)">
            {{ t('client.quickCmdCopy') }}
          </NButton>
        </NInputGroup>
        <NText
          v-if="!quickInfo.tlsFingerprint"
          depth="3"
          style="font-size: 12px; display: block; margin-top: 4px"
        >
          {{ t('client.quickCmdFpHint') }}
        </NText>
      </div>

      <div>
        <NSpace justify="space-between" align="center" style="width: 100%">
          <NText strong>{{ t('client.quickCmdCompose') }}</NText>
          <NButton size="small" type="primary" @click="copyText(cmdCompose)">
            {{ t('client.quickCmdCopy') }}
          </NButton>
        </NSpace>
        <NInput
          :value="cmdCompose"
          type="textarea"
          readonly
          :autosize="{ minRows: 8, maxRows: 16 }"
          style="margin-top: 6px; font-family: monospace"
        />
      </div>
    </NSpace>
  </NModal>
</template>
