<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
import {
  NButton,
  NCard,
  NDataTable,
  NDrawer,
  NDrawerContent,
  NForm,
  NFormItem,
  NInput,
  NPopconfirm,
  NSelect,
  NSpace,
  NSwitch,
  NTag,
  useMessage,
  type DataTableColumns,
  type PaginationProps,
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { hostApi, type Host, type HostPayload } from '@/api/host'
import { clientApi } from '@/api/client'
import type { Client } from '@/api/types'
import { ApiError } from '@/api/request'

const { t } = useI18n()
const message = useMessage()

const rows = ref<Host[]>([])
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
    const page = await hostApi.list({
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

onMounted(load)

async function onToggle(row: Host) {
  try {
    await hostApi.changeStatus(row.Id, row.IsClose)
    void load()
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : String(e))
  }
}
async function onDelete(row: Host) {
  try {
    await hostApi.remove(row.Id)
    void load()
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : String(e))
  }
}

// ---- form drawer ----------------------------------------------------------
const drawer = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const form = reactive<HostPayload>({
  clientId: 0,
  host: '',
  target: '',
  localProxy: false,
  header: '',
  hostchange: '',
  remark: '',
  location: '/',
  scheme: 'all',
  keyFilePath: '',
  certFilePath: '',
  autoHttps: false,
})

const schemeOptions = [
  { label: 'http', value: 'http' },
  { label: 'https', value: 'https' },
  { label: 'all', value: 'all' },
]

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

function reset() {
  form.clientId = 0
  form.host = ''
  form.target = ''
  form.localProxy = false
  form.header = ''
  form.hostchange = ''
  form.remark = ''
  form.location = '/'
  form.scheme = 'all'
  form.keyFilePath = ''
  form.certFilePath = ''
  form.autoHttps = false
}

async function openCreate() {
  await loadClients()
  reset()
  editingId.value = null
  drawer.value = true
}
async function openEdit(row: Host) {
  await loadClients()
  editingId.value = row.Id
  form.clientId = row.Client?.Id ?? 0
  form.host = row.Host
  form.target = row.Target?.TargetStr ?? ''
  form.localProxy = !!row.Target?.LocalProxy
  form.header = row.HeaderChange
  form.hostchange = row.HostChange
  form.remark = row.Remark
  form.location = row.Location || '/'
  form.scheme = row.Scheme || 'all'
  form.keyFilePath = row.KeyFilePath
  form.certFilePath = row.CertFilePath
  form.autoHttps = !!row.AutoHttps
  drawer.value = true
}

async function submit() {
  if (!form.clientId) {
    message.warning(t('host.client') + ' ?')
    return
  }
  if (!form.host) {
    message.warning(t('host.host') + ' ?')
    return
  }
  saving.value = true
  try {
    if (editingId.value) {
      await hostApi.update(editingId.value, { ...form })
    } else {
      await hostApi.create({ ...form })
    }
    drawer.value = false
    void load()
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : String(e))
  } finally {
    saving.value = false
  }
}

const columns = computed<DataTableColumns<Host>>(() => [
  { title: 'ID', key: 'Id', width: 70 },
  { title: t('host.host'), key: 'Host', minWidth: 160 },
  { title: t('host.scheme'), key: 'Scheme', width: 80 },
  { title: t('host.location'), key: 'Location', width: 100 },
  {
    title: t('host.target'),
    key: 'target',
    minWidth: 180,
    render: (r) => r.Target?.TargetStr ?? '',
  },
  {
    title: t('host.client'),
    key: 'client',
    width: 120,
    render: (r) => r.Client?.Remark || r.Client?.VerifyKey,
  },
  { title: t('host.remark'), key: 'Remark', minWidth: 100 },
  {
    title: t('common.status'),
    key: 'status',
    width: 100,
    render: (r) =>
      h(
        NTag,
        { type: !r.IsClose ? 'success' : 'default', size: 'small', round: true },
        () => (!r.IsClose ? t('common.open') : t('common.close')),
      ),
  },
  {
    title: t('common.actions'),
    key: 'actions',
    width: 220,
    fixed: 'right',
    render: (r) =>
      h(NSpace, { size: 4 }, () => [
        h(
          NButton,
          {
            size: 'tiny',
            type: !r.IsClose ? 'warning' : 'primary',
            onClick: () => onToggle(r),
          },
          () => (!r.IsClose ? '⏸' : '▶'),
        ),
        h(
          NButton,
          { size: 'tiny', type: 'success', onClick: () => openEdit(r) },
          () => t('common.edit'),
        ),
        h(
          NPopconfirm,
          { onPositiveClick: () => onDelete(r) },
          {
            trigger: () =>
              h(NButton, { size: 'tiny', type: 'error' }, () =>
                t('common.delete'),
              ),
            default: () => t('host.deleteConfirm'),
          },
        ),
      ]),
  },
])

const isHttps = computed(
  () => form.scheme === 'https' || form.scheme === 'all',
)
</script>

<template>
  <NCard :title="t('host.list')">
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

    <NDataTable
      :columns="columns"
      :data="rows"
      :loading="loading"
      :pagination="pagination"
      :row-key="(r: Host) => r.Id"
      remote
      :bordered="false"
      :scroll-x="1300"
      striped
    />

    <NDrawer v-model:show="drawer" :width="560" placement="right">
      <NDrawerContent
        :title="editingId ? t('host.edit') : t('host.add')"
        closable
      >
        <NForm label-placement="top">
          <NFormItem :label="t('host.client')">
            <NSelect
              v-model:value="form.clientId"
              :options="clientOptions"
              filterable
            />
          </NFormItem>
          <NFormItem :label="t('host.host')">
            <NInput v-model:value="form.host" placeholder="example.com" />
          </NFormItem>
          <NFormItem :label="t('host.scheme')">
            <NSelect v-model:value="form.scheme" :options="schemeOptions" />
          </NFormItem>
          <NFormItem :label="t('host.location')">
            <NInput v-model:value="form.location" />
          </NFormItem>
          <NFormItem :label="t('host.target')">
            <NInput
              v-model:value="form.target"
              type="textarea"
              :rows="3"
              :placeholder="t('tunnel.targetTip')"
            />
          </NFormItem>
          <NFormItem :label="t('host.hostchange')">
            <NInput v-model:value="form.hostchange" />
          </NFormItem>
          <NFormItem :label="t('host.header')">
            <NInput
              v-model:value="form.header"
              type="textarea"
              :rows="3"
              :placeholder="t('host.headerTip')"
            />
          </NFormItem>
          <NFormItem :label="t('host.localProxy')">
            <NSwitch v-model:value="form.localProxy" />
          </NFormItem>
          <template v-if="isHttps">
            <NFormItem :label="t('host.autoHttps')">
              <NSwitch v-model:value="form.autoHttps" />
            </NFormItem>
            <NFormItem :label="t('host.certFilePath')">
              <NInput v-model:value="form.certFilePath" />
            </NFormItem>
            <NFormItem :label="t('host.keyFilePath')">
              <NInput v-model:value="form.keyFilePath" />
            </NFormItem>
          </template>
          <NFormItem :label="t('host.remark')">
            <NInput v-model:value="form.remark" />
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
