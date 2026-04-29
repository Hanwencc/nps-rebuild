<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  NButton,
  NCard,
  NDataTable,
  NInput,
  NPopconfirm,
  NSpace,
  NTag,
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

async function onCopyCmd(row: Client) {
  let info: QuickInfo
  try {
    info = await clientApi.quickInfo(row.Id)
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : String(e))
    return
  }
  const winSuffix =
    /windows/i.test(navigator.userAgent) || /Win/.test(navigator.platform)
      ? '.exe'
      : ''
  const cmd = `./npc${winSuffix} -server=${info.ip}:${info.bridgePort} -vkey=${info.vkey} -type=${info.bridgeType}`
  await copyText(cmd)
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
    width: 100,
    render: (row) =>
      h(
        NTag,
        { type: row.IsConnect ? 'info' : 'default', size: 'small', round: true },
        () => (row.IsConnect ? t('common.online') : t('common.offline')),
      ),
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
</template>
