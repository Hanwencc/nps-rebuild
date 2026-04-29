<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue'
import {
  NAlert,
  NButton,
  NCard,
  NDataTable,
  NDatePicker,
  NForm,
  NFormItem,
  NInput,
  NInputGroup,
  NModal,
  NPopconfirm,
  NSpace,
  NSwitch,
  NTag,
  NText,
  useMessage,
  type DataTableColumns,
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import {
  tokenApi,
  type ApiToken,
  type TokenSecretReveal,
  type TokenWriteRequest,
} from '@/api/token'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const message = useMessage()
const auth = useAuthStore()

const loading = ref(false)
const rows = ref<ApiToken[]>([])

async function load() {
  if (!auth.isAdmin) return
  loading.value = true
  try {
    rows.value = await tokenApi.list()
  } catch (e: unknown) {
    message.error((e as Error).message)
  } finally {
    loading.value = false
  }
}
onMounted(load)

// ---------- Create / Edit form ----------------------------------------
interface FormState {
  id: number | null
  remark: string
  allowedPathPrefix: string
  allowedMethods: string // CSV input
  allowIpsText: string // newline-separated
  expiresAt: number | null // ms epoch for NDatePicker
  disabled: boolean
}
const blank = (): FormState => ({
  id: null,
  remark: '',
  allowedPathPrefix: '',
  allowedMethods: '',
  allowIpsText: '',
  expiresAt: null,
  disabled: false,
})
const showForm = ref(false)
const submitting = ref(false)
const form = reactive<FormState>(blank())

function openCreate() {
  Object.assign(form, blank())
  showForm.value = true
}
function openEdit(row: ApiToken) {
  Object.assign(form, {
    id: row.id,
    remark: row.remark,
    allowedPathPrefix: row.allowedPathPrefix,
    allowedMethods: (row.allowedMethods ?? []).join(','),
    allowIpsText: (row.allowIps ?? []).join('\n'),
    expiresAt: row.expiresAt > 0 ? row.expiresAt * 1000 : null,
    disabled: row.disabled,
  })
  showForm.value = true
}

function buildPayload(): TokenWriteRequest {
  return {
    remark: form.remark.trim(),
    allowedPathPrefix: form.allowedPathPrefix.trim(),
    allowedMethods: form.allowedMethods
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean),
    allowIps: form.allowIpsText
      .split(/\r?\n/)
      .map((s) => s.trim())
      .filter(Boolean),
    expiresAt: form.expiresAt ? Math.floor(form.expiresAt / 1000) : 0,
    disabled: form.disabled,
  }
}

// ---------- secret reveal ---------------------------------------------
const revealOpen = ref(false)
const revealKeyId = ref('')
const revealSecret = ref('')

function showSecret(r: TokenSecretReveal) {
  revealKeyId.value = r.token.keyId
  revealSecret.value = r.secret
  revealOpen.value = true
}

async function copyText(s: string) {
  try {
    await navigator.clipboard.writeText(s)
    message.success(t('common.copied'))
  } catch {
    message.error('clipboard unavailable')
  }
}

async function submit() {
  submitting.value = true
  try {
    const payload = buildPayload()
    if (form.id == null) {
      const r = await tokenApi.create(payload)
      message.success(t('common.saved'))
      showForm.value = false
      await load()
      showSecret(r)
    } else {
      await tokenApi.update(form.id, payload)
      message.success(t('common.saved'))
      showForm.value = false
      await load()
    }
  } catch (e: unknown) {
    message.error((e as Error).message)
  } finally {
    submitting.value = false
  }
}

async function rotate(row: ApiToken) {
  try {
    const r = await tokenApi.rotate(row.id)
    await load()
    showSecret(r)
  } catch (e: unknown) {
    message.error((e as Error).message)
  }
}

async function remove(row: ApiToken) {
  try {
    await tokenApi.remove(row.id)
    message.success(t('common.saved'))
    await load()
  } catch (e: unknown) {
    message.error((e as Error).message)
  }
}

function fmtTs(s: number) {
  if (!s) return '-'
  return new Date(s * 1000).toLocaleString()
}

const columns = computed<DataTableColumns<ApiToken>>(() => [
  { title: t('token.keyId'), key: 'keyId', width: 220 },
  { title: t('token.remark'), key: 'remark' },
  {
    title: t('token.scope'),
    key: 'scope',
    render: (r) =>
      h(NSpace, { size: 4, wrap: true }, () => {
        const tags = []
        if (r.allowedPathPrefix)
          tags.push(h(NTag, { type: 'info', size: 'small' }, () => r.allowedPathPrefix))
        if (r.allowedMethods?.length)
          tags.push(
            h(NTag, { type: 'success', size: 'small' }, () =>
              r.allowedMethods.join(','),
            ),
          )
        if (r.allowIps?.length)
          tags.push(
            h(NTag, { type: 'warning', size: 'small' }, () =>
              `${r.allowIps.length} IP`,
            ),
          )
        if (!tags.length) tags.push(h(NTag, { size: 'small' }, () => '*'))
        return tags
      }),
  },
  {
    title: t('common.status'),
    key: 'status',
    width: 100,
    render: (r) =>
      r.disabled
        ? h(NTag, { type: 'error', size: 'small' }, () => t('token.disabled'))
        : h(NTag, { type: 'success', size: 'small' }, () => t('common.open')),
  },
  {
    title: t('token.expiresAt'),
    key: 'expiresAt',
    width: 180,
    render: (r) => (r.expiresAt > 0 ? fmtTs(r.expiresAt) : t('token.neverExpire')),
  },
  {
    title: t('token.lastUsed'),
    key: 'lastUsedAt',
    width: 220,
    render: (r) => `${fmtTs(r.lastUsedAt)} ${r.lastUsedIp ? '(' + r.lastUsedIp + ')' : ''}`,
  },
  {
    title: t('common.actions'),
    key: 'actions',
    width: 240,
    render: (r) =>
      h(NSpace, { size: 6 }, () => [
        h(
          NButton,
          { size: 'small', onClick: () => openEdit(r) },
          () => t('common.edit'),
        ),
        h(
          NPopconfirm,
          {
            onPositiveClick: () => rotate(r),
          },
          {
            default: () => t('token.rotateConfirm'),
            trigger: () =>
              h(NButton, { size: 'small', type: 'warning' }, () => t('token.rotate')),
          },
        ),
        h(
          NPopconfirm,
          {
            onPositiveClick: () => remove(r),
          },
          {
            default: () => t('token.deleteConfirm'),
            trigger: () =>
              h(NButton, { size: 'small', type: 'error' }, () => t('common.delete')),
          },
        ),
      ]),
  },
])
</script>

<template>
  <div class="space-y-4">
    <NAlert v-if="!auth.isAdmin" type="warning">
      {{ t('token.adminOnly') }}
    </NAlert>

    <NCard v-else :title="t('token.list')" :loading="loading">
      <template #header-extra>
        <NSpace>
          <NButton @click="load">{{ t('common.refresh') }}</NButton>
          <NButton type="primary" @click="openCreate">
            {{ t('token.create') }}
          </NButton>
        </NSpace>
      </template>

      <NAlert type="info" :show-icon="false" class="mb-3">
        {{ t('token.usageHint') }}
      </NAlert>

      <NDataTable
        :columns="columns"
        :data="rows"
        :bordered="false"
        :row-key="(r: ApiToken) => r.id"
        size="small"
      />
    </NCard>

    <!-- Create / Edit modal -->
    <NModal
      v-model:show="showForm"
      preset="card"
      style="width: 640px"
      :title="form.id == null ? t('token.create') : t('common.edit')"
      :mask-closable="false"
    >
      <NForm label-placement="top" :disabled="submitting">
        <NFormItem :label="t('token.remark')">
          <NInput v-model:value="form.remark" maxlength="120" clearable />
        </NFormItem>
        <NFormItem :label="t('token.pathPrefix')">
          <NInput
            v-model:value="form.allowedPathPrefix"
            placeholder="/api/v1/clients"
            clearable
          />
          <template #feedback>{{ t('token.pathPrefixTip') }}</template>
        </NFormItem>
        <NFormItem :label="t('token.methods')">
          <NInput
            v-model:value="form.allowedMethods"
            placeholder="GET,POST"
            clearable
          />
          <template #feedback>{{ t('token.methodsTip') }}</template>
        </NFormItem>
        <NFormItem :label="t('token.allowIps')">
          <NInput
            v-model:value="form.allowIpsText"
            type="textarea"
            :autosize="{ minRows: 3, maxRows: 8 }"
          />
          <template #feedback>{{ t('token.allowIpsTip') }}</template>
        </NFormItem>
        <NFormItem :label="t('token.expiresAt')">
          <NDatePicker
            v-model:value="form.expiresAt"
            type="datetime"
            clearable
            style="width: 100%"
          />
          <template #feedback>{{ t('token.neverExpire') }} = {{ t('common.empty') }}</template>
        </NFormItem>
        <NFormItem :label="t('token.disabled')">
          <NSwitch v-model:value="form.disabled" />
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showForm = false" :disabled="submitting">
            {{ t('common.cancel') }}
          </NButton>
          <NButton type="primary" :loading="submitting" @click="submit">
            {{ t('common.save') }}
          </NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- Secret reveal modal -->
    <NModal
      v-model:show="revealOpen"
      preset="card"
      style="width: 560px"
      :title="t('token.revealTitle')"
      :mask-closable="false"
    >
      <NAlert type="warning" :show-icon="true" class="mb-3">
        {{ t('token.revealTip') }}
      </NAlert>
      <NFormItem :label="t('token.keyId')" label-placement="top">
        <NInputGroup>
          <NInput :value="revealKeyId" readonly />
          <NButton @click="copyText(revealKeyId)">{{ t('common.copy') }}</NButton>
        </NInputGroup>
      </NFormItem>
      <NFormItem label="Secret" label-placement="top">
        <NInputGroup>
          <NInput :value="revealSecret" readonly />
          <NButton type="primary" @click="copyText(revealSecret)">
            {{ t('common.copy') }}
          </NButton>
        </NInputGroup>
      </NFormItem>
      <NText depth="3" style="font-size: 12px">
        Authorization: Bearer {{ revealKeyId }}.{{ revealSecret }}
      </NText>
      <template #footer>
        <NSpace justify="end">
          <NButton type="primary" @click="revealOpen = false">
            {{ t('common.confirm') }}
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>
