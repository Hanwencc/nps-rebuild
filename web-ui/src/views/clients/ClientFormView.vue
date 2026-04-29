<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  NButton,
  NCard,
  NForm,
  NFormItem,
  NGrid,
  NGridItem,
  NInput,
  NInputNumber,
  NSpace,
  NSwitch,
  useMessage,
} from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { clientApi } from '@/api/client'
import type { ClientPayload } from '@/api/types'
import { useAuthStore } from '@/stores/auth'
import { ApiError } from '@/api/request'

const props = defineProps<{ id?: string | number }>()

const { t } = useI18n()
const router = useRouter()
const message = useMessage()
const auth = useAuthStore()

const editing = computed(() => props.id !== undefined && props.id !== '')
const loading = ref(false)
const saving = ref(false)

const form = reactive<Required<ClientPayload>>({
  vkey: '',
  remark: '',
  u: '',
  p: '',
  compress: false,
  crypt: false,
  configConnAllow: true,
  rateLimit: 0,
  maxConn: 0,
  maxTunnel: 0,
  flowLimit: 0,
  webUsername: '',
  webPassword: '',
  blackIpList: [],
  ipWhite: false,
  ipWhitePass: '',
  ipWhiteList: [],
})

const blackText = ref('')
const whiteText = ref('')

async function load() {
  if (!editing.value) return
  loading.value = true
  try {
    const c = await clientApi.get(Number(props.id))
    form.vkey = c.VerifyKey
    form.remark = c.Remark
    form.u = c.Cnf?.U ?? ''
    form.p = c.Cnf?.P ?? ''
    form.compress = !!c.Cnf?.Compress
    form.crypt = !!c.Cnf?.Crypt
    form.configConnAllow = !!c.ConfigConnAllow
    form.rateLimit = c.RateLimit ?? 0
    form.maxConn = c.MaxConn ?? 0
    form.maxTunnel = c.MaxTunnelNum ?? 0
    form.flowLimit = c.Flow?.FlowLimit ?? 0
    form.webUsername = c.WebUserName ?? ''
    form.webPassword = c.WebPassword ?? ''
    form.blackIpList = c.BlackIpList ?? []
    form.ipWhite = !!c.IpWhite
    form.ipWhitePass = c.IpWhitePass ?? ''
    form.ipWhiteList = c.IpWhiteList ?? []
    blackText.value = form.blackIpList.join('\n')
    whiteText.value = form.ipWhiteList.join('\n')
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : String(e))
  } finally {
    loading.value = false
  }
}

async function submit() {
  saving.value = true
  form.blackIpList = blackText.value.split(/\r?\n/).map((s) => s.trim()).filter(Boolean)
  form.ipWhiteList = whiteText.value.split(/\r?\n/).map((s) => s.trim()).filter(Boolean)
  try {
    if (editing.value) {
      await clientApi.update(Number(props.id), { ...form })
    } else {
      await clientApi.create({ ...form })
    }
    message.success('OK')
    router.push({ name: 'clients' })
  } catch (e) {
    message.error(e instanceof ApiError ? e.message : String(e))
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <NCard :title="editing ? t('client.editTitle') : t('client.addTitle')">
    <template #header-extra>
      <NButton @click="router.back()">{{ t('common.back') }}</NButton>
    </template>

    <NForm label-placement="top" :disabled="loading">
      <NGrid :cols="2" :x-gap="16" :y-gap="0" responsive="screen">
        <NGridItem>
          <NFormItem :label="t('client.remark')">
            <NInput v-model:value="form.remark" />
          </NFormItem>
        </NGridItem>
        <NGridItem>
          <NFormItem :label="t('client.vkey')" v-if="auth.isAdmin">
            <NInput v-model:value="form.vkey" placeholder="留空则自动生成" />
          </NFormItem>
        </NGridItem>

        <NGridItem>
          <NFormItem :label="t('client.basicUser')">
            <NInput v-model:value="form.u" />
          </NFormItem>
        </NGridItem>
        <NGridItem>
          <NFormItem :label="t('client.basicPass')">
            <NInput v-model:value="form.p" type="password" show-password-on="mousedown" />
          </NFormItem>
        </NGridItem>

        <NGridItem>
          <NFormItem :label="t('client.webUser')">
            <NInput v-model:value="form.webUsername" />
          </NFormItem>
        </NGridItem>
        <NGridItem>
          <NFormItem :label="t('client.webPass')">
            <NInput v-model:value="form.webPassword" type="password" show-password-on="mousedown" />
          </NFormItem>
        </NGridItem>

        <NGridItem v-if="auth.isAdmin">
          <NFormItem :label="t('client.flowLimit')">
            <NInputNumber v-model:value="form.flowLimit" :min="0" class="!w-full" />
          </NFormItem>
        </NGridItem>
        <NGridItem v-if="auth.isAdmin">
          <NFormItem :label="t('client.rateLimit')">
            <NInputNumber v-model:value="form.rateLimit" :min="0" class="!w-full" />
          </NFormItem>
        </NGridItem>
        <NGridItem v-if="auth.isAdmin">
          <NFormItem :label="t('client.maxConn')">
            <NInputNumber v-model:value="form.maxConn" :min="0" class="!w-full" />
          </NFormItem>
        </NGridItem>
        <NGridItem v-if="auth.isAdmin">
          <NFormItem :label="t('client.maxTunnel')">
            <NInputNumber v-model:value="form.maxTunnel" :min="0" class="!w-full" />
          </NFormItem>
        </NGridItem>

        <NGridItem>
          <NFormItem :label="t('client.compress')">
            <NSwitch v-model:value="form.compress" />
          </NFormItem>
        </NGridItem>
        <NGridItem>
          <NFormItem :label="t('client.crypt')">
            <NSwitch v-model:value="form.crypt" />
          </NFormItem>
        </NGridItem>
        <NGridItem>
          <NFormItem :label="t('client.configConnAllow')">
            <NSwitch v-model:value="form.configConnAllow" />
          </NFormItem>
        </NGridItem>
        <NGridItem>
          <NFormItem :label="t('client.ipWhite')">
            <NSwitch v-model:value="form.ipWhite" />
          </NFormItem>
        </NGridItem>
        <NGridItem>
          <NFormItem :label="t('client.ipWhitePass')">
            <NInput v-model:value="form.ipWhitePass" />
          </NFormItem>
        </NGridItem>

        <NGridItem :span="2">
          <NFormItem :label="t('client.blackIpList')">
            <NInput v-model:value="blackText" type="textarea" :rows="4" placeholder="每行一个 IP" />
          </NFormItem>
        </NGridItem>
        <NGridItem :span="2">
          <NFormItem :label="t('client.ipWhiteList')">
            <NInput v-model:value="whiteText" type="textarea" :rows="4" placeholder="每行一个 IP" />
          </NFormItem>
        </NGridItem>
      </NGrid>

      <NSpace justify="end" class="mt-4">
        <NButton @click="router.back()">{{ t('common.cancel') }}</NButton>
        <NButton type="primary" :loading="saving" @click="submit">
          {{ t('common.save') }}
        </NButton>
      </NSpace>
    </NForm>
  </NCard>
</template>
